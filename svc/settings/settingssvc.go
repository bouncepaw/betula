// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

// Package settingssvc handles the settings service, which is just logging setup for now.
package settingssvc

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"git.sr.ht/~bouncepaw/betula/pkg/ecs"
	settingsports "git.sr.ht/~bouncepaw/betula/ports/settings"
)

type Service struct {
	repo     settingsports.Repository
	version  string
	domainFn func() string
}

var _ settingsports.Service = (*Service)(nil)

func New(repo settingsports.Repository, version string, domainFn func() string) *Service {
	return &Service{
		repo:     repo,
		version:  version,
		domainFn: domainFn,
	}
}

func (svc *Service) GetLoggingSettings(ctx context.Context) (settingsports.LoggingSettings, error) {
	return svc.repo.GetLoggingSettings(ctx)
}

func (svc *Service) SaveLoggingSettings(ctx context.Context, ls settingsports.LoggingSettings) error {
	if err := svc.repo.SetLoggingSettings(ctx, ls); err != nil {
		return fmt.Errorf("failed to save logging settings: %w", err)
	}
	logger, err := svc.newLogger(ctx)
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	slog.SetDefault(logger.With("app", "betula", "domain", svc.domainFn()))
	slog.Info("Hello Betula!", "version", svc.version)
	return nil
}

func (svc *Service) newLogger(ctx context.Context) (*slog.Logger, error) {
	ls, err := svc.repo.GetLoggingSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get logging settings: %w", err)
	}

	method := settingsports.LoggingMethodDefault
	if ls.Method != nil {
		method = *ls.Method
	}

	switch method {
	case settingsports.LoggingMethodDefault:
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})), nil

	case settingsports.LoggingMethodECSNoAuth:
		if ls.URL == nil {
			return nil, fmt.Errorf("%q requires url", method)
		}
		return ecs.NewNoAuthLogger(*ls.URL), nil

	case settingsports.LoggingMethodECSBasicAuth:
		if ls.URL == nil || ls.Username == nil || ls.Token == nil {
			return nil, fmt.Errorf("%q requires url, username, and token", method)
		}
		return ecs.NewBasicAuthLogger(*ls.URL, *ls.Username, *ls.Token), nil

	case settingsports.LoggingMethodECSBearer:
		if ls.URL == nil || ls.Token == nil {
			return nil, fmt.Errorf("%q requires url and token", method)
		}
		return ecs.NewBearerLogger(*ls.URL, *ls.Token), nil

	default:
		return nil, fmt.Errorf("unknown logging method: %q", method)
	}
}
