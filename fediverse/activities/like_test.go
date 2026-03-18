// SPDX-FileCopyrightText: 2026 Danila Gorelko
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"fmt"
	"io"
	"testing"

	"github.com/nalgeon/be"
)

func readParseGuess(t *testing.T, fileName string) (any, []byte) {
	f, err := fs.Open(fmt.Sprintf("testdata/%s", fileName))
	be.Err(t, err, nil)

	bs, err := io.ReadAll(f)
	be.Err(t, err, nil)

	report, err := Guess(bs)
	be.Err(t, err, nil)

	return report, bs
}

func TestGuessLike(t *testing.T) {
	t.Parallel()

	t.Run("GoToSocial", func(t *testing.T) {
		t.Parallel()

		report, bs := readParseGuess(t, "Like from GTS.json")
		want := LikeReport{
			ID:       "https://social.agor.ai/users/fish/liked/01KE2Z6HTKFHN4Z5S27JNTJNSG",
			ActorID:  "https://social.agor.ai/users/fish",
			ObjectID: "https://activitypub.academy/users/difricus_koloddath/statuses/115833494550206707",
			Activity: bs,
		}
		be.Equal(t, report, any(want))
	})

	t.Run("Mastodon", func(t *testing.T) {
		t.Parallel()

		report, bs := readParseGuess(t, "Like from Mastodon.json")
		want := LikeReport{
			ID:       "https://activitypub.academy/users/difricus_koloddath#likes/1374",
			ActorID:  "https://activitypub.academy/users/difricus_koloddath",
			ObjectID: "https://alice.bouncepaw.com/8",
			Activity: bs,
		}
		be.Equal(t, report, any(want))
	})
}

func TestGuessUndoLike(t *testing.T) {
	t.Parallel()

	t.Run("GoToSocial", func(t *testing.T) {
		t.Parallel()

		report, bs := readParseGuess(t, "Undo{Like} from GTS.json")
		want := UndoLikeReport{
			ID: "https://social.agor.ai/01PHQBF5739JVDHNJTE14SSAB8",
			Object: LikeReport{
				ID:       "https://social.agor.ai/users/fish/liked/01KE2Z6HTKFHN4Z5S27JNTJNSG",
				ActorID:  "https://social.agor.ai/users/fish",
				ObjectID: "https://activitypub.academy/users/difricus_koloddath/statuses/115833494550206707",
			},
			Activity: bs}
		be.Equal(t, report, any(want))
	})

	t.Run("Mastodon", func(t *testing.T) {
		t.Parallel()

		report, bs := readParseGuess(t, "Undo{Like} from Mastodon.json")
		want := UndoLikeReport{
			ID: "https://activitypub.academy/users/difricus_koloddath#likes/1374/undo",
			Object: LikeReport{
				ID:       "https://activitypub.academy/users/difricus_koloddath#likes/1374",
				ActorID:  "https://activitypub.academy/users/difricus_koloddath",
				ObjectID: "https://alice.bouncepaw.com/8",
			},
			Activity: bs}
		be.Equal(t, report, any(want))
	})
}
