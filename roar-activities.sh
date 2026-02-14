#!/bin/sh

# SPDX-FileCopyrightText: 2023 Timur Ismagilov <https://bouncepaw.com>
#
# SPDX-License-Identifier: AGPL-3.0-only

SallyCreated=$(
cat << 'EOT'
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally created a note",
  "type": "Create",
  "actor": {
    "type": "Person",
    "name": "Sally"
  },
  "object": {
    "type": "Note",
    "name": "A Simple Note",
    "content": "This is a simple note"
  }
}
EOT
)

SallyAnnouncedArrival=$(
cat << 'EOT'
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "summary": "Sally announced that she had arrived at work",
  "type": "Announce",
  "actor": {
    "type": "Person",
    "id": "http://sally.example.org",
    "name": "Sally"
  },
  "object": {
    "type": "Arrive",
    "actor": "http://sally.example.org",
    "location": {
      "type": "Place",
      "name": "Work"
    }
  }
}
EOT
)

MeaningfulAnnounce=$(
cat << 'EOT'
{
  "@context": "https://www.w3.org/ns/activitystreams",
  "type": "Announce",
  "object": "https://bouncepaw.com/url of your post",
  "url": "https://parasocial.mycorrhiza.wiki/23"
}
EOT
)

Inbox='http://localhost:1738/inbox'


Post() {
  Data=$1
  shift
  curl -isS "$@" "$Inbox" -X POST --header "Content-Type: application/activity+json" --data "$Data"
}

Post "Junk"
Post "$SallyCreated"
Post "$SallyAnnouncedArrival"
Post "$MeaningfulAnnounce"