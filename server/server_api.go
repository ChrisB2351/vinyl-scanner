package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gorm.io/gorm"
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

		album, err := s.db.GetAlbumByTag(ctx, tagID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				link := fmt.Sprintf("%s/albums/new?log=true&tag=%s", s.baseURL, tagID)
				message := fmt.Sprintf("Unknown tag scanned: %s.\n\nCreate new album at %s.", tagID, link)
				s.sendToTelegram(message)
			} else {
				slog.Error("could not load album", "error", err)
				s.sendToTelegram(fmt.Sprintf("Could not load albums: %s", err))
			}

			return
		}

		s.sendToTelegram("Scanned vinyl " + album.String())
		err = s.db.CreateLog(ctx, album)
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
