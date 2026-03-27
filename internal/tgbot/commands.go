package tgbot

import (
	"context"
	"crypto_bot_parser/internal/cryptobot/http"
	"crypto_bot_parser/pkg/helpers/markdown"
	"crypto_bot_parser/pkg/helpers/phone"
	"fmt"
	"log/slog"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotCommand string

const (
	StartCommand          BotCommand = "start"
	SettingsCommand       BotCommand = "settings"
	UpdateSettingsCommand BotCommand = "update_settings"
	MinCommand            BotCommand = "min"
	MaxCommand            BotCommand = "max"
	ConnectionsCommand    BotCommand = "connections"
	StartWSCommand        BotCommand = "start_ws"
	StopWSCommand         BotCommand = "stop_ws"
	StatusWSCommand       BotCommand = "status_ws"
	ResetAuthCommand      BotCommand = "reset_auth"
)

var botCommands = []tgbotapi.BotCommand{
	{
		Command:     string(StartCommand),
		Description: "Запуск бота",
	},
	{
		Command:     string(StartWSCommand),
		Description: "Запустить прослушку веб-сокета",
	},
	{
		Command:     string(StopWSCommand),
		Description: "Остановить прослушку веб-сокета",
	},
	{
		Command:     string(StatusWSCommand),
		Description: "Статус веб-сокета",
	},
	{
		Command:     string(SettingsCommand),
		Description: "Текущие настройки",
	},
	{
		Command:     string(UpdateSettingsCommand),
		Description: "Обновить настройки из крипто бота",
	},
	{
		Command:     string(MinCommand),
		Description: "Установить минимальную сумму",
	},
	{
		Command:     string(MaxCommand),
		Description: "Установить максимальную сумму",
	},
	{
		Command:     string(ConnectionsCommand),
		Description: "Установить количество подключений",
	},
	{
		Command:     string(ResetAuthCommand),
		Description: "Повторить аутентификацию",
	},
}

func (b *Bot) handleMessage(ctx context.Context, msg *tgbotapi.Message) {
	chatID := msg.From.ID
	user := b.UserStore.GetOrCreate(chatID)

	switch user.Data.State {
	case StateNew:
		if msg.Command() == string(StartCommand) {
			b.handleStart(user, chatID)
		} else {
			msg := fmt.Sprintf(
				"▶️ Запустите бота командой /%s",
				StartCommand,
			)
			b.sendMessage(
				chatID,
				markdown.EscapeMarkdownV2(msg),
			)
		}
	case StateWaitSecret:
		b.handleSecrete(msg, user, chatID)
	case StateWaitPhone:
		b.handlePhone(ctx, msg, user, chatID)
	case StateAuthInProgress:
		b.sendMessage(chatID, "⏳ Аутентификация в процессе\n⏳ Пожалуйста подождите")
	case StateAuthenticated:
		if err := b.ensureUserRuntime(user); err != nil {
			b.sendMessage(chatID, "❌ Ошибка инициализации сессии")
			return
		}

		if msg.Command() != "" {
			b.handleCommand(ctx, msg, chatID, user)
			return
		}

		b.sendMessage(chatID, "✅ Вы аутентифицированы\n👷🏻‍♂️ Бот запущен\n📥 Ожидайте запросов на оплату")
	default:
		msg := fmt.Sprintf(
			"🤷‍♂️ Неизвестное состояние бота\n▶️ Перезапустите пожалуйста командой /%s",
			StartCommand,
		)
		b.sendMessage(
			chatID,
			markdown.EscapeMarkdownV2(msg),
		)
	}
}

func (b *Bot) handleStart(user *User, chatID int64) {
	switch user.Data.State {
	case StateNew:
		b.sendMessage(chatID, "🤫 Для начала введите секретный код:")
		user.Data.State = StateWaitSecret
	case StateAuthenticated:
		b.sendMessage(chatID, "_Вы аутентифицированы\\. Бот запущен\\. Ожидайте запросов на оплату\\. Если запросы не приходят, проверьте статус вебхука\\._")
	}
}

func (b *Bot) handleSecrete(msg *tgbotapi.Message, user *User, chatID int64) {
	if msg.Text != b.secretCode {
		b.sendMessage(chatID, "🚨🔑 Неверный код доступа")
		return
	}

	user.Data.State = StateWaitPhone
	b.sendMessage(chatID, "_Разрешение получено_\n📞 Введите номер телефона:")
}

func (b *Bot) handlePhone(ctx context.Context, msg *tgbotapi.Message, user *User, chatID int64) {
	phone, err := phone.NormalizeRUPhone(msg.Text)
	if err != nil {
		b.sendMessage(chatID, "📵 Не верный формат телефона\\. Попробуйте еще:")
		return
	}

	user.Data.Phone = phone
	user.Data.State = StateAuthInProgress

	if err := b.ensureUserRuntime(user); err != nil {
		b.sendMessage(chatID, "❌ Ошибка инициализации сессии")
		return
	}

	b.sendMessage(chatID, "⏳ Начал аутентификацию в CryptoBot")

	go b.runAuth(ctx, chatID, user)
}

func (b *Bot) handleCommand(ctx context.Context, msg *tgbotapi.Message, chatID int64, user *User) {
	switch msg.Command() {
	case string(StartCommand):
		b.handleStart(user, chatID)

	case string(SettingsCommand):
		isP2cMerchant := "Нет"
		if user.Data.CryptoBotSettings.IsP2cMerchant {
			isP2cMerchant = "Да"
		}

		msg := fmt.Sprintf(
			"⚙️ Настройки\nP2C доступно: *%s*\nP2C мин\\. сумма: *%d*\nP2C макс\\. сумма: *%d*\nКоличество соединений: *%d*",
			isP2cMerchant,
			user.Data.MinAmount,
			user.Data.MaxAmount,
			user.Data.WsConnections,
		)

		b.sendMessage(
			chatID,
			msg,
		)

	case string(UpdateSettingsCommand):
		userSettings, err := http.CheckUserSettings(ctx, user.userAgent, user.client)
		if err != nil {
			msg := fmt.Sprintf(
				"❌ Не получилось запросить настройки из крипто бота\nОбновите данные аутентификации /%s",
				ResetAuthCommand,
			)
			b.sendMessage(
				chatID,
				markdown.EscapeMarkdownV2(msg),
			)
			return
		}
		b.UserStore.SetCryptoBotSettings(chatID, userSettings)
		b.sendMessage(chatID, "✅ Ваши настройки из крипто бота обновленны")

	case string(MinCommand):
		if len(msg.CommandArguments()) == 0 {
			msg := fmt.Sprintf(
				"🙈 Пример: /%s <сумма>",
				MinCommand,
			)
			b.sendMessage(
				chatID,
				markdown.EscapeMarkdownV2(msg),
			)
			return
		}
		v, err := strconv.ParseInt(msg.CommandArguments(), 10, 64)
		if err != nil {
			b.sendMessage(chatID, "❌ Неверное число")
			return
		}
		b.UserStore.SetMin(chatID, v)
		b.sendMessage(chatID, fmt.Sprintf("✅ Минималь сумма выставлена в %d", v))

	case string(MaxCommand):
		if len(msg.CommandArguments()) == 0 {
			msg := fmt.Sprintf(
				"🙈 Пример: /%s <сумма>",
				MaxCommand,
			)
			b.sendMessage(
				chatID,
				markdown.EscapeMarkdownV2(msg),
			)
			return
		}
		v, err := strconv.ParseInt(msg.CommandArguments(), 10, 64)
		if err != nil {
			b.sendMessage(chatID, "❌ Неверное число")
			return
		}
		b.UserStore.SetMax(chatID, v)
		b.sendMessage(chatID, fmt.Sprintf("✅ Максимальная сумма выставлена в %d", v))

	case string(ConnectionsCommand):
		if len(msg.CommandArguments()) == 0 {
			msg := fmt.Sprintf(
				"🙈 Пример: /%s <количество>",
				ConnectionsCommand,
			)
			b.sendMessage(
				chatID,
				markdown.EscapeMarkdownV2(msg),
			)
			return
		}
		v, err := strconv.ParseInt(msg.CommandArguments(), 10, 64)
		if err != nil {
			b.sendMessage(chatID, "❌ Неверное число")
			return
		}
		b.UserStore.SetWsConnections(chatID, v)
		b.sendMessage(chatID, fmt.Sprintf("✅ Число соединений выставлено в %d", v))

	case string(StartWSCommand):
		b.startWS(ctx, chatID, user)

	case string(StopWSCommand):
		b.stopWS(chatID, user, false)

	case string(StatusWSCommand):
		b.statusWS(chatID, user)

	case string(ResetAuthCommand):
		msg.Text = user.Data.Phone
		b.handlePhone(ctx, msg, user, chatID)

	default:
		b.sendMessage(chatID, "🤷‍♂️ Не известная команда")
	}
}

func (b *Bot) registerCommands() error {
	cfg := tgbotapi.NewSetMyCommands(botCommands...)

	_, err := b.api.Request(cfg)
	if err != nil {
		return err
	}

	slog.Info("telegram commands registered")
	return nil
}
