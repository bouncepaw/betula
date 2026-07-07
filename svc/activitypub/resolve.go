// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package apsvc

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	apports "git.sr.ht/~bouncepaw/betula/ports/activitypub"
	webfingerports "git.sr.ht/~bouncepaw/betula/ports/webfinger"
	"git.sr.ht/~bouncepaw/betula/types"
)

func (svc *FollowService) resolveActor2Methods(ctx context.Context, input string) (apports.Actor, error) {
	input = strings.TrimSpace(input)
	opts := apports.GetActorsOpts{GetPublicKey: true}

	if acct, ok := webFingerAcct(input); ok {
		id, err := svc.webfinger.DereferenceAcct(acct)
		if err != nil {
			return nil, fmt.Errorf("failed to webfinger %s: %w", input, err)
		}
		if id == "" {
			return nil, fmt.Errorf("no ActivityPub actor found for %s", input)
		}
		return svc.activityPub.ActorByID(ctx, id, opts)
	}

	return svc.activityPub.ActorByID(ctx, input, opts)
}

func (svc *FollowService) resolveActor3Methods(ctx context.Context, input string) (apports.Actor, error) {
	input = strings.TrimSpace(input)
	actor, err := svc.resolveActor2Methods(ctx, input)
	if err == nil {
		return actor, nil
	}

	slog.Info("Direct actor dereference failed; trying HTML alternate links",
		"input", input, "err", err)

	alternates, htmlErr := svc.www.RelAlternates(input)
	if htmlErr != nil {
		return nil, fmt.Errorf("dereference %s failed (%w) and fetching its HTML failed: %w",
			input, err, htmlErr)
	}
	for _, alt := range alternates {
		if !types.ContainsActivityType(alt.Type) || alt.Href == "" {
			continue
		}
		href := alt.ResolveHref(input)
		slog.Info("Found ActivityPub alternate link", "input", input, "href", href)
		return svc.activityPub.ActorByID(ctx, href, apports.GetActorsOpts{GetPublicKey: true})
	}
	return nil, fmt.Errorf("no ActivityPub actor link found at %s: %w", input, err)
}

func webFingerAcct(input string) (acct webfingerports.Acct, ok bool) {
	input = strings.TrimPrefix(input, "@")
	user, host, ok := strings.Cut(input, "@")
	if !ok || user == "" || host == "" || strings.Contains(user, "/") || strings.Contains(host, "/") {
		return webfingerports.Acct{}, false
	}
	return webfingerports.Acct{User: user, Host: host}, true
}
