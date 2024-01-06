package fediverse

import (
	"encoding/json"
	"fmt"
	"git.sr.ht/~bouncepaw/betula/fediverse/signing"
	"git.sr.ht/~bouncepaw/betula/types"
	"io"
	"log"
	"net/http"
)

// RequestActor fetches the actor activity on the specified address.
func RequestActor(actorID string) (actor *types.Actor, err error) {
	cope := func(err error) error {
		return fmt.Errorf("requesting actor: %w", err)
	}

	req, err := http.NewRequest("GET", actorID, nil)
	if err != nil {
		return nil, cope(err)
	}
	req.Header.Set("Accept", types.ActivityType)
	signing.SignRequest(req, nil)

	resp, err := client.Do(req)
	if err != nil {
		return nil, cope(err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("requesting actor: status not 200")
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, cope(err)
	}

	var a types.Actor
	if err = json.Unmarshal(data, &a); err != nil {
		return nil, cope(err)
	}

	return &a, nil
}

func RequestActorInbox(actorID string) string {
	actor, err := RequestActor(actorID)
	if err != nil {
		log.Printf("When requesting actor %s inbox: %s\n", actorID, err)
		return ""
	}
	return actor.Inbox
}
