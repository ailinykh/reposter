package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/ailinykh/reposter/v2/internal/core"
)

func CreateJsonCookieJar(l core.Logger, cookieFile string) http.CookieJar {
	return &JsonCookieJar{l, cookieFile, []*http.Cookie{}}
}

type JsonCookieJar struct {
	l        core.Logger
	filename string
	cookies  []*http.Cookie
}

func (jar *JsonCookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = jar.merge(jar.cookies, cookies)
	data, err := json.MarshalIndent(jar.cookies, "", "  ")
	if err != nil {
		jar.l.Error(err)
		return
	}

	err = os.WriteFile(jar.filename, data, 0644)
	if err != nil {
		jar.l.Error(err)
		return
	}
}

func (jar *JsonCookieJar) Cookies(u *url.URL) []*http.Cookie {
	data, err := os.ReadFile(jar.filename)
	if err != nil {
		jar.l.Error(err)
		jar.cookies = []*http.Cookie{}
		return jar.cookies
	}

	err = json.Unmarshal(data, &jar.cookies)
	if err != nil {
		jar.l.Error(err)
		jar.cookies = []*http.Cookie{}
	}

	return jar.cookies
}

func (j *JsonCookieJar) merge(lhs []*http.Cookie, rhs []*http.Cookie) []*http.Cookie {
	for _, r := range rhs {
		found := false
		for _, l := range lhs {
			if l.Name == r.Name {
				found = true
				l = r
				break
			}
		}
		if !found {
			lhs = append(lhs, r)
		}
	}
	return lhs
}
