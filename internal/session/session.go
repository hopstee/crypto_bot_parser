package session

import (
	tgauth "crypto_bot_parser/internal/auth/tg"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
)

type Session struct {
	mu sync.Mutex

	Cookies    map[string][]*http.Cookie `json:"cookies"`
	AuthResult *tgauth.TgAuthResult      `json:"auth_result"`

	path string
}

func NewSession(path string) *Session {
	_ = os.MkdirAll(filepath.Dir(path), 0755)

	s := &Session{
		path: path,
	}
	_ = s.load()

	if s.Cookies == nil {
		s.Cookies = make(map[string][]*http.Cookie)
	}

	return s
}

func (s *Session) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0600)
}

func (s *Session) SetAuthResult(authResult *tgauth.TgAuthResult) {
	s.mu.Lock()
	s.AuthResult = authResult
	s.mu.Unlock()
}

func (s *Session) load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s)
}

func (s *Session) SyncFromClient(client *http.Client, domains ...string) {
	if client == nil || client.Jar == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, raw := range domains {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		s.Cookies[raw] = client.Jar.Cookies(u)
	}
}

func (s *Session) ApplyToClient(client *http.Client) {
	if client == nil || client.Jar == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for raw, cookies := range s.Cookies {
		u, err := url.Parse(raw)
		if err != nil {
			continue
		}
		client.Jar.SetCookies(u, cookies)
	}
}

func (s *Session) Lock() {
	s.mu.Lock()
}

func (s *Session) Unlock() {
	s.mu.Unlock()
}
