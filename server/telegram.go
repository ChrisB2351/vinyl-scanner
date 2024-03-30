package main

import (
	"strings"

	tele "gopkg.in/telebot.v3"
)

func (s *server) handleClear(ctx tele.Context) error {
	s.lastMu.Lock()
	defer s.lastMu.Unlock()

	s.lastID = ""
	return nil
}

func (s *server) handleName(ctx tele.Context) error {
	args := ctx.Args()
	if len(args) < 2 {
		return ctx.Send("Not enough arguments: /name <id> <name>")
	}

	id := args[0]
	name := strings.Join(args[1:], " ")

	albums, err := s.loadAlbums()
	if err != nil {
		return err
	}

	album := Album{}
	if a, ok := albums[id]; ok {
		album = a
	}
	album.Name = name
	albums[id] = album

	return s.writeAlbums(albums)
}

func (s *server) handleArtist(ctx tele.Context) error {
	args := ctx.Args()
	if len(args) < 2 {
		return ctx.Send("Not enough arguments: /artist <id> <name>")
	}

	id := args[0]
	artist := strings.Join(args[1:], " ")

	albums, err := s.loadAlbums()
	if err != nil {
		return err
	}

	album := Album{}
	if a, ok := albums[id]; ok {
		album = a
	}
	album.Artist = artist
	albums[id] = album

	return s.writeAlbums(albums)
}
