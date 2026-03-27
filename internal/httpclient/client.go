package httpclient

import (
	"crypto_bot_parser/internal/session"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

var sharedTransport = &http.Transport{
	MaxIdleConns:        1000,
	MaxIdleConnsPerHost: 1000,
	MaxConnsPerHost:     1000,

	IdleConnTimeout:       15 * time.Minute,
	TLSHandshakeTimeout:   5 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,

	DisableCompression: true,
	ForceAttemptHTTP2:  true,
	DisableKeepAlives:  false,
}

func NewHttpClient(s *session.Session) *http.Client {
	jar, _ := cookiejar.New(nil)

	client := &http.Client{
		Jar:       jar,
		Timeout:   10 * time.Second,
		Transport: sharedTransport,
	}

	if s != nil {
		for rawURL, cookies := range s.Cookies {
			u, err := url.Parse(rawURL)
			if err != nil {
				continue
			}
			jar.SetCookies(u, cookies)
		}
	}

	return client
}
