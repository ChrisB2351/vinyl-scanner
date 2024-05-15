package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type config struct {
	tgToken   string
	tgChatIDs []int64

	apiToken string
	dataDir  string

	jwtSecret string
	username  string
	password  string
}

type server struct {
	mux *chi.Mux
	db  *database

	tagMu sync.Mutex
	tag   string

	tgToken   string
	tgChatIDs []string

	jwtAuth  *jwtauth.JWTAuth
	username string
	password string
}

func newServer(cfg *config) (*server, error) {
	err := os.MkdirAll(cfg.dataDir, 0777)
	if err != nil {
		return nil, err
	}

	db, err := newDatabase(filepath.Join(cfg.dataDir, "data.db"))
	if err != nil {
		return nil, err
	}

	s := &server{
		mux:      chi.NewRouter(),
		db:       db,
		tgToken:  cfg.tgToken,
		jwtAuth:  jwtauth.New("HS256", []byte(base64.StdEncoding.EncodeToString([]byte(cfg.jwtSecret))), nil),
		username: cfg.username,
		password: cfg.password,
	}
	for _, chatID := range cfg.tgChatIDs {
		s.tgChatIDs = append(s.tgChatIDs, strconv.FormatInt(chatID, 10))
	}

	// Build Router
	s.mux.Use(jwtauth.Verifier(s.jwtAuth))
	s.mux.Get("/assets*", http.FileServer(http.FS(assetsFS)).ServeHTTP)
	s.mux.Get("/login", s.loginGet)
	s.mux.Post("/login", s.loginPost)
	s.mux.Get("/logout", s.logoutGet)
	s.mux.Group(func(r chi.Router) {
		r.Use(s.mustLoggedIn)
		r.Get("/", s.getIndex)
		r.Get("/albums", s.getAlbums)
		r.Get("/albums/new", s.getNewAlbum)
		r.Post("/albums/new", s.postNewAlbum)
		r.Get("/albums/{album-id}", s.getAlbum)
		r.Post("/albums/{album-id}", s.postAlbum)
		r.Get("/albums/{album-id}/delete", s.getDeleteAlbum)
		r.Post("/albums/{album-id}/delete", s.postDeleteAlbum)

		r.Get("/tags", s.getTags)
		r.Get("/tags/{tag-id}/connect/{album-id}", s.getTagConnect)
		r.Post("/tags/{tag-id}/connect/{album-id}", s.postTagConnect)
		r.Get("/tags/{tag-id}/delete", s.getDeleteTag)
		r.Post("/tags/{tag-id}/delete", s.postDeleteTag)

		r.Get("/logs", s.getLogs)
		r.Get("/logs/{ts}/delete", s.getDeleteLog)
		r.Post("/logs/{ts}/delete", s.postDeleteLog)
	})
	s.mux.Group(func(r chi.Router) {
		if cfg.apiToken == "" {
			slog.Warn("authorization token not set, api endpoint is unprotected")
		} else {
			r.Use(mustApiToken(cfg.apiToken))
		}
		r.Post("/api/tag", s.postApiUpdate)
	})

	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *server) getIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/albums", http.StatusTemporaryRedirect)
}

func (s *server) getAlbums(w http.ResponseWriter, r *http.Request) {
	var (
		tag string
	)

	s.tagMu.Lock()
	tag = s.tag
	s.tagMu.Unlock()

	albums, err := s.db.GetAlbums(r.Context())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "albums.html", map[string]interface{}{
		"Title":  "Albums",
		"Albums": albums,
		"Tag":    tag,
	})
}

func (s *server) getNewAlbum(w http.ResponseWriter, r *http.Request) {
	s.renderTemplate(w, http.StatusOK, "album.html", map[string]interface{}{
		"Title": "New Album",
		"Album": &Album{},
	})
}

func (s *server) postNewAlbum(w http.ResponseWriter, r *http.Request) {
	s.createOrUpdateAlbum(w, r, nil)
}

func (s *server) getAlbum(w http.ResponseWriter, r *http.Request) {
	id, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	album, err := s.db.GetAlbum(r.Context(), id)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "album.html", map[string]interface{}{
		"Title": "Update Album",
		"Album": album,
	})
}

func (s *server) postAlbum(w http.ResponseWriter, r *http.Request) {
	id, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	_, err = s.db.GetAlbum(r.Context(), id)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.createOrUpdateAlbum(w, r, &id)
}

