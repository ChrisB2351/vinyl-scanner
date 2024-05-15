package main

import (
	"encoding/base64"
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
	"golang.org/x/crypto/bcrypt"
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
			Name:    "api-token",
			Usage:   "api endpoint authentication token",
			EnvVars: []string{"VINYL_API_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "jwt-secret",
			Usage:   "jwt tokens secret",
			EnvVars: []string{"VINYL_JWT_SECRET"},
		},
		&cli.StringFlag{
			Name:    "login-username",
			Usage:   "admin interface username",
			EnvVars: []string{"VINYL_LOGIN_USERNAME"},
		},
		&cli.StringFlag{
			Name:    "login-password",
			Usage:   "admin interface base64 hashed password generated with 'password' subcommand",
			EnvVars: []string{"VINYL_LOGIN_PASSWORD"},
		},
	}
	app.Action = func(ctx *cli.Context) error {
		cfg := &config{
			tgToken:   ctx.String("telegram-token"),
			tgChatIDs: ctx.Int64Slice("telegram-chat-id"),
			apiToken:  ctx.String("api-token"),
			dataDir:   ctx.String("data-directory"),
			jwtSecret: ctx.String("jwt-secret"),
			username:  ctx.String("login-username"),
			password:  ctx.String("login-password"),
		}

		handler, err := newServer(cfg)
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

	app.Commands = append(app.Commands, &cli.Command{
		Name:      "password",
		Usage:     "Generate a password hash to use on the configuration",
		Args:      true,
		ArgsUsage: "[password]",
		Before: func(ctx *cli.Context) error {
			if ctx.NArg() != 1 {
				return errors.New("this command must have one and only one argument")
			}
			return nil
		},
		Action: func(ctx *cli.Context) error {
			pwd := ctx.Args().First()
			hash, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
			if err != nil {
				return err
			}

			fmt.Println(base64.StdEncoding.EncodeToString(hash))
			return nil
		},
	})

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
