package chbot

import (
	"context"
	"os"

	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/rwlist/food/internal/conf"
)

func StartBot(ctx context.Context, cfg *conf.App) {
	opts := []bot.Option{
		bot.WithDebug(),
		bot.WithDefaultHandler(handler),
	}

	b, err := bot.New(cfg.TelegramBotToken, opts...)
	if err != nil {
		slog.Error("failed to create bot", "err", err.Error())
		os.Exit(1)
	}

	b.Start(ctx)
}

func handler(ctx context.Context, b *bot.Bot, upd *models.Update) {
	slog.Info("received update", "update", upd)
}
