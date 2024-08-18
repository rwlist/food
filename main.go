package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/jackc/pgx/v5"
	"github.com/rwlist/food/internal/blobs"
	"github.com/rwlist/food/internal/chbot"
	"github.com/rwlist/food/internal/conf"
	"github.com/rwlist/food/internal/db"
	"github.com/rwlist/food/internal/seed"
	"github.com/rwlist/food/internal/web"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	cfg, err := conf.ParseEnv()
	if err != nil {
		slog.Error("failed to parse config from env", "err", err.Error())
		os.Exit(1)
	}

	// handle interrupt signal
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		slog.Info("received interrupt signal")
		cancel()
	}()

	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.R2KeyID, cfg.R2KeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		slog.Error("failed to load cloudflare config", "err", err.Error())
		os.Exit(1)
	}

	s3Client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID))
	})

	conn, err := pgx.Connect(ctx, cfg.PostgresURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "err", err.Error())
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	queries := db.New(conn)

	cmd := ""
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}

	photos := blobs.NewPhotos(s3Client, cfg.R2BucketName)

	switch cmd {
	case "server":
		// start http server
		server := web.NewServer(cfg, queries, photos)
		err := server.Start(ctx)
		if err != nil {
			slog.Error("server error", "err", err.Error())
		}
	case "bot":
		chbot.StartBot(ctx, cfg)
	case "import":
		// import data from archive.json
		dir := os.Args[2]
		archiver := seed.ArchiveUploader{
			Dir:     dir,
			Queries: queries,
			Photos:  photos,
		}
		err := archiver.LoadFromArchive(ctx)
		if err != nil {
			slog.Error("failed to import data from archive", "err", err.Error())
		}
	default:
		slog.Error("unknown command", "cmd", cmd)
		os.Exit(1)
	}
}
