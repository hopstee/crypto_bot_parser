package tgbot

import (
	"context"
	"crypto_bot_parser/internal/cryptobot/http"
	"fmt"
	"log/slog"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) handleCallback(ctx context.Context, q *tgbotapi.CallbackQuery) {
	chatID := q.Message.Chat.ID
	user := b.UserStore.GetOrCreate(chatID)

	if err := b.ensureUserRuntime(user); err != nil {
		b.sendMessage(chatID, "Ошибка инициализации пользователя")
		return
	}

	// if user.HasActiveRequest.Load() {
	// 	return
	// }

	callbackParts := strings.Split(q.Data, ":")
	var dealID string
	if len(callbackParts) > 1 {
		dealID = callbackParts[1]
	}

	switch {
	case strings.HasPrefix(q.Data, "deal_paid:"):
		b.finishDeal(ctx, user, chatID, true, dealID)

	case strings.HasPrefix(q.Data, "deal_cancel:"):
		b.finishDeal(ctx, user, chatID, false, dealID)

	case strings.HasPrefix(q.Data, "deal_status:"):
		b.checkDealStatus(ctx, user, chatID, dealID)
	}

	answer := tgbotapi.NewCallback(q.ID, "")
	if _, err := b.api.Request(answer); err != nil {
		slog.Error("failed to answer callback", slog.Any("error", err.Error()))
	}
}

func (b *Bot) checkDealStatus(ctx context.Context, user *User, chatID int64, dealID string) {
	if dealID == "" {
		return
	}

	resp, err := http.CheckDeal(ctx, dealID, user.userAgent, user.client)
	if err != nil {
		b.sendMessage(chatID, "❌ Не получилось проверить статус заявки")
		return
	}
	b.sendMessage(chatID, fmt.Sprintf("📋 Заявка находится в статусе %s", resp.Status))
}

func (b *Bot) finishDeal(ctx context.Context, user *User, chatID int64, paid bool, dealID string) {
	if dealID == "" {
		return
	}

	if paid {
		if err := http.ConfirmDeal(ctx, dealID, user.Data.CryptoBotBankAccounts[0].ID, user.userAgent, user.client); err != nil {
			b.sendMessage(chatID, "❌ Не получилось подтвердить оплату заявки")
			return
		}
		b.sendMessage(chatID, "🤝 Заявка принята как оплаченная")
	} else {
		if err := http.CancelDeal(ctx, dealID, user.userAgent, user.client); err != nil {
			b.sendMessage(chatID, "❌ Не получилось отклонить оплату заявки")
			return
		}
		b.sendMessage(chatID, "🗑️ Заявка отклонена")
	}

	user.HasActiveRequest.Store(false)
}
