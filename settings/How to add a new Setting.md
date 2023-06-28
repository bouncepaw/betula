# How to add a new Setting?

2023-06-28 I decided to add a new Setting: Custom CSS. Thought that I should document it for a future Settingist.

As you can see, there are quite a lot of steps to add a new setting. That should discourage you from adding settings. No one really wants them.

## Update `types.Settings`
In `types.go`:

```go
package types

// ...

type Settings struct {
	// ...
	CustomCSS                 string
}
```

## Update form
In `settings.gohtml`:

```html
<div>
    <label for="custom-css">Custom CSS</label>
	<textarea id="custom-css" name="custom-css" placeholder="p { color: red }">{{.CustomCSS}}</textarea>
	<p class="input-caption">
        This stylesheet will be served right after the original Betula stylesheet.
    </p>
</div>
```

The `<input>`'s `id` and `name` should be the same to confuse people less.

## Add DB key

In `betula-meta-keys.go` add a key for the entry in `BetulaMeta` table. When choosing the key, try to come up with something that makes when looked at in the database.

```go
package db

type BetulaMetaKey string

const (
   // ...
	BetulaMetaCustomCSS         BetulaMetaKey = "Custom CSS"
)
```

## Update `settings` package
In `settings.go` set up the cached getter for the setting, update `SetSettings`

```go
package settings

// ...

func Index() {
	// ...
	cache.CustomCSS = db.MetaEntry[string](db.BetulaMetaCustomCSS)
}

// the cached getter:
func CustomCSS() string { return cache.CustomCSS }

func SetSettings(settings types.Settings) {
	// ...
	db.SetMetaEntry(db.BetulaMetaCustomCSS, settings.CustomCSS)
	Index()
}
```

## Update the handler
In `handlers.go` in `handlerSettings` mention the new field of `types.Settings` everywhere while making sense.

## Implement the feature
There is no guide for that, every feature is unique.

## Test
Manual test is mandatory, an automatic test would be even better!