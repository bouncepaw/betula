# Federation capabilities in Betula

**This document is outdated, sorry.**

Betula uses a homebrew mixture of ActivityPub and whatnot. Sometimes your system might work with Betula out of the box. Usually not. Most importantly, Betula does not implement HTTP signatures, which is a deal breaker for most other implementations. We'll have them later.

This document describes all outgoing and incoming Activities. Supporting them ensures your compatibility. Betula will also try to be more standards-compliant later, but it's not a priority for now.

## User identification
Every Betula is a single-user installation. Thus, Betula's URL is also its admin's URL.

The inbox is found at `/inbox`. Unknown Activities are dropped.

## Verification
HTTP Signatures implementation taken from Honk.

## Voting system
Betula only supports likes. No dislikes or emoji reactions.

## Repost notification
Public reposts are reported to authors of original posts.

A notification like this is made when Alice reposts Bob's post 42, and her repost gets number 84:

```json
{
    "@context": "https://www.w3.org/ns/activitystreams",
    "type": "Announce",
    "actor": {
        "type": "Person",
        "id": "https://links.alice",
        "inbox": "https://links.alice/inbox",
        "name": "Alice",
        "preferredUsername": "alice"
    },
    "id": "https://links.alice/84",
    "object": "https://links.bob/42"
}
```

You are to verify the repost yourself. We use microformats.

## Repost cancellation
If a repost is turned into a regular post or deleted, you will get a notification like this:

```json
{
    "@context": "https://www.w3.org/ns/activitystreams",
    "type": "Undo",
    "actor": {
        "type": "Person",
        "id": "https://links.alice",
        "inbox": "https://links.alice/inbox",
        "name": "Alice",
        "preferredUsername": "alice"
    },
    "object": {
        "type": "Announce",
        "id": "https://links.alice/84",
        "actor": "https://links.alice",
        "object": "https://links.bob/42"
    }
}
```

You are to verify the lack of repost yourself. We use microformats.
