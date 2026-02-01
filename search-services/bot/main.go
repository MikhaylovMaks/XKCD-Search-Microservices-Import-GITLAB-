package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"yadro.com/course/bot/config"
	"yadro.com/course/bot/handlers"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)

	log := mustMakeLogger(cfg.LogLevel)

	if cfg.BotToken == "" {
		log.Error("BOT_TOKEN is not set")
		os.Exit(1)
	}

	if err := run(cfg, log); err != nil {
		log.Error("failed to run bot", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting telegram bot")
	log.Debug("debug messages are enabled")

	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return fmt.Errorf("failed to create bot: %v", err)
	}

	bot.Debug = cfg.LogLevel == "DEBUG"
	log.Info("authorized", "username", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("shutting down bot")
		bot.StopReceivingUpdates()
		os.Exit(0)
	}()

	// Process updates
	log.Info("bot is running, waiting for updates...")
	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Debug("received message",
			"chat_id", update.Message.Chat.ID,
			"text", update.Message.Text,
			"from", update.Message.From.UserName,
		)

		// Handle commands
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				handlers.HandleStartCommand(bot, update)
			case "help":
				handlers.HandleHelpCommand(bot, update)
			case "search":
				handlers.HandleSearchCommand(bot, update, cfg.APIAddress, log)
			default:
				handlers.HandleUnknownCommand(bot, update)
			}
		} else {
			chatID := update.Message.Chat.ID
			msg := tgbotapi.NewMessage(chatID, "💡 Используйте команду /search <фраза> для поиска комиксов.\n\nПример: /search quantum")
			bot.Send(msg)
		}
	}

	return nil
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level, AddSource: true})
	return slog.New(handler)
}
