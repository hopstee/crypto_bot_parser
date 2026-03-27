package botauth

import (
	"bytes"
	"context"
	tgauth "crypto_bot_parser/internal/auth/tg"
	"crypto_bot_parser/internal/constants"
	"crypto_bot_parser/internal/endpoints"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type BotAuth struct {
	client *http.Client

	userAgent string

	userData UserSettings
}

type AccessTokenResponseBody struct {
	Settings UserSettings `json:"settings"`
}

func NewBotAuth(
	c *http.Client,
	ua string,
) *BotAuth {
	return &BotAuth{
		client:    c,
		userAgent: ua,
	}
}

func (a *BotAuth) CheckAccessTokenAndSettings(ctx context.Context) (*UserSettings, bool) {
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoints.CryptoBotTgCheckAuthURL(),
		nil,
	)

	req.Header.Set("User-Agent", a.userAgent)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var userSettings UserSettings
		if err := json.NewDecoder(resp.Body).Decode(&userSettings); err != nil {
			return nil, false
		}

		return &userSettings, true
	}

	return nil, false
}

func (a *BotAuth) ObtainAccessToken(
	ctx context.Context,
	authResult *tgauth.TgAuthResult,
) (*UserSettings, error) {
	data, err := json.Marshal(&authResult)
	if err != nil {
		slog.Error("failed marshar auth result", slog.Any("error", err.Error()))
		return nil, fmt.Errorf("failed marshar auth result: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		endpoints.CryptoBotTgMakeAuthURL(),
		bytes.NewReader(data),
	)
	if err != nil {
		slog.Error("failed prepare crypto bot auth", slog.Any("error", err.Error()))
		return nil, fmt.Errorf("failed prepare crypto bot auth: %w", err)
	}

	req.Header.Set("User-Agent", a.userAgent)
	req.Header.Set("Content-Type", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		slog.Error("failed crypto bot auth", slog.Any("error", err.Error()))
		return nil, fmt.Errorf("failed crypto bot auth: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("obtain access token bad status", slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("obtain access token bad status: %s", resp.Status)
	}

	var userSettings AccessTokenResponseBody
	if err := json.NewDecoder(resp.Body).Decode(&userSettings); err != nil {
		return nil, fmt.Errorf("failed decode obtain access token: %s", resp.Status)
	}

	cookies := resp.Cookies()
	for _, c := range cookies {

		c.Path = "/"
	}
	u, _ := url.Parse(constants.CryptBotBaseURL)
	a.client.Jar.SetCookies(u, cookies)

	return &userSettings.Settings, nil
}
