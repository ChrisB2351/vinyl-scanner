package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"

	"github.com/joho/godotenv"
)

func authMiddleware(next http.Handler, token string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Token "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func sendToTelegram(token, chatID, uid string) {
	data := url.Values{}
	data.Set("chat_id", chatID)
	data.Set("disable_web_page_preview", "true")
	data.Set("text", fmt.Sprintf("Read %s", uid))

	u, err := url.Parse(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token))
	if err != nil {
		slog.Warn("failed to send make telegram url", "error", err)
		return
	}
	u.RawQuery = data.Encode()

	resp, err := http.DefaultClient.Get(u.String())
	if err != nil {
		slog.Warn("failed to send request to telegram", "error", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Warn("unexpected status code from telegram", "statusCode", resp.StatusCode)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		id := string(body)
		slog.Info("received new uid", "uid", id)

		sendToTelegram(os.Getenv("TG_TOKEN"), os.Getenv("TG_CHAT_ID"), id)
		w.WriteHeader(http.StatusOK)
	})

	var handler http.Handler = mux

	if secret := os.Getenv("TOKEN"); secret != "" {
		handler = authMiddleware(mux, secret)
	}

	http.ListenAndServe(":8080", handler)
}
