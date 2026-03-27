package tgbot

import (
	"context"
	"crypto_bot_parser/internal/cryptobot/http"
	"crypto_bot_parser/pkg/helpers/markdown"
	"fmt"
)

type BotAuthListener struct {
	ctx    context.Context
	bot    *Bot
	chatID int64
}

func NewBotAuthListener(c context.Context, b *Bot, cid int64) *BotAuthListener {
	return &BotAuthListener{ctx: c, bot: b, chatID: cid}
}

func (b *BotAuthListener) OnWaitingConfirmation() {
	b.bot.sendMessage(
		b.chatID,
		"🚨 Подтвердите аутентификацию в официальном чате TG",
	)
}

func (b *BotAuthListener) OnAuthorized() {
	b.bot.sendMessage(
		b.chatID,
		"✅ Вы успешно аутентифицированы",
	)

	u := b.bot.UserStore.GetOrCreate(b.chatID)

	if u.keepAliveCancel != nil {
		u.keepAliveCancel()
	}

	ctx, cancel := context.WithCancel(b.ctx)
	u.keepAliveCtx = ctx
	u.keepAliveCancel = cancel

	http.KeepAlive(ctx, u.client)
}

func (b *BotAuthListener) OnError(err error) {
	text := fmt.Sprintf("❌ %s\n🔁 Введите телефон повторно", markdown.EscapeMarkdownV2(err.Error()))
	b.bot.sendMessage(
		b.chatID,
		text,
	)
}
