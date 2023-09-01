# Federation capabilities in Betula

Betula uses a homebrew mixture of ActivityPub and whatnot. Sometimes your system might work with Betula out of the box. Sometimes not.

This document describes all outgoing and incoming Activities. Supporting them ensures your compatibility. Betula will also try to be more standards-compliant later, but it's not a priority for now.

## Federation-connected instances
A Betula instance is considered _federation-connected_ if it:

* Has a domain set up.
* Does not have federation turned off in the settings.

## User identification
Every Betula is a single-user installation. Thus, Betula's URL is also its admin's URL.

The inbox is found at `/inbox`. Unknown Activities are dropped.

## Verification
Betula does not yet implement HTTP signatures and relies on manual resource fetching instead. HTTP signatures support may be implemented in the future.

## Repost notification
Public federation-connected reposts are reported to authors of original posts.

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
