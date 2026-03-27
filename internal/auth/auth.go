package auth

import (
	"context"
	botauth "crypto_bot_parser/internal/auth/bot"
	"crypto_bot_parser/internal/auth/listener"
	tgauth "crypto_bot_parser/internal/auth/tg"
	"crypto_bot_parser/internal/constants"
	"crypto_bot_parser/internal/endpoints"
	"crypto_bot_parser/internal/session"
	"log/slog"
	"net/http"
)

type Auth struct {
	client  *http.Client
	session *session.Session

	phone string
	botID string

	queryParams string
	userAgent   string

	listener listener.AuthListener

	userSettings *botauth.UserSettings
}

func NewAuth(
	c *http.Client,
	s *session.Session,
	ua string,
	p string,
	bid string,
) *Auth {
	qp := endpoints.GetQueryParams(bid, constants.TGOrigin)

	return &Auth{
		client:       c,
		session:      s,
		phone:        p,
		botID:        bid,
		queryParams:  qp,
		userAgent:    ua,
		userSettings: &botauth.UserSettings{},
	}
}

func (a *Auth) Run(ctx context.Context) error {
	a.session.Lock()
	authResult := a.session.AuthResult
	a.session.Unlock()

	ta := tgauth.NewTgAuth(
		a.phone,
		a.botID,
		a.client,
		a.queryParams,
		a.userAgent,
		a.listener,
	)

	ba := botauth.NewBotAuth(
		a.client,
		a.userAgent,
	)

	userSettings, ok := ba.CheckAccessTokenAndSettings(ctx)
	if ok {
		slog.Info("Access token is valid")
		a.userSettings = userSettings
		a.listener.OnAuthorized()
		return nil
	}

	if authResult != nil {
		slog.Info("Trying to obtain access token via authResult")
		err := a.obtainAccessToken(ctx, ba, authResult)
		if err == nil {
			a.listener.OnAuthorized()
			return nil
		}
		slog.Warn("Failed to obtain token via authResult", slog.Any("error", err.Error()))
	}

	slog.Info("Running full TG auth")
	result, err := ta.Run(ctx)
	if err != nil {
		return err
	}

	a.session.SetAuthResult(result)

	if err := a.obtainAccessToken(ctx, ba, result); err != nil {
		a.listener.OnError(err)
		slog.Error("obtain access token error", slog.Any("error", err.Error()))
		return err
	}

	a.session.SyncFromClient(a.client,
		constants.TGOauthURL,
		constants.CryptBotBaseURL,
	)

	if err := a.session.Save(); err != nil {
		slog.Error("failed to save session", slog.Any("error", err.Error()))
	}

	a.listener.OnAuthorized()
	return nil
}

func (a *Auth) SetListener(l listener.AuthListener) {
	a.listener = l
}

func (a *Auth) obtainAccessToken(ctx context.Context, ba *botauth.BotAuth, authResult *tgauth.TgAuthResult) error {
	userSettings, err := ba.ObtainAccessToken(ctx, authResult)
	if err != nil {
		slog.Error("failed obtain access token", slog.Any("error", err.Error()))
		return err
	}

	a.userSettings = userSettings
	return nil
}

func (a *Auth) GetCryptoBotSettings() *botauth.UserSettings {
	return a.userSettings
}
