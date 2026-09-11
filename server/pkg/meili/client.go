package meili

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client 是 Meilisearch 的最小 HTTP 客户端实现。
// 只覆盖本项目用到的接口（建索引、改配置、写文档、删文档、查询），避免引入额外依赖。
type Client struct {
	baseUrl string
	apiKey  string
	http    *http.Client
}

// NewClient 创建客户端，timeout 为单次请求超时时间。
func NewClient(baseUrl, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &Client{
		baseUrl: strings.TrimRight(baseUrl, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: timeout},
	}
}

// Document 是写入搜索引擎的主题文档。
// 注意：搜索引擎只作为“倒排索引 + 过滤 + 排序”使用，
// 真正的业务数据仍然以 MySQL 为准，查询后会用主键回表。
type Document struct {
	ID              string   `json:"id"`
	TopicId         int64    `json:"topicId"`
	IsOldBBS        bool     `json:"isOldBBS"`
	Title           string   `json:"title"`
	Content         string   `json:"content"`
	AuthorName      string   `json:"authorName"`
	Tags            []string `json:"tags"`
	NodeId          int64    `json:"nodeId"`
	NodeName        string   `json:"nodeName"`
	UserId          int64    `json:"userId"`
	AccessLv        int      `json:"accessLv"`
	Status          int      `json:"status"`
	Recommend       bool     `json:"recommend"`
	CreateTime      int64    `json:"createTime"`
	LastCommentTime int64    `json:"lastCommentTime"`
	ViewCount       int64    `json:"viewCount"`
	CommentCount    int64    `json:"commentCount"`
	LikeCount       int64    `json:"likeCount"`
}

// Settings 索引配置，filterableAttributes/sortableAttributes 必须显式声明，
// 否则 filter/sort 参数会被 Meilisearch 忽略。
type Settings struct {
	SearchableAttributes []string    `json:"searchableAttributes"`
	FilterableAttributes []string    `json:"filterableAttributes"`
	SortableAttributes   []string    `json:"sortableAttributes"`
	RankingRules         []string    `json:"rankingRules"`
	Pagination           *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	MaxTotalHits int `json:"maxTotalHits"`
}

// SearchRequest 查询请求。
type SearchRequest struct {
	Q                    string   `json:"q"`
	Filter               string   `json:"filter,omitempty"`
	Limit                int      `json:"limit,omitempty"`
	Offset               int      `json:"offset,omitempty"`
	Sort                 []string `json:"sort,omitempty"`
	AttributesToRetrieve []string `json:"attributesToRetrieve,omitempty"`
}

// SearchResponse 查询响应。
// 使用 EstimatedTotalHits（近似总数）而不是精确 count，避免每次搜索都做全量统计。
type SearchResponse struct {
	Hits               []Document `json:"hits"`
	EstimatedTotalHits int64      `json:"estimatedTotalHits"`
	Offset             int        `json:"offset"`
	Limit              int        `json:"limit"`
	ProcessingTimeMs   int        `json:"processingTimeMs"`
}

type apiError struct {
	Message string `json:"message"`
	Code    string `json:"code"`
	Type    string `json:"type"`
}

func (e *apiError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("meilisearch error: %s (code=%s, type=%s)", e.Message, e.Code, e.Type)
}

type taskResponse struct {
	TaskUid int64     `json:"taskUid"`
	Uid     int64     `json:"uid"`
	Status  string    `json:"status"`
	Error   *apiError `json:"error"`
}

func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.baseUrl+path, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusMultipleChoices {
		e := &apiError{}
		if json.Unmarshal(data, e) != nil || e.Message == "" {
			return fmt.Errorf("meilisearch http %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
		}
		return e
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

// Health 健康检查。
func (c *Client) Health() error {
	var out map[string]interface{}
	return c.do(http.MethodGet, "/health", nil, &out)
}

// EnsureIndex 确保索引存在（已存在则忽略）。
func (c *Client) EnsureIndex(uid, primaryKey string) error {
	var task taskResponse
	err := c.do(http.MethodPost, "/indexes", map[string]string{
		"uid":        uid,
		"primaryKey": primaryKey,
	}, &task)
	if err != nil {
		// 索引已创建时 Meilisearch 返回 index_already_exists，属于正常情况。
		if isIndexAlreadyExists(err) {
			return nil
		}
		return err
	}
	// 注意：重复创建时 HTTP 状态仍是 202，错误只体现在任务结果里，
	// 所以这里也必须把 index_already_exists 当成成功，否则重启后会中断初始化。
	if err := c.waitTask(task.TaskUid, 15*time.Second); err != nil {
		if isIndexAlreadyExists(err) {
			return nil
		}
		return err
	}
	return nil
}

func isIndexAlreadyExists(err error) bool {
	return err != nil && strings.Contains(err.Error(), "index_already_exists")
}

// UpdateSettings 更新索引配置并等待生效。
// 必须等待，否则在配置生效前发起查询会拿到错误的过滤/排序结果。
func (c *Client) UpdateSettings(uid string, settings Settings) error {
	var task taskResponse
	if err := c.do(http.MethodPatch, "/indexes/"+uid+"/settings", settings, &task); err != nil {
		return err
	}
	return c.waitTask(task.TaskUid, 15*time.Second)
}

// AddDocuments 批量写入/更新文档（按主键幂等 upsert），并等待任务结束。
// 等待是为了让“写入被拒绝”这类错误（如文档 ID 非法）能被发现，
// 而不是静默丢失数据；单条写入通常只需几毫秒。
func (c *Client) AddDocuments(uid string, documents interface{}) error {
	var task taskResponse
	if err := c.do(http.MethodPut, "/indexes/"+uid+"/documents?primaryKey=id", documents, &task); err != nil {
		return err
	}
	return c.waitTask(task.TaskUid, 5*time.Second)
}

// DeleteDocuments 批量删除文档（不存在的文档会被忽略）。
func (c *Client) DeleteDocuments(uid string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}
	var task taskResponse
	if err := c.do(http.MethodPost, "/indexes/"+uid+"/documents/delete-batch", ids, &task); err != nil {
		return err
	}
	return c.waitTask(task.TaskUid, 5*time.Second)
}

// DeleteAllDocuments 清空索引。
func (c *Client) DeleteAllDocuments(uid string) error {
	var task taskResponse
	if err := c.do(http.MethodDelete, "/indexes/"+uid+"/documents", nil, &task); err != nil {
		return err
	}
	return c.waitTask(task.TaskUid, 15*time.Second)
}

// Search 执行查询。
func (c *Client) Search(uid string, req SearchRequest) (*SearchResponse, error) {
	resp := &SearchResponse{}
	if err := c.do(http.MethodPost, "/indexes/"+uid+"/search", req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// waitTask 等待异步任务完成。Meilisearch 的写入/配置接口都是异步的。
func (c *Client) waitTask(taskUid int64, timeout time.Duration) error {
	if taskUid == 0 {
		return nil
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	deadline := time.Now().Add(timeout)
	for {
		var task taskResponse
		if err := c.do(http.MethodGet, fmt.Sprintf("/tasks/%d", taskUid), nil, &task); err != nil {
			return err
		}
		switch task.Status {
		case "succeeded":
			return nil
		case "failed", "canceled":
			if task.Error != nil {
				return task.Error
			}
			return fmt.Errorf("meilisearch task %d status=%s", taskUid, task.Status)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("meilisearch task %d timeout", taskUid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
