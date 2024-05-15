package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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
