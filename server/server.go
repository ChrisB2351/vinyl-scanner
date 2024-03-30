package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	tele "gopkg.in/telebot.v3"
	"gopkg.in/telebot.v3/middleware"
)

type Albums map[string]Album

type Album struct {
	Name   string `json:"name,omitempty"`
	Artist string `json:"artist,omitempty"`
}

type server struct {
	bot      *tele.Bot
	chatIDs  []int64
	dataDir  string
	logMu    sync.Mutex
	albumsMu sync.Mutex
	lastMu   sync.Mutex
	lastID   string
}

func newServer(tgToken string, tgChatIDs []int64, dataDir string) (*server, error) {
	err := os.MkdirAll(dataDir, 0666)
	if err != nil {
		return nil, err
	}

	pref := tele.Settings{
		Token:  tgToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	s := &server{
		chatIDs: tgChatIDs,
		bot:     bot,
		dataDir: dataDir,
	}

	bot.Use(middleware.Whitelist(tgChatIDs...))
	bot.Handle("/name", s.handleName)
	bot.Handle("/artist", s.handleArtist)
	bot.Handle("/clear", s.handleClear)

	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id := string(body)
	slog.Info("received new id", "id", id)
	w.WriteHeader(http.StatusOK)

	go s.handleID(id)
}

func (s *server) handleID(id string) {
	if s.isLastID(id) {
		return
	}

	go s.logID(id)

	albums, err := s.loadAlbums()
	if err != nil {
		s.sendMessage(fmt.Sprintf("could not load albums: %s", err))
		return
	}

	var msg string
	album, ok := albums[id]
	if !ok {
		msg = fmt.Sprintf("Please associate a name with this vinyl using \"/name %s <name>\"", id)
	} else {
		msg = fmt.Sprintf("Scanned vinyl \"%s\"", album.Name)
	}

	s.sendMessage(msg)
}

func (s *server) isLastID(id string) bool {
	s.lastMu.Lock()
	defer s.lastMu.Unlock()

	if s.lastID == id {
		return true
	}

	s.lastID = id
	return false
}

func (s *server) logID(id string) {
	entry := fmt.Sprintf("%s,%s\n", id, time.Now().UTC().Format(time.RFC3339))

	s.logMu.Lock()
	defer s.logMu.Unlock()

	fd, err := os.OpenFile(filepath.Join(s.dataDir, "listens.log"), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Printf("error while opening listens.log: %s", err)
		return
	}

	_, err = fd.Write([]byte(entry))
	if err != nil {
		log.Printf("error while writing to listens.log: %s", err)
		return
	}
}

func (s *server) sendMessage(msg string) {
	for _, id := range s.chatIDs {
		_, err := s.bot.Send(&tele.Chat{ID: id}, msg)
		if err != nil {
			log.Printf("error while sending to telegram: %s", err)
		}
	}
}

func (s *server) loadAlbums() (Albums, error) {
	s.albumsMu.Lock()
	defer s.albumsMu.Unlock()

	filename := filepath.Join(s.dataDir, "albums.json")

	raw, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return Albums{}, nil
	} else if err != nil {
		return nil, err
	}

	var albums Albums
	err = json.Unmarshal(raw, &albums)
	if err != nil {
		return nil, err
	}

	return albums, nil
}

func (s *server) writeAlbums(albums Albums) error {
	raw, err := json.Marshal(albums)
	if err != nil {
		return err
	}

	s.albumsMu.Lock()
	defer s.albumsMu.Unlock()

	filename := filepath.Join(s.dataDir, "albums.json")
	return os.WriteFile(filename, raw, 0666)
}