func (s *server) createOrUpdateAlbum(w http.ResponseWriter, r *http.Request, id *uint64) {
	err := r.ParseForm()
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	name := strings.TrimSpace(r.Form.Get("name"))
	artist := strings.TrimSpace(r.Form.Get("artist"))

	if name == "" || artist == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("name or artist n=missing"))
		return
	}

	album := &Album{
		Name:   name,
		Artist: artist,
	}

	if id == nil {
		err = s.db.CreateAlbum(r.Context(), album)
	} else {
		album.ID = *id
		err = s.db.UpdateAlbum(r.Context(), album)
	}
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/albums#"+strconv.FormatUint(album.ID, 10), http.StatusSeeOther)
}

func (s *server) getDeleteAlbum(w http.ResponseWriter, r *http.Request) {
	id, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	album, err := s.db.GetAlbum(r.Context(), id)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "album-delete.html", map[string]interface{}{
		"Title": "Delete Album",
		"Album": album,
	})
}

func (s *server) postDeleteAlbum(w http.ResponseWriter, r *http.Request) {
	id, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	err = s.db.DeleteAlbum(r.Context(), id)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/albums", http.StatusSeeOther)
}

func (s *server) getTags(w http.ResponseWriter, r *http.Request) {
	tags, err := s.db.GetTags(r.Context())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "tags.html", map[string]interface{}{
		"Title": "Tags",
		"Tags":  tags,
	})
}

func (s *server) getDeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID := chi.URLParam(r, "tag-id")
	if tagID == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("tag id is missing"))
		return
	}

	tag, err := s.db.GetTag(r.Context(), tagID)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "tag-delete.html", map[string]interface{}{
		"Title": "Delete Tag",
		"Tag":   tag,
	})
}

func (s *server) postDeleteTag(w http.ResponseWriter, r *http.Request) {
	tagID := chi.URLParam(r, "tag-id")
	if tagID == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("tag id is missing"))
		return
	}

	err := s.db.DeleteTag(r.Context(), tagID)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/tags", http.StatusSeeOther)
}

func (s *server) getTagConnect(w http.ResponseWriter, r *http.Request) {
	albumID, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	tag := chi.URLParam(r, "tag-id")
	if tag == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("tag id is missing"))
		return
	}

	album, err := s.db.GetAlbum(r.Context(), albumID)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "tag-connect.html", map[string]interface{}{
		"Title": "Connect Tag",
		"Tag":   tag,
		"Album": album,
	})
}

func (s *server) postTagConnect(w http.ResponseWriter, r *http.Request) {
	albumID, err := extractAlbumID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	tag := chi.URLParam(r, "tag-id")
	if tag == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("tag id is missing"))
		return
	}

	defer func() {
		s.tagMu.Lock()
		if tag == s.tag {
			s.tag = ""
		}
		s.tagMu.Unlock()
	}()

	album, err := s.db.GetAlbum(r.Context(), albumID)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	err = s.db.CreateTag(r.Context(), tag, album)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/tags#"+tag, http.StatusSeeOther)
}

func (s *server) getLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := s.db.GetLogs(r.Context())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "logs.html", map[string]interface{}{
		"Title": "Logs",
		"Logs":  logs,
	})
}

func (s *server) getDeleteLog(w http.ResponseWriter, r *http.Request) {
	ts := chi.URLParam(r, "ts")
	if ts == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("timestamp missing"))
		return
	}

	log, err := s.db.GetLog(r.Context(), ts)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	s.renderTemplate(w, http.StatusOK, "log-delete.html", map[string]interface{}{
		"Title": "Delete Log Entry",
		"Log":   log,
	})
}

func (s *server) postDeleteLog(w http.ResponseWriter, r *http.Request) {
	ts := chi.URLParam(r, "ts")
	if ts == "" {
		s.renderError(w, http.StatusBadRequest, errors.New("timestamp missing"))
		return
	}

	err := s.db.DeleteLog(r.Context(), ts)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/logs", http.StatusSeeOther)
}

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

func extractAlbumID(r *http.Request) (uint64, error) {
	idStr := chi.URLParam(r, "album-id")
	return strconv.ParseUint(idStr, 10, 64)
}

func (s *server) sendToTelegram(msg string) {
	for _, chatID := range s.tgChatIDs {
		data := url.Values{}
		data.Set("chat_id", chatID)
		data.Set("disable_web_page_preview", "true")
		data.Set("text", msg)

		u, err := url.Parse(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", s.tgToken))
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
