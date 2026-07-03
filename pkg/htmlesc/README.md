# HTML Escaping
Inspired by Mastodon, but further adjust to be more permissive.

## Original Mastodon description
From https://docs.joinmastodon.org/spec/activitypub/#sanitization

Mastodon sanitizes incoming HTML in order to not break assumptions for API client developers. Supported elements will be kept as-is, and unsupported elements will be converted or removed. Supported attributes will be kept, and all other attributes will be stripped. The following elements and attributes are supported:

* `<p>`
* `<span>` (`class`)
* `<br>`
* `<a>` (`href`, `rel`, `class`)
* lists will be converted to `<p>`, and list items will be separated with `<br>`

Since Mastodon v4.2, the following elements and attributes are supported:

* `<p>`
* `<span>` (`class`)
* `<br>`
* `<a>` (`href`, `rel`, `class`)
* `<del>`
* `<pre>`
* `<code>`
* `<em>`
* `<strong>`
* `<b>`
* `<i>`
* `<u>`
* `<ul>`
* `<ol>` (`start`, `reversed`)
* `<li>` (`value`)
* `<blockquote>`
* headings will be converted to `<strong>` and then wrapped in `<p>`

The sanitizer will keep classes if they begin with microformats prefixes or are semantic classes:

* `h-*`
* `p-*`
* `u-*`
* `dt-*`
* `e-*`
* `mention`
* `hashtag`
* `ellipsis`
* `invisible`

Links will be kept if the protocol is supported, and converted to text otherwise. The following link protocols are supported:

* `http`
* `https`
* `dat`
* `dweb`
* `ipfs`
* `ipns`
* `ssb`
* `gopher`
* `xmpp`
* `magnet`
* `gemini`

## Logic for Betula
Differences highlighted.

Supported elements:

* `<p>`
* `<span>` (`class`)
* `<br>`
* `<a>` (`href`, `rel`, `class`)
* `<del>`
* `<pre>`
* `<code>`
* `<em>`
* `<strong>`
* `<b>`
* `<i>`
* `<u>`
* `<ul>`
* `<ol>` (`start`, `reversed`)
* `<li>` (`value`)
* `<blockquote>`
* headings will be converted to `<strong>` and then wrapped in `<p>`
* _add:_ `<mark>`
* _add:_ `<sub>`
* _add:_ `<sup>`
* _add:_ `<img>`

The sanitizer will keep classes if they begin with microformats prefixes or are semantic classes:

* `h-*`
* `p-*`
* `u-*`
* `dt-*`
* `e-*`
* `mention`
  * _The innermost `<a>` will be replaced with a local link to remote profile._ 
* `hashtag`
  * _The innermost `<a>` will be replaced with a local link to hashtag page._
* `ellipsis`
* `invisible`

Links will be kept if the protocol is supported, and converted to text otherwise. _Kept links will get classes `wikilink` ,`wikilink_external`, and `wikilink_<protocol>`._ The following link protocols are supported:

* `http`
* `https`
* `dat`
* `dweb`
* `ipfs`
* `ipns`
* `ssb`
* `gopher`
* `xmpp`
* `magnet`
* `gemini`
* _add:_ `mailto`
