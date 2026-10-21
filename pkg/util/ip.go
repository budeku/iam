package util

import (
	"net"
	"net/http"
	"slices"
	"strings"
	"sync"
)

var (
	trustedMu      sync.RWMutex
	trustedProxies []*net.IPNet = make([]*net.IPNet, 0)
)

func SetTrustedProxies(cidrs []string) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		// 如果没有指定掩码，则自动补充单 IP 掩码（IPv4 /32，IPv6 /128）
		if !strings.Contains(c, "/") {
			if ip := net.ParseIP(c); ip != nil {
				if ip.To4() != nil {
					c += "/32"
				} else {
					c += "/128"
				}
			}
		}

		_, ipNet, err := net.ParseCIDR(c)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}

	trustedMu.Lock()
	trustedProxies = nets
	trustedMu.Unlock()
}

func isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}
	trustedMu.RLock()
	defer trustedMu.RUnlock()

	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func ClientIP(r *http.Request) string {
	remoteIP := remoteAddrIP(r.RemoteAddr)

	// 如果直接请求的远端不是受信任代理，则直接返回其 IP 或原始 RemoteAddr
	if remoteIP == nil || !isTrustedProxy(remoteIP) {
		if remoteIP != nil {
			return remoteIP.String()
		}
		return r.RemoteAddr
	}

	// 从 X-Forwarded-For 中从右向左寻找第一个非受信任的客户端 IP
	if ip := clientIPFromXFF(r.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}

	// 尝试解析 X-Real-IP
	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		if ip := net.ParseIP(xrip); ip != nil {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func clientIPFromXFF(xff string) string {
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")

	// 从右向左遍历，最右侧是最近一层代理，最左侧通常是客户端原始 IP
	for _, part := range slices.Backward(parts) {
		candidate := strings.TrimSpace(part)
		ip := net.ParseIP(candidate)
		if ip == nil {
			continue
		}

		// 返回第一个不是受信任代理的 IP，即为真实客户端
		if !isTrustedProxy(ip) {
			return ip.String()
		}
	}
	return ""
}

func remoteAddrIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(host)
}

func UserAgent(r *http.Request) string {
	return strings.TrimSpace(r.UserAgent())
}
