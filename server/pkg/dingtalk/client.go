package dingtalk

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseSize = 64 << 10

var httpClient = &http.Client{
	Timeout: 5 * time.Second,
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type Message struct {
	Title string
	Text  string
}

type markdownPayload struct {
	MsgType  string   `json:"msgtype"`
	Markdown markdown `json:"markdown"`
}

type markdown struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

type response struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func ValidateWebhook(rawURL string) error {
	webhook, err := url.ParseRequestURI(strings.TrimSpace(rawURL))
	if err != nil {
		return errors.New("钉钉 Webhook 地址格式错误")
	}
	if webhook.Scheme != "https" || !strings.EqualFold(webhook.Hostname(), "oapi.dingtalk.com") {
		return errors.New("钉钉 Webhook 必须使用 https://oapi.dingtalk.com")
	}
	if webhook.User != nil || webhook.Port() != "" || webhook.Path != "/robot/send" || webhook.Fragment != "" {
		return errors.New("钉钉 Webhook 地址格式错误")
	}
	if strings.TrimSpace(webhook.Query().Get("access_token")) == "" {
		return errors.New("钉钉 Webhook 缺少 access_token")
	}
	return nil
}

func Send(rawURL, secret string, message Message) error {
	if err := ValidateWebhook(rawURL); err != nil {
		return err
	}
	webhook, err := signedWebhook(rawURL, secret, time.Now())
	if err != nil {
		return err
	}
	payload, err := json.Marshal(markdownPayload{
		MsgType: "markdown",
		Markdown: markdown{
			Title: message.Title,
			Text:  message.Text,
		},
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, webhook, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return describeNetworkError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, maxResponseSize))
		return fmt.Errorf("钉钉 Webhook 返回 HTTP %d", resp.StatusCode)
	}
	var result response
	if err = json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&result); err != nil {
		return fmt.Errorf("解析钉钉 Webhook 响应失败: %w", err)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("钉钉 Webhook 返回错误: %s (%d)", result.ErrMsg, result.ErrCode)
	}
	return nil
}

func describeNetworkError(err error) error {
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return errors.New("发送钉钉通知失败：服务器缺少 HTTPS CA 根证书，请更新后端镜像")
	}
	var dnsError *net.DNSError
	if errors.As(err, &dnsError) {
		return errors.New("发送钉钉通知失败：服务器无法解析 oapi.dingtalk.com")
	}
	var networkError net.Error
	if errors.As(err, &networkError) && networkError.Timeout() {
		return errors.New("发送钉钉通知失败：连接 oapi.dingtalk.com 超时")
	}
	return errors.New("发送钉钉通知失败：服务器无法连接 oapi.dingtalk.com:443")
}

func signedWebhook(rawURL, secret string, now time.Time) (string, error) {
	webhook, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return webhook.String(), nil
	}
	timestamp := now.UnixMilli()
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stringToSign))
	sign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	query := webhook.Query()
	query.Set("timestamp", strconv.FormatInt(timestamp, 10))
	query.Set("sign", sign)
	webhook.RawQuery = query.Encode()
	return webhook.String(), nil
}
