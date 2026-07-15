// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package assembly

import (
	"encoding/base64"
	"encoding/json"
	"path"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
)

func (asm *Assembler) NewLike(likedObjectID, recipientID string) (json.RawMessage, error) {
	encID := base64.URLEncoding.EncodeToString([]byte(likedObjectID))
	activity := Dict{
		"@context": atContext,
		"id":       path.Join(asm.siteURLFn(), "likes", encID),
		"type":     "Like",
		"actor":    asm.actor(),
		"object":   likedObjectID,
		"to":       recipientID,
	}
	return json.Marshal(activity)
}

func (asm *Assembler) NewUndoLike(likedObjectID, recipientID string) (json.RawMessage, error) {
	encID := base64.URLEncoding.EncodeToString([]byte(likedObjectID))
	activity := Dict{
		"@context": atContext,
		"id":       path.Join(asm.siteURLFn(), "temp", bxstr.RandomWhatever()),
		"type":     "Undo",
		"actor":    asm.actor(),
		"to":       recipientID,
		"object": Dict{
			"actor":  asm.actor(),
			"id":     path.Join(asm.siteURLFn(), "likes", encID),
			"object": likedObjectID,
			"to":     recipientID,
			"type":   "Like",
		},
	}
	return json.Marshal(activity)
}
