package services

import (
	"net"
	"net/url"

	"bbs-go/pkg/config"
)

func AllowedPresenceOrigins() []string {
	if len(config.Instance.Presence.AllowedOrigins) > 0 {
		return config.Instance.Presence.AllowedOrigins
	}
	if config.Instance.BaseUrl != "" {
		if parsed, err := url.Parse(config.Instance.BaseUrl); err == nil && parsed.Scheme != "" && parsed.Host != "" {
			return []string{parsed.Scheme + "://" + parsed.Host}
		}
	}
	return nil
}

func IsTrustedProxy(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	for _, cidr := range config.Instance.Presence.TrustedProxies {
		_, network, err := net.ParseCIDR(cidr)
		if err == nil && network.Contains(parsedIP) {
			return true
		}
	}
	return parsedIP.IsLoopback()
}
