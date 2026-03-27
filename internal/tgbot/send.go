package tgbot

import (
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("failed to send message", slog.Any("error", err.Error()))
	}
}

func (b *Bot) sendDealMessage(
	chatID int64,
	qr []byte,
	text string,
	dealID int64,
) {
	dealIDstr := strconv.FormatInt(dealID, 10)
	kb := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Оплатил", "deal_paid:"+dealIDstr),
			tgbotapi.NewInlineKeyboardButtonData("⏳ Статус", "deal_status:"+dealIDstr),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отказ", "deal_cancel:"+dealIDstr),
		),
	)

	msg := tgbotapi.NewPhoto(
		chatID,
		tgbotapi.FileBytes(tgbotapi.FileBytes{
			Name:  "",
			Bytes: qr,
		}),
	)
	msg.Caption = text
	msg.ReplyMarkup = kb

	if _, err := b.api.Send(msg); err != nil {
		slog.Error("send deal message failed", slog.Any("error", err.Error()))
	}
}
