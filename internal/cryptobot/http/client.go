package http

import (
	"bytes"
	"context"
	botauth "crypto_bot_parser/internal/auth/bot"
	"crypto_bot_parser/internal/endpoints"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

var (
	ErrCloudflareBadResponse = errors.New("bad response")
	ErrForbidden             = errors.New("forbidden request")
	ErrUnauthorized          = errors.New("unauthorized request")
)

type TakeDealRaceData struct {
	data *TakeDealData
	err  error
}

type TakeDealResponseData struct {
	Data TakeDealData `json:"data"`
}

type CheckBankAccountsResponseData struct {
	Data []*botauth.UserBankAccount `json:"data"`
}

func TakeDeal(
	ctx context.Context,
	eventID string,
	userAgent string,
	client *http.Client,
) (*TakeDealData, error) {
	url := endpoints.CryptoBotTakeOrderURL(eventID)

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed take request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result TakeDealResponseData
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}
		return &result.Data, nil
	}

	if resp.StatusCode == 520 {
		return nil, ErrCloudflareBadResponse
	}

	if resp.StatusCode == 403 {
		return nil, ErrForbidden
	}

	if resp.StatusCode == 401 {
		return nil, ErrUnauthorized
	}

	return nil, fmt.Errorf("take request bad status: %s. Code: %d", resp.Status, resp.StatusCode)
}

func CheckDeal(
	ctx context.Context,
	eventID string,
	userAgent string,
	client *http.Client,
) (*CheckDealData, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoints.CryptoBotCheckOrderURL(eventID), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("check request bad status", slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("check request bad status: %d", resp.StatusCode)
	}

	var result CheckDealData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed decode obtain access token: %s", resp.Status)
	}
	return &result, nil
}

func ConfirmDeal(
	ctx context.Context,
	reqID string,
	methodID string,
	userAgent string,
	client *http.Client,
) error {
	data, err := json.Marshal(map[string]string{"method": methodID})
	if err != nil {
		slog.Error("failed marshar confirm data", slog.Any("error", err.Error()))
		return fmt.Errorf("failed marshar confirm data: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoints.CryptoBotCompleteOrderURL(reqID), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("confirm request bad status", slog.Int("status", resp.StatusCode))
		return fmt.Errorf("confirm request bad status: %d", resp.StatusCode)
	}

	return nil
}

func CancelDeal(ctx context.Context, reqID, userAgent string, client *http.Client) error {
	data, err := json.Marshal(map[string]string{"reason": "bank"})
	if err != nil {
		slog.Error("failed marshar cancel data", slog.Any("error", err.Error()))
		return fmt.Errorf("failed marshar cancel data: %w", err)
	}

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoints.CryptoBotCancelOrderURL(reqID), bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("cancel request bad status", slog.Int("status", resp.StatusCode))
		return fmt.Errorf("cancel request bad status: %d", resp.StatusCode)
	}
	return nil
}

func CheckUserSettings(ctx context.Context, userAgent string, client *http.Client) (*botauth.UserSettings, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoints.CryptoBotTgCheckAuthURL(), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("check user settings request bad status", slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("check user settings request bad status: %d", resp.StatusCode)
	}

	var userSettings botauth.UserSettings
	if err := json.NewDecoder(resp.Body).Decode(&userSettings); err != nil {
		slog.Error("failed get data form check user settings request", slog.Any("error", err.Error()))
		return nil, fmt.Errorf("failed get data form check user settings request: %d", resp.StatusCode)
	}

	return &userSettings, nil
}

func CheckBankAccounts(ctx context.Context, userAgent string, client *http.Client) ([]*botauth.UserBankAccount, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoints.CryptoBotBankAccountsURL(), nil)
	req.Header.Set("Content-Type", "application/json")

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("check bank accounts request bad status", slog.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("check user settings request bad status: %d", resp.StatusCode)
	}

	var result CheckBankAccountsResponseData
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		slog.Error("failed get data form check bank accounts request", slog.Any("error", err.Error()))
		return nil, fmt.Errorf("failed get data form check bank accounts request: %d", resp.StatusCode)
	}

	return result.Data, nil
}

func KeepAlive(ctx context.Context, c *http.Client) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				req, _ := http.NewRequestWithContext(ctx, http.MethodGet, endpoints.CryptoBotVersion(), nil)
				if _, err := c.Do(req); err != nil {
					slog.Error("keepAlive error", slog.Any("error", err))
				}
			}
		}
	}()
}
