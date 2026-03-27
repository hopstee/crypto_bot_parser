package tgbot

import (
	"context"
	"crypto_bot_parser/internal/auth"
	cryptobothttp "crypto_bot_parser/internal/cryptobot/http"
	"crypto_bot_parser/internal/cryptobot/ws"
	"crypto_bot_parser/pkg/helpers/converter"
	"crypto_bot_parser/pkg/helpers/markdown"
	"crypto_bot_parser/pkg/helpers/qr"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

var upstreamBlockedUntil atomic.Int64

func (b *Bot) runAuth(rootCtx context.Context, chatID int64, user *User) {
	ctx, cancel := context.WithCancel(rootCtx)
	defer cancel()

	authFlow := auth.NewAuth(
		user.client,
		user.session,
		user.userAgent,
		user.Data.Phone,
		b.botID,
	)
	authFlow.SetListener(NewBotAuthListener(rootCtx, b, chatID))

	if err := authFlow.Run(ctx); err != nil {
		user.Data.State = StateWaitPhone
		return
	}

	user.Data.State = StateAuthenticated
	user.Data.CryptoBotSettings = authFlow.GetCryptoBotSettings()

	_ = b.UserStore.Save(user.Data)
}

func (b *Bot) startWS(rootCtx context.Context, chatID int64, user *User) {
	if user.wsRunning.Load() {
		b.sendMessage(chatID, "❕ WebSocket уже запущен")
		return
	}

	if user.Data.CryptoBotSettings == nil || !user.Data.CryptoBotSettings.IsP2cMerchant {
		b.sendMessage(chatID, "🚧 WebSocket не может быть запущен\\. Вы не являетесь P2C мерчантом\\.")
		return
	}

	cryptoBotBankAccounts, err := cryptobothttp.CheckBankAccounts(rootCtx, user.userAgent, user.client)
	if err != nil {
		b.sendMessage(chatID, "❌ Не удается получить ваши банковские аккануты\\. Попробуйте еще раз\\.")
		return
	}

	if len(cryptoBotBankAccounts) == 0 {
		b.sendMessage(chatID, "🚧 WebSocket не может быть запущен\\. У вас нет подключенных способов оплаты\\.")
		return
	}
	b.UserStore.SetCryptoBotBankAccounts(chatID, cryptoBotBankAccounts)

	ctx, cancel := context.WithCancel(rootCtx)
	user.wsCtx = ctx
	user.wsCancel = cancel
	user.wsRunning.Store(true)

	b.wg.Go(func() {
		defer func() {
			user.wsCtx = nil
			user.wsCancel = nil
			user.wsRunning.Store(false)
		}()

		cryptoBot := ws.NewClient(
			user.client,
			user.userAgent,
			int(user.Data.WsConnections),
			func(data *ws.ListUpdateData) {
				b.preceedWsData(rootCtx, chatID, user, data)
			},
		)
		cryptoBot.Run(user.wsCtx, &upstreamBlockedUntil)
	})

	b.sendMessage(chatID, "👍 Прослушка WebSocket запущена")
}

func (b *Bot) stopWS(chatID int64, user *User, bad bool) {
	if !user.wsRunning.Load() || user.wsCancel == nil {
		b.sendMessage(chatID, "❕ WebSocket не запущен")
		return
	}

	user.wsCancel()

	user.wsRunning.Store(false)
	user.HasActiveRequest.Store(false)

	if bad {
		b.sendMessage(chatID, "🚨Прослушка WebSocket остановлена, нужна повторная аутентификация🚨")
	}
	b.sendMessage(chatID, "👍 Прослушка WebSocket остановлена")
}

func (b *Bot) statusWS(chatID int64, user *User) {
	if !user.wsRunning.Load() || user.wsCancel == nil {
		b.sendMessage(chatID, "🤷‍♂️ WebSocket не запущен")
		return
	} else if user.wsRunning.Load() {
		b.sendMessage(chatID, "👷🏻‍♂️ WebSocket запущен")
		return
	}

	msg := fmt.Sprintf(
		"🐳 Не известный статус WebSocket\\. Перезапустите бота /%s",
		StartCommand,
	)
	b.sendMessage(
		chatID,
		markdown.EscapeMarkdownV2(msg),
	)
}

func (b *Bot) preceedWsData(ctx context.Context, chatID int64, user *User, data *ws.ListUpdateData) {
	if until := upstreamBlockedUntil.Load(); until > 0 {
		if time.Now().UnixNano() < until {
			return
		}
	}

	orderKey := b.dedupKey(data.ID, data.ExpiresAt)
	if _, loaded := b.seenOrders.LoadOrStore(orderKey, SeenOrder{expiresAt: data.ExpiresAt}); loaded {
		return
	}

	if user.HasActiveRequest.Load() {
		return
	}

	amount := converter.FastAmount(data.InAmount)
	if (user.Data.MinAmount > 0 && amount < user.Data.MinAmount) ||
		(user.Data.MaxAmount > 0 && amount > user.Data.MaxAmount) {
		return
	}

	takeDealResponseData, err := cryptobothttp.TakeDeal(
		ctx,
		data.ID,
		user.userAgent,
		user.client,
	)

	if err != nil {
		if errors.Is(err, cryptobothttp.ErrCloudflareBadResponse) {
			blockUntil := time.Now().Add(2 * time.Minute).UnixNano()
			upstreamBlockedUntil.Store(blockUntil)

			slog.Warn(
				"upstream unstable, blocking takeDeal",
				slog.Duration("block_for", 2*time.Minute),
			)
		}
		if errors.Is(err, cryptobothttp.ErrForbidden) {
			blockUntil := time.Now().Add(2 * time.Minute).UnixNano()
			upstreamBlockedUntil.Store(blockUntil)

			slog.Warn(
				"user temporary blocked",
				slog.Duration("for", 2*time.Minute),
			)
		}
		if errors.Is(err, cryptobothttp.ErrUnauthorized) {
			b.stopWS(chatID, user, true)
		}
		slog.Warn("failed take deal", slog.Any("error", err.Error()))
		return
	}

	user.HasActiveRequest.Store(true)

	go func() {
		qrCodePng, err := qr.GenerateQRPNG(takeDealResponseData.URL, 256)
		if err != nil {
			user.HasActiveRequest.Store(false)
			slog.Error("failed to generate qr code", slog.Any("error", err.Error()))
			return
		}

		b.sendDealMessage(chatID, qrCodePng, b.formatDeal(takeDealResponseData), takeDealResponseData.ID)
	}()
}
