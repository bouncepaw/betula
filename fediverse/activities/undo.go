package activities

import (
	"encoding/json"
	"git.sr.ht/~bouncepaw/betula/settings"
)

type UndoAnnounceReport struct {
	AnnounceReport
}

type UndoFollowReport struct {
	FollowReport
}

func newUndo(objectId string, object dict) ([]byte, error) {
	object["id"] = objectId
	return json.Marshal(dict{
		"@context": atContext,
		"type":     "Undo",
		"actor":    betulaActor,
		"id":       objectId + "?undo",
		"object":   object,
	})
}

func NewUndoAnnounce(repostURL string, originalPostURL string) ([]byte, error) {
	return newUndo(
		repostURL,
		dict{
			"type":   "Announce",
			"actor":  settings.SiteURL(),
			"object": originalPostURL,
		})
}

func guessUndo(activity dict) (reportMaybe any, err error) {
	var (
		report    UndoAnnounceReport
		objectMap dict
	)

	if err := mustHaveSuchField(
		activity, "object", ErrNoObject,
		func(v map[string]any) {
			objectMap = v
		},
	); err != nil {
		return nil, err
	}

	switch objectMap["type"] {
	case "Announce":
		switch repost := objectMap["id"].(type) {
		case string:
			report.RepostPage = repost
		}
		switch original := objectMap["object"].(type) {
		case string:
			report.OriginalPage = original
		}
		switch actor := objectMap["actor"].(type) {
		case dict:
			switch username := actor["preferredUsername"].(type) {
			case string:
				report.ReposterUsername = username
			}
		}
		return report, nil
	case "Follow":
		if objectMap == nil {
			return nil, ErrNoObject
		}
		followReport, err := guessFollow(objectMap)
		if err != nil {
			return nil, err
		}
		return UndoFollowReport{followReport.(FollowReport)}, nil
	default:
		return nil, ErrUnknownType
	}
}
