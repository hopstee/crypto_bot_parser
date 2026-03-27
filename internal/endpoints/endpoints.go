package endpoints

import (
	"crypto_bot_parser/internal/constants"
	"net/url"
	"strings"
)

func GetQueryParams(botID, origin string) string {
	return strings.Join(
		[]string{
			"bot_id=",
			botID,
			"&origin=",
			origin,
			"&lang=ru",
		},
		"",
	)
}

func AuthRequestURL(p string) string {
	return strings.Join(
		[]string{
			constants.TGOauthURL,
			"/auth/request?",
			p,
		},
		"",
	)
}

func AuthLoginURL(p string) string {
	return strings.Join(
		[]string{
			constants.TGOauthURL,
			"/auth/login?",
			p,
		},
		"",
	)
}

func AuthHashURL(p string) string {
	return strings.Join(
		[]string{
			constants.TGOauthURL,
			"/auth?",
			p,
		},
		"",
	)
}

func AuthConfirmURL(p, hash string) string {
	return strings.Join(
		[]string{
			constants.TGOauthURL,
			"/auth/auth?",
			p,
			"&confirm=1&hash=",
			hash,
		},
		"",
	)
}

func AuthPushURL(p string) string {
	return strings.Join(
		[]string{
			constants.TGOauthURL,
			"/auth/push?",
			p,
		},
		"",
	)
}

func CryptoBotTgMakeAuthURL() string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIBaseURL, "/authentication/telegram")
	return e
}

func CryptoBotTgCheckAuthURL() string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIBaseURL, "/user/settings")
	return e
}

func CryptoBotBankAccountsURL() string {
	return constants.CryptBotInternalAPIP2cBankAccountsURL
}

func CryptoBotTakeOrderURL(id string) string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIP2cPaymentsURL, "/take", id)
	return e
}

func CryptoBotCheckOrderURL(id string) string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIP2cPaymentsURL, id)
	return e
}

func CryptoBotCancelOrderURL(id string) string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIP2cPaymentsURL, id, "/cancel")
	return e
}

func CryptoBotCompleteOrderURL(id string) string {
	e, _ := url.JoinPath(constants.CryptBotInternalAPIP2cPaymentsURL, id, "/complete")
	return e
}

func CryptoBotVersion() string {
	return constants.CryptBotVersion
}
