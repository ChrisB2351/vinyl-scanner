package main

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

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
