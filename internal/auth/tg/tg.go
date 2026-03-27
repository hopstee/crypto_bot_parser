package tgauth

import (
	"context"
	"crypto_bot_parser/internal/auth/listener"
	"crypto_bot_parser/internal/endpoints"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type TgAuth struct {
	phone  string
	botID  string
	client *http.Client

	queryParams string
	hashValue   string
	authResult  *TgAuthResult

	userAgent string

	listener listener.AuthListener
}

func NewTgAuth(
	phone,
	botID string,
	c *http.Client,
	qp string,
	ua string,
	l listener.AuthListener,
) *TgAuth {
	return &TgAuth{
		phone:       phone,
		botID:       botID,
		client:      c,
		queryParams: qp,
		userAgent:   ua,
		listener:    l,
	}
}

func (a *TgAuth) Run(ctx context.Context) (*TgAuthResult, error) {
	slog.Info("tg auth started")

	var err error
	err = a.checkAuth(ctx)

	if err != nil {
		if err := a.requestLogin(ctx); err != nil {
			a.listener.OnError(errors.New("Не получилось запросить аутентификацию в телеграме"))
			return nil, err
		}

		a.listener.OnWaitingConfirmation()
		if err := a.pollLogin(ctx); err != nil {
			a.listener.OnError(errors.New("Не получилось запросить аутентификацию в телеграме"))
			return nil, err
		}

		err = a.checkAuth(ctx)
	}

	if err != nil {
		a.listener.OnError(errors.New("Не получилось запросить аутентификацию в телеграме"))
		return nil, errors.New("failed telegram authentication")
	}

	slog.Info("tg auth completed")
	return a.authResult, nil
}

func (a *TgAuth) checkAuth(ctx context.Context) error {
	if err := a.obtainHashOrResult(ctx); err != nil {
		return err
	}

	if a.hashValue != "" {
		if err := a.confirmAuth(ctx); err != nil {
			return err
		}
		if err := a.pushAuth(ctx); err != nil {
			return err
		}
	}

	if a.authResult == nil {
		return errors.New("auth result empty")
	}

	return nil
}

func (a *TgAuth) requestLogin(ctx context.Context) error {
	slog.Info("request login")

	data := url.Values{}
	data.Set("phone", a.phone)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoints.AuthRequestURL(a.queryParams),
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	req.Header.Set("User-Agent", a.userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		slog.Error("login request failed", slog.Any("error", err.Error()))
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Warn(
			"unexpected login status",
			slog.Int("status", resp.StatusCode),
		)
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	slog.Info("login request accepted")
	return nil
}

func (a *TgAuth) pollLogin(ctx context.Context) error {
	slog.Info("polling login status")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	timeout := time.After(1 * time.Minute)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("login polling timed out")
		case <-ticker.C:
			ok, err := a.checkLoginOnce(ctx)
			if err != nil {
				slog.Warn("login check failed", slog.Any("error", err.Error()))
				continue
			}
			if ok {
				slog.Info("login confirmed")
				return nil
			}
		}
	}
}

func (a *TgAuth) checkLoginOnce(ctx context.Context) (bool, error) {
	slog.Info("need to check telegram and confirm request")
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoints.AuthLoginURL(a.queryParams),
		nil,
	)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(bodyBytes)) == "true", nil
}

func (a *TgAuth) obtainHashOrResult(ctx context.Context) error {
	slog.Info("obtaining hash or auth result")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoints.AuthHashURL(a.queryParams),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	body := strings.TrimSpace(string(bodyBytes))

	if hash := ObtainAccessHash(body); hash != "" {
		a.hashValue = hash
		slog.Info("access hash received")
		return nil
	}

	result, ok, err := ParseAuthResult(body)
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("no hash and no auth result found")
	}

	a.authResult = result
	slog.Info("auth result received directly")
	return nil
}

func (a *TgAuth) confirmAuth(ctx context.Context) error {
	if a.hashValue == "" {
		return errors.New("empty hash value")
	}

	slog.Info("confirming auth")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoints.AuthConfirmURL(a.queryParams, a.hashValue),
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", a.userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusFound {
		return fmt.Errorf("confirm auth failed: %s", resp.Status)
	}

	return nil
}

func (a *TgAuth) pushAuth(ctx context.Context) error {
	slog.Info("pushing auth")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoints.AuthPushURL(a.queryParams),
		nil,
	)
	if err != nil {
		return err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	result, ok, err := ParseAuthResult(string(body))
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("push auth: no auth result")
	}

	a.authResult = result
	return nil
}
