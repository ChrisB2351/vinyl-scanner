package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

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

func extractAlbumID(r *http.Request) (uint64, error) {
	idStr := chi.URLParam(r, "album-id")
	return strconv.ParseUint(idStr, 10, 64)
}
