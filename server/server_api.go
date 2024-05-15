package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

func (s *server) postApiUpdate(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tagID := string(body)
	slog.Info("received new tag", "tag", tagID)
	w.WriteHeader(http.StatusOK)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		tag, err := s.db.GetTag(ctx, tagID)
		if err != nil {
			if errors.Is(err, errNoItem) {
				s.tagMu.Lock()
				s.tag = tagID
				s.tagMu.Unlock()
				s.sendToTelegram("Unknown tag scanned. Go to the web interface to connect it.")
			} else {
				slog.Error("could not load album", "error", err)
				s.sendToTelegram(fmt.Sprintf("Could not load albums: %s", err))
			}

			return
		}

		s.sendToTelegram("Scanned vinyl " + tag.Album.String())
		err = s.db.CreateLog(ctx, tag.Album)
		if err != nil {
			slog.Error("could not log album", "error", err)
			s.sendToTelegram(fmt.Sprintf("Could not log album: %s", err))
		}
	}()
}

func mustApiToken(token string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Token "+token {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
