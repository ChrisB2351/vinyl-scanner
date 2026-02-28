package main

import (
	"net/http"
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
	id, err := extractID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	log, err := s.db.GetLog(r.Context(), id)
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
	id, err := extractID(r)
	if err != nil {
		s.renderError(w, http.StatusBadRequest, err)
		return
	}

	err = s.db.DeleteLog(r.Context(), id)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, "/logs", http.StatusSeeOther)
}
