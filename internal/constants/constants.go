package constants

const (
	TGOauthURL = "https://oauth.telegram.org"

	CryptBotBaseURL                       = "https://app.send.tg"
	CryptBotVersion                       = CryptBotBaseURL + "/version.json"
	CryptBotInternalAPIBaseURL            = CryptBotBaseURL + "/internal/v1"
	CryptBotInternalAPIP2cURL             = CryptBotInternalAPIBaseURL + "/p2c"
	CryptBotInternalAPIP2cPaymentsURL     = CryptBotInternalAPIP2cURL + "/payments"
	CryptBotInternalAPIP2cBankAccountsURL = CryptBotInternalAPIP2cURL + "/accounts"

	CryptBotWsBaseURL = "wss://app.send.tg"
	CryptBotWsP2cURL  = CryptBotWsBaseURL + "/internal/v1/p2c-socket/?EIO=4&transport=websocket"

	TGOrigin       = "https://crypt.bot"
	CryptBotOrigin = "https://app.send.tg"
)
