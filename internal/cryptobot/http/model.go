package http

type TakeDealData struct {
	ID           int64  `json:"id"`
	Payload      string `json:"payload"`
	URL          string `json:"url"`
	BrandName    string `json:"brand_name"`
	InAsset      string `json:"in_asset"`
	OutAsset     string `json:"out_asset"`
	Provider     string `json:"provider"`
	InAmount     string `json:"in_amount"`
	OutAmount    string `json:"out_amount"`
	FeeAmount    string `json:"fee_amount"`
	ExchangeRate string `json:"exchange_rate"`
	ExpiresAt    string `json:"expires_at"`
}

type CheckDealData struct {
	ID           string `json:"id"`
	Payload      string `json:"payload"`
	URL          string `json:"url"`
	BrandName    string `json:"brand_name"`
	InAsset      string `json:"in_asset"`
	OutAsset     string `json:"out_asset"`
	Provider     string `json:"provider"`
	InAmount     string `json:"in_amount"`
	OutAmount    string `json:"out_amount"`
	FeeAmount    string `json:"fee_amount"`
	ExchangeRate string `json:"exchange_rate"`
	ExpiresAt    string `json:"expires_at"`
	Status       string `json:"status"`
}
