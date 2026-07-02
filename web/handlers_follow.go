// SPDX-FileCopyrightText: 2025 Timur Ismagilov <https://bouncepaw.com>
// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"log/slog"
	"net/http"
	"net/url"

	"git.sr.ht/~bouncepaw/betula/pkg/bxstr"
	"git.sr.ht/~bouncepaw/betula/types"
)

type renderedActor struct {
	types.Actor
	// Redirected here after clicking the follow button.
	Next string
}

type dataActorList struct {
	*dataCommon

	Actors []renderedActor
}

func getFollowersWeb(w http.ResponseWriter, rq *http.Request) {
	actors, err := ctrl.RepoActor.GetFollowers(rq.Context())
	if err != nil {
		slog.Error("Failed to get followers", "err", err)
		handlerBadRequest(w, rq)
		return
	}

	renderedActors := make([]renderedActor, len(actors))
	for i, actor := range actors {
		renderedActors[i] = renderedActor{
			Actor: actor,
			Next:  "/followers#" + url.PathEscape(actor.Acct()),
		}
	}

	templateExec(w, rq, templateFollowers, dataActorList{
		dataCommon: emptyCommon(),
		Actors:     renderedActors,
	})
}

func getFollowingWeb(w http.ResponseWriter, rq *http.Request) {
	actors, err := ctrl.RepoActor.GetFollowing(rq.Context())
	if err != nil {
		slog.Error("Failed to get following", "err", err)
		handlerBadRequest(w, rq)
		return
	}

	renderedActors := make([]renderedActor, len(actors))
	for i, actor := range actors {
		renderedActors[i] = renderedActor{
			Actor: actor,
			Next:  "/following#" + url.PathEscape(actor.Acct()),
		}
	}

	templateExec(w, rq, templateFollowing, dataActorList{
		dataCommon: emptyCommon(),
		Actors:     renderedActors,
	})
}

// postUnfollow is similar to postFollow excepts it's unfollow.
func postUnfollow(w http.ResponseWriter, rq *http.Request) {
	var (
		nickname = rq.FormValue("account")
		next     = rq.FormValue("next")
	)
	slog.Info("Next", "next", next)

	if nickname == "" || next == "" {
		slog.Warn("/unfollow: required parameters were not passed")
		handlerNotFound(w, rq)
		return
	}

	err := ctrl.SvcFollow.Unfollow(rq.Context(), nickname)
	if err != nil {
		slog.Error("/unfollow: failed to unfollow", "nickname", nickname, "err", err)
		next = bxstr.ValidURLWithQuery(next, map[string]string{
			"unfollow-err": err.Error(),
		})
	} else {
		slog.Info("/unfollow: unfollowed successfully", "nickname", nickname)
		next = bxstr.ValidURLWithQuery(next, map[string]string{
			"unfollow-ok": "true",
		})
	}
	slog.Info("Next", "next", next)

	http.Redirect(w, rq, next, http.StatusSeeOther)
}

// postFollow follows the account specified and redirects to next with a notification.
// Both parameters are required.
//
//	/follow?account=@bouncepaw@links.bouncepaw.com&next=/@bouncepaw@links.bouncepaw.com
func postFollow(w http.ResponseWriter, rq *http.Request) {
	var (
		nickname = rq.FormValue("account")
		next     = rq.FormValue("next")
	)
	slog.Info("Next", "next", next)

	if nickname == "" || next == "" {
		slog.Warn("/follow: required parameters were not passed")
		handlerNotFound(w, rq)
		return
	}

	err := ctrl.SvcFollow.Follow(rq.Context(), nickname)
	if err != nil {
		slog.Error("/follow: failed to follow", "nickname", nickname, "err", err)
		next = bxstr.ValidURLWithQuery(next, map[string]string{
			"follow-err": err.Error(),
		})
	} else {
		slog.Info("/follow: followed successfully", "nickname", nickname)
		next = bxstr.ValidURLWithQuery(next, map[string]string{
			"follow-ok": "true",
		})
	}
	slog.Info("Next", "next", next)

	http.Redirect(w, rq, next, http.StatusSeeOther)
}
