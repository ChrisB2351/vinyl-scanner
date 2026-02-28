package main

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
)

type config struct {
	tgToken   string
	tgChatIDs []int64

	apiToken string
	dataDir  string
	baseURL  string

	jwtSecret string
	username  string
	password  string
}

type server struct {
	mux     *chi.Mux
	db      *database
	baseURL string

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

	db, err := newDatabase(filepath.Join(cfg.dataDir, "data.sqlite3"))
	if err != nil {
		return nil, err
	}

	pwd, err := base64.StdEncoding.DecodeString(cfg.password)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 bcrypt hashed password: %w", err)
	}

	s := &server{
		mux:      chi.NewRouter(),
		db:       db,
		baseURL:  cfg.baseURL,
		tgToken:  cfg.tgToken,
		jwtAuth:  jwtauth.New("HS256", []byte(base64.StdEncoding.EncodeToString([]byte(cfg.jwtSecret))), nil),
		username: cfg.username,
		password: string(pwd),
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
		r.Get("/albums/{id}", s.getAlbum)
		r.Post("/albums/{id}", s.postAlbum)
		r.Get("/albums/{id}/delete", s.getDeleteAlbum)
		r.Post("/albums/{id}/delete", s.postDeleteAlbum)

		r.Get("/logs", s.getLogs)
		r.Get("/logs/{id}/delete", s.getDeleteLog)
		r.Post("/logs/{id}/delete", s.postDeleteLog)
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

type pagination struct {
	Page       int
	TotalPages int
	PrevURL    string
	NextURL    string
}

func parseOrder(r *http.Request, defaultOrder string) string {
	order := r.URL.Query().Get("order")
	if order != "asc" && order != "desc" {
		return defaultOrder
	}
	return order
}

func newPagination(r *http.Request, total int64, pageURL func(int) string) pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	totalPages := max(int((total+pageSize-1)/pageSize), 1)
	if page > totalPages {
		page = totalPages
	}

	p := pagination{Page: page, TotalPages: totalPages}
	if page > 1 {
		p.PrevURL = pageURL(page - 1)
	}

	if page < totalPages {
		p.NextURL = pageURL(page + 1)
	}

	return p
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
