// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package ecs provides slog loggers that emit Elastic Common Schema (ECS)
// formatted JSON to a remote HTTP ingestion endpoint.
package ecs

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
)

func NewNoAuthLogger(url string) *slog.Logger {
	return slog.New(newHandler(io.MultiWriter(os.Stdout, newHTTPWriter(url, ""))))
}

func NewBasicAuthLogger(url, username, password string) *slog.Logger {
	creds := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	return slog.New(newHandler(io.MultiWriter(os.Stdout, newHTTPWriter(url, "Basic "+creds))))
}

func NewBearerLogger(url, token string) *slog.Logger {
	return slog.New(newHandler(io.MultiWriter(os.Stdout, newHTTPWriter(url, "Bearer "+token))))
}

func newHandler(w io.Writer) slog.Handler {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if len(groups) > 0 {
				return a
			}
			switch a.Key {
			case slog.TimeKey:
				return slog.Attr{Key: "@timestamp", Value: a.Value}
			case slog.LevelKey:
				level, _ := a.Value.Any().(slog.Level)
				return slog.String("log.level", strings.ToLower(level.String()))
			case slog.MessageKey:
				return slog.Attr{Key: "message", Value: a.Value}
			}
			return a
		},
	})
	return h.WithAttrs([]slog.Attr{
		slog.Group("ecs", slog.String("version", "9.3.0")),
	})
}

type httpWriter struct {
	mu     sync.Mutex
	url    string
	auth   string
	client http.Client
}

func newHTTPWriter(url, auth string) *httpWriter {
	return &httpWriter{url: url, auth: auth}
}

func (w *httpWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	req, err := http.NewRequest(http.MethodPost, w.url, bytes.NewReader(p))
	if err != nil {
		return 0, fmt.Errorf("ecs: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if w.auth != "" {
		req.Header.Set("Authorization", w.auth)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("ecs: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("ecs: server returned %d", resp.StatusCode)
	}

	return len(p), nil
}
