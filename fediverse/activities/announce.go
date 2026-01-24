// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"encoding/json"

	"git.sr.ht/~bouncepaw/betula/pkg/stricks"
	"git.sr.ht/~bouncepaw/betula/settings"
)

func NewAnnounce(originalURL string, repostURL string) ([]byte, error) {
	activity := map[string]any{
		"@context": atContext,
		"type":     "Announce",
		"actor": map[string]string{
			"id":                betulaActor,
			"preferredUsername": settings.AdminUsername(),
		},
		"id":     repostURL,
		"object": originalURL,
	}
	return json.Marshal(activity)
}

type AnnounceReport struct {
	ReposterUsername string
	RepostPage       string // page where the repost is
	OriginalPage     string // page that was reposted
}

func mustHaveSuchField[T any](activity Dict, field string, errOnLack error, lambdaOnPresence func(T)) error {
	if val, ok := activity[field]; !ok {
		return errOnLack
	} else {
		switch v := val.(type) {
		case T:
			lambdaOnPresence(v)
			return nil
		default:
			return errOnLack
		}
	}
}

func guessAnnounce(activity Dict) (reportMaybe any, err error) {
	var (
		actorMap Dict
		report   AnnounceReport
	)

	if err := mustHaveSuchField(
		activity, "actor", ErrNoActor,
		func(v Dict) {
			actorMap = v
		},
	); err != nil {
		return nil, err
	}

	if err := mustHaveSuchField(
		actorMap, "preferredUsername", ErrNoActorUsername,
		func(v string) {
			report.ReposterUsername = v
		},
	); err != nil {
		return nil, err
	}

	if err := mustHaveSuchField(
		activity, "object", ErrNoObject,
		func(v string) {
			report.OriginalPage = v
		},
	); err != nil {
		return nil, err
	}

	if err := mustHaveSuchField(
		activity, "id", ErrNoId,
		func(v string) {
			report.RepostPage = v
		},
	); err != nil {
		return nil, err
	}

	if !stricks.ValidURL(report.OriginalPage) {
		return nil, ErrNoObject
	}

	if !stricks.ValidURL(report.RepostPage) {
		return nil, ErrNoId
	}

	return report, nil
}
