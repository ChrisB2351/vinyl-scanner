package main

import (
	"embed"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"text/template"
)

var (
	//go:embed templates/*.html
	templatesFS embed.FS
	templates   = template.Must(template.ParseFS(templatesFS, "templates/*.html"))

	//go:embed assets/*
	assetsFS embed.FS
)

func (s *server) renderTemplate(w http.ResponseWriter, code int, template string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	err := templates.ExecuteTemplate(w, template, data)
	if err != nil {
		slog.Error("serving html", "error.html", err)
	}
}

func (s *server) renderError(w http.ResponseWriter, code int, reqErr error) {
	if errors.Is(reqErr, errNoItem) {
		code = http.StatusNotFound
	}

	data := map[string]interface{}{
		"Title":  fmt.Sprintf("%d %s", code, http.StatusText(code)),
		"Status": code,
	}

	if reqErr != nil {
		slog.Error("serving html", "error", reqErr)
		data["Message"] = reqErr.Error()
	}

	s.renderTemplate(w, code, "error.html", data)
}
