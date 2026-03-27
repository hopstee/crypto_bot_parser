package tgbot

import (
	cryptobothttp "crypto_bot_parser/internal/cryptobot/http"
	"crypto_bot_parser/internal/httpclient"
	"crypto_bot_parser/internal/session"
	"crypto_bot_parser/pkg/helpers/crypto"
	"crypto_bot_parser/pkg/helpers/useragent"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"time"
)

func (b *Bot) closeClient(c *http.Client) {
	if tr, ok := c.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
}

func (b *Bot) ensureUserRuntime(user *User) error {
	if user.client != nil && user.session != nil {
		return nil
	}

	if user.Data.Phone == "" {
		user.Data.State = StateWaitPhone
		return fmt.Errorf("user has no phone")
	}

	sessPath := filepath.Join(b.dataDir, user.Data.Phone, "session.json")
	user.session = session.NewSession(sessPath)

	if user.userAgent != "" {
		user.Data.UserAgent = user.userAgent
	} else {
		user.userAgent = useragent.RandomUserAgent()
		user.Data.UserAgent = user.userAgent
		b.UserStore.Save(user.Data)
	}

	user.client = httpclient.NewHttpClient(user.session)
	user.session.ApplyToClient(user.client)

	slog.Info("user checked")

	return nil
}

func (b *Bot) formatDeal(dealData *cryptobothttp.TakeDealData) string {
	return fmt.Sprintf(
		"Валюта оплаты: %s\nОбъем: %s\n\nВалюта получения: %s\nОбъем: %s\n\nКомиссия: %s\n\nСсылка на оплату: %s",
		dealData.InAsset,
		dealData.InAmount,
		dealData.OutAsset,
		crypto.ParseCryptoAmount(dealData.OutAmount),
		crypto.ParseCryptoAmount(dealData.FeeAmount),
		dealData.URL,
	)
}

func (b *Bot) dedupKey(id, expAt string) string {
	return id + "|" + expAt
}

func (b *Bot) clearSeenOrders() {
	b.seenOrders.Range(func(key, value any) bool {
		order, ok := value.(SeenOrder)
		if ok && b.orderExpired(order.expiresAt) {
			b.seenOrders.Delete(key)
		}
		return true
	})
}

func (b *Bot) orderExpired(expiresAt string) bool {
	t, err := time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return true
	}
	return time.Now().UTC().After(t)
}
