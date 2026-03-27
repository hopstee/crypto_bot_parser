package tgbot

import (
	"context"
	botauth "crypto_bot_parser/internal/auth/bot"
	"crypto_bot_parser/internal/session"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type UserState string

const (
	StateNew            UserState = "new"
	StateWaitSecret     UserState = "wait_secret"
	StateWaitPhone      UserState = "wait_phone"
	StateAuthInProgress UserState = "auth_in_progress"
	StateAuthenticated  UserState = "authenticated"
)

type UserData struct {
	ID        int64     `json:"id"`
	State     UserState `json:"state"`
	Phone     string    `json:"phone"`
	UserAgent string    `json:"user_agent"`

	MinAmount     int64 `json:"min_amount"`
	MaxAmount     int64 `json:"max_amount"`
	WsConnections int64 `json:"ws_connections"`

	CryptoBotSettings     *botauth.UserSettings      `json:"crypto_bot_settings"`
	CryptoBotBankAccounts []*botauth.UserBankAccount `json:"crypto_bot_bank_accounts"`
}

type User struct {
	ID int64

	Data *UserData

	wsCtx     context.Context
	wsCancel  context.CancelFunc
	wsRunning atomic.Bool

	client    *http.Client
	session   *session.Session
	userAgent string

	keepAliveCtx    context.Context
	keepAliveCancel context.CancelFunc

	HasActiveRequest atomic.Bool
}

type UserStore struct {
	users map[int64]*User
	dir   string

	mu sync.Mutex
}

func NewUserStore(dataDir string) *UserStore {
	dir := filepath.Join(dataDir, "users")
	_ = os.MkdirAll(dir, 0755)

	return &UserStore{
		users: make(map[int64]*User),
		dir:   dir,
	}
}

func (s *UserStore) GetOrCreate(id int64) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.users[id]; ok {
		return u
	}

	path := s.path(id)
	u := &User{
		ID: id,
		Data: &UserData{
			ID:    id,
			State: StateNew,
		},
	}

	if b, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(b, u.Data)
	}

	s.users[id] = u
	return u
}

func (s *UserStore) Save(u *UserData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	b, err := json.MarshalIndent(u, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path(u.ID), b, 0600)
}

func (s *UserStore) SetMin(id int64, min int64) {
	u := s.GetOrCreate(id)
	u.Data.MinAmount = min
	s.Save(u.Data)
}

func (s *UserStore) SetMax(id int64, max int64) {
	u := s.GetOrCreate(id)
	u.Data.MaxAmount = max
	s.Save(u.Data)
}

func (s *UserStore) SetState(id int64, state UserState) {
	u := s.GetOrCreate(id)
	u.Data.State = state
	s.Save(u.Data)
}

func (s *UserStore) SetCryptoBotSettings(id int64, settins *botauth.UserSettings) {
	u := s.GetOrCreate(id)
	u.Data.CryptoBotSettings = settins
	s.Save(u.Data)
}

func (s *UserStore) SetCryptoBotBankAccounts(id int64, accounts []*botauth.UserBankAccount) {
	u := s.GetOrCreate(id)
	u.Data.CryptoBotBankAccounts = accounts
	s.Save(u.Data)
}

func (s *UserStore) SetWsConnections(id int64, count int64) {
	u := s.GetOrCreate(id)
	u.Data.WsConnections = count
	s.Save(u.Data)
}

func (s *UserStore) path(id int64) string {
	return filepath.Join(s.dir, fmt.Sprintf("%d.json", id))
}
