package botauth

type UserSettings struct {
	ID               int             `json:"id"`
	FirstName        string          `json:"first_name"`
	IsTester         bool            `json:"is_tester"`
	LocalCurrency    string          `json:"local_currency"`
	IsCountryAllowed bool            `json:"is_country_allowed"`
	IsSubscribed     bool            `json:"is_subscribed"`
	MarketName       string          `json:"market_name"`
	TelegramUsername string          `json:"telegram_username"`
	IsP2cMerchant    bool            `json:"is_p2c_merchant"`
	Language         string          `json:"language"`
	Timezone         string          `json:"timezone"`
	IsNewUser        bool            `json:"is_new_user"`
	Flags            map[string]bool `json:"flags"`
	Permissions      map[string]any  `json:"permissions"`
	Compliance       struct {
		Lock                   bool `json:"lock"`
		PendingActionsFromUser bool `json:"pending_actions_from_user"`
		DeclinedDepositsCount  int  `json:"declined_deposits_count"`
	} `json:"compliance"`
	P2cLimitsRub struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"p2c_limits_rub"`
}

type UserBankAccount struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Bank          string `json:"bank"`
	AccountNumber string `json:"account_number"`
	Status        string `json:"status"`
}
