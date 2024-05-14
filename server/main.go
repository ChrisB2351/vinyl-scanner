package main

import (
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/urfave/cli/v2"
)

func main() {
	err := godotenv.Load()
	if os.IsNotExist(err) {
		log.Printf("no .env file found, skipping")
	} else if err != nil {
		log.Fatalf("failed loading .env file: %s", err)
	}

	app := cli.NewApp()
	app.Name = "vinyl-server"
	app.Usage = "Vinyl scanner server and storage."
	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Value:   8080,
			Usage:   "port to run server on",
			EnvVars: []string{"VINYL_PORT"},
		},
		&cli.StringFlag{
			Name:     "telegram-token",
			Usage:    "telegram bot token",
			EnvVars:  []string{"VINYL_TG_TOKEN"},
			Required: true,
		},
		&cli.Int64SliceFlag{
			Name:     "telegram-chat-id",
			Usage:    "telegram bot chat id or comma-separated ids",
			EnvVars:  []string{"VINYL_TG_CHAT_ID"},
			Required: true,
		},
		&cli.StringFlag{
			Name:     "data-directory",
			Usage:    "data directory where the logs and the vinyl data is stored",
			EnvVars:  []string{"VINYL_DATA_DIR"},
			Required: true,
		},
		&cli.StringFlag{
			Name:    "auth-token",
			Usage:   "http server endpoint authentication token",
			EnvVars: []string{"VINYL_AUTH_TOKEN"},
		},
	}
	app.Action = func(ctx *cli.Context) error {
		handler, err := newServer(ctx.String("telegram-token"), ctx.Int64Slice("telegram-chat-id"), ctx.String("auth-token"), ctx.String("data-directory"))
		if err != nil {
			return err
		}

		// Start HTTP handler.
		quit := make(chan os.Signal, 2)
		var wg sync.WaitGroup
		wg.Add(1)

		server := &http.Server{Addr: ":" + strconv.Itoa(ctx.Int("port")), Handler: handler}

		go func() {
			defer wg.Done()

			slog.Info("serving", "address", server.Addr)

			err := server.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				fmt.Fprintf(os.Stderr, "failed to start server: %s\n", err)
				quit <- os.Interrupt
			}
		}()

		signal.Notify(
			quit,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGHUP,
		)
		<-quit

		slog.Info("Server shutting down...")

		go server.Close()

		wg.Wait()
		return nil
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
