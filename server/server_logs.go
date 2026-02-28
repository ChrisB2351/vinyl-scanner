package main

import (
	"fmt"
	"net/http"
)

func (s *server) getLogs(w http.ResponseWriter, r *http.Request) {
	order := parseOrder(r, "desc")

	total, err := s.db.CountLogs(r.Context())
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	p := newPagination(r, total, func(pg int) string {
		return fmt.Sprintf("/logs?order=%s&page=%d", order, pg)
	})

	logs, err := s.db.GetLogs(r.Context(), order, (p.Page-1)*pageSize, pageSize)
	if err != nil {
		s.renderError(w, http.StatusInternalServerError, err)
		return
	}

	toggleOrder := "asc"
	if order == "asc" {
		toggleOrder = "desc"
	}

	s.renderTemplate(w, http.StatusOK, "logs.html", map[string]interface{}{
		"Title":       "Logs",
		"Logs":        logs,
		"Total":       total,
		"Order":       order,
		"SortTimeURL": fmt.Sprintf("/logs?order=%s&page=1", toggleOrder),
		"Pagination":  p,
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
