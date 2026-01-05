// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package activities

import (
	"fmt"
	"io"
	"reflect"
	"testing"
)

func scream(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func readParseGuess(t *testing.T, fileName string) (any, []byte) {
	f, err := fs.Open(fmt.Sprintf("testdata/%s", fileName))
	scream(t, err)

	bs, err := io.ReadAll(f)
	scream(t, err)

	report, err := Guess(bs)
	scream(t, err)

	return report, bs
}

func deepEqual(t *testing.T, got any, want any) {
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v\nwant %v", got, want)
	}
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
		deepEqual(t, report, want)
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
		deepEqual(t, report, want)
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
		deepEqual(t, report, want)
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
		deepEqual(t, report, want)
	})
}
