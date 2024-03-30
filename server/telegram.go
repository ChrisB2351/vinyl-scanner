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

func (s *server) handleSetName(ctx tele.Context) error {
	args := ctx.Args()
	if len(args) < 2 {
		return ctx.Send("Wrong arguments: /set_name <id> <name>")
	}

	id := args[0]
	name := strings.Join(args[1:], " ")

	return s.updateAlbums(func(albums Albums) (Albums, error) {
		album := Album{}
		if a, ok := albums[id]; ok {
			album = a
		}
		album.Name = name
		albums[id] = album
		return albums, nil
	})
}

func (s *server) handleSetArtist(ctx tele.Context) error {
	args := ctx.Args()
	if len(args) < 2 {
		return ctx.Send("Wrong arguments: /set_artist <id> <name>")
	}

	id := args[0]
	artist := strings.Join(args[1:], " ")

	return s.updateAlbums(func(albums Albums) (Albums, error) {
		album := Album{}
		if a, ok := albums[id]; ok {
			album = a
		}
		album.Artist = artist
		albums[id] = album
		return albums, nil
	})
}

func (s *server) handleUpdateID(ctx tele.Context) error {
	args := ctx.Args()
	if len(args) != 2 {
		return ctx.Send("Wrong arguments: /update_id <old_id> <new_id>")
	}

	oldID := args[0]
	newID := args[1]

	return s.updateAlbums(func(albums Albums) (Albums, error) {
		album, ok := albums[oldID]
		if !ok {
			return albums, nil
		}

		album.OldIDs = append(album.OldIDs, oldID)
		delete(albums, oldID)
		albums[newID] = album
		return albums, nil
	})
}

func (s *server) handleAlbums(ctx tele.Context) error {
	albums, err := s.loadAlbums()
	if err != nil {
		return err
	}

	str := ""
	for _, a := range albums {
		str += a.String() + "\n"
	}

	return ctx.Send(str)
}
