# Betula's Approach to Federated Search
There is already an ongoing federated search initiative present, called [Fediverse Discovery Providers](https://www.fediscovery.org), or Fediscovery for short. It's not ready yet, no specs are found, and it's probably going to be very general. I do not particularly want to wait for them, hence I'm rolling out my own ad-hoc protocol suited for federated search of bookmarks.

I tried to design it universal enough for all fellow Fediverse bookmarking services, but of course it's going to be very biased. After all, I am the one who develops Betula, so of course I take it into consideration first.

## Search results
Alice wants to ask Bob for search results. She makes a POST request to `https://BOB/.well-known/betula-federated-search` similar to the following JSON signed with an HTTP signature.

```json
{
   "version": "v1",
   "query": "#solarpunk #software",
   "limit": 6,
   "offset": 0,
   "from": "https://ALICE/@alice",
   "to": "https://BOB/@bob"
}
```

The cursor might be `null`. Reverse-chronological order is assumed.

If the signature is fine, the status is 200 and Bob returns an object like this:

```json
{
	"moreAvailable": 0,
	"bookmarks": [
		{
			"@context": [
				"https://www.w3.org/ns/activitystreams",
				{
					"Hashtag": "https://www.w3.org/ns/activitystreams#Hashtag"
				}
			],
			"actor": "https://BOB/@bob",
			"attributedTo": "https://BOB/@bob",
			"content": "\u003ch3\u003e\u003ca href=\"https://www.datagubbe.se/adosmyst/\"'\u003e 403 Forbidden\n\u003c/a\u003e\u003c/h3\u003e\u003carticle class=\"mycomarkup-doc\"\u003e\u003cp\u003eCute.\n\u003c/p\u003e\u003c/article\u003e\u003cp\u003e\u003ca href=\"https://BOB/tag/retrocomputing\" class=\"mention hashtag\" rel=\"tag\"\u003e#\u003cspan\u003eretrocomputing\u003c/span\u003e\u003c/a\u003e\u003c/p\u003e",
			"id": "https://BOB/1083",
			"source": {
				"content": "Cute.",
				"mediaType": "text/mycomarkup"
			},
			"name": " 403 Forbidden\n",
			"attachment": [
				{
					"type": "Link",
					"href": "https://www.datagubbe.se/adosmyst/"
				}
			],
			"published": "2024-01-31T19:47:02Z",
			"tag": [
				{
					"href": "https://BOB/tag/retrocomputing",
					"name": "#retrocomputing",
					"type": "Hashtag"
				}
			],
			"to": [
				"https://www.w3.org/ns/activitystreams#Public",
				"https://BOB/followers"
			],
			"type": "Note"
		}
	]
}
```

The `moreAvailable` field tells how many more bookmarks can be requested by adjusting the query. The `items` is a list of `Note` objects. They are the same as if received over regular ActivityPub broadcasting, except there is no wrapping `Create` or `Update` activity. Alice might save these bookmarks to her database.

Possible errors are:

* `400 Bad Request`: couldn't parse the request or something.
* `401 Unauthorized`: something is wrong with the HTTP signature
* `403 Forbidden`: Bob doesn't want to share bookmarks with Alice this way (not mutuals, for example)
* `404 Not Found`: Bob doesn't have federated search enabled or at all
* `500 Internal Server Error`: something went wrong on Bob's side

I should have used an OpenAPI spec, haven't I?

## Considered alternatives
* Straight up parsing everything. Rude!
* Asking for results with a custom activity. That'd be harder and too formal.