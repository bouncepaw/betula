package activities

import (
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/stricks"
)

func NewAnnounce(originalURL string, repostURL string) ([]byte, error) {
	activity := map[string]any{
		"@context": atContext,
		"type":     "Announce",
		"actor":    betulaActor,
		"id":       repostURL,
		"object":   originalURL,
	}
	return json.Marshal(activity)
}

type AnnounceReport struct {
	ReposterUsername string
	RepostPage       string // page where the repost is
	OriginalPage     string // page that was reposted
}

func mustHaveSuchField[T any](activity map[string]any, field string, errOnLack error, lambdaOnPresence func(T)) error {
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

func guessAnnounce(activity map[string]any) (reportMaybe any, err error) {
	var (
		actorMap map[string]any
		report   AnnounceReport
	)

	if err := mustHaveSuchField(
		activity, "actor", ErrNoActor,
		func(v map[string]any) {
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
