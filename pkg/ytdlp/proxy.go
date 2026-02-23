package ytdlp

import (
	"os"
	"strings"
	"time"
)

func NewProxyList(prefix string) *ProxyList {
	proxyList := ProxyList{}
	for _, env := range os.Environ() {
		if idx := strings.Index(env, "="); idx > 0 {
			if strings.HasPrefix(env[:idx], prefix) {
				proxyList = append(proxyList, NewProxy(env[idx+1:]))
			}
		}
	}
	return &proxyList
}

type ProxyList []*Proxy

func (pl *ProxyList) Next() *Proxy {
	for _, proxy := range *pl {
		if !proxy.isBanned() {
			return proxy
		}
	}
	return nil
}

func (pl *ProxyList) MarkBanned(proxy *Proxy) {
	for _, p := range *pl {
		if p == proxy {
			p.banned = time.Now()
		}
	}
}

func NewProxy(url string) *Proxy {
	return &Proxy{
		url:    url,
		banned: time.Now().Add(-1 * time.Hour),
	}
}

type Proxy struct {
	url    string
	banned time.Time
}

func (p *Proxy) isBanned() bool {
	return time.Since(p.banned) < time.Hour
}
