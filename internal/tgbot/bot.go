package tgbot

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type SeenOrder struct {
	expiresAt string
}

type Bot struct {
	api       *tgbotapi.BotAPI
	updates   tgbotapi.UpdatesChannel
	UserStore *UserStore

	botID      string
	secretCode string
	dataDir    string

	seenOrders sync.Map
	wg         *sync.WaitGroup
}

func NewBot(botID, secretCode, dataDir string, wg *sync.WaitGroup) *Bot {
	return &Bot{
		UserStore: NewUserStore(dataDir),

		botID:      botID,
		secretCode: secretCode,
		dataDir:    dataDir,

		wg: wg,
	}
}

func (b *Bot) Init(token string) error {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("failed init bot api: %w", err)
	}

	api.Debug = true
	slog.Info("telegram bot started", slog.String("username", api.Self.UserName))

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := api.GetUpdatesChan(u)

	b.api = api
	b.updates = updates

	if err := b.registerCommands(); err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	return nil
}

func (b *Bot) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ctx.Done():
			b.Stop()
			return
		case <-ticker.C:
			b.clearSeenOrders()
		case u := <-b.updates:
			switch {
			case u.Message != nil:
				b.handleMessage(ctx, u.Message)
			case u.CallbackQuery != nil:
				b.handleCallback(ctx, u.CallbackQuery)
			}
		}
	}
}

func (b *Bot) Stop() {
	slog.Info("stop bot processes")
	b.api.StopReceivingUpdates()

	for _, user := range b.UserStore.users {
		if user.wsCancel != nil {
			user.wsCancel()
		}

		if user.client != nil && user.session != nil {
			b.closeClient(user.client)
		}
	}

	b.seenOrders.Range(func(key, value any) bool {
		b.seenOrders.Delete(key)
		return true
	})

	if b.wg != nil {
		b.wg.Wait()
	}
}
