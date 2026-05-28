// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package ticks

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/nalgeon/be"
)

func TestTimeRFC3339_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("UTC", func(t *testing.T) {
		t.Parallel()
		ts := TimeRFC3339(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC))
		data, err := json.Marshal(ts)
		be.Equal(t, err, nil)
		be.Equal(t, string(data), `"2024-06-15T12:00:00Z"`)
	})

	t.Run("Non-UTC normalized to UTC", func(t *testing.T) {
		t.Parallel()
		loc := time.FixedZone("UTC+3", 3*60*60)
		ts := TimeRFC3339(time.Date(2024, 6, 15, 15, 0, 0, 0, loc))
		data, err := json.Marshal(ts)
		be.Equal(t, err, nil)
		be.Equal(t, string(data), `"2024-06-15T12:00:00Z"`)
	})
}

func TestTimeRFC3339_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("Valid RFC3339", func(t *testing.T) {
		t.Parallel()
		var ts TimeRFC3339
		err := json.Unmarshal([]byte(`"2024-01-10T08:00:00Z"`), &ts)
		be.Equal(t, err, nil)
		be.Equal(t, ts, TimeRFC3339(time.Date(2024, 1, 10, 8, 0, 0, 0, time.UTC)))
	})

	t.Run("Invalid string", func(t *testing.T) {
		t.Parallel()
		var ts TimeRFC3339
		err := json.Unmarshal([]byte(`"not a time"`), &ts)
		be.True(t, err != nil)
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		t.Parallel()
		var ts TimeRFC3339
		err := json.Unmarshal([]byte(`123`), &ts)
		be.True(t, err != nil)
	})
}

func TestTimeRFC3339_Roundtrip(t *testing.T) {
	t.Parallel()
	original := TimeRFC3339(time.Date(2024, 3, 5, 17, 45, 0, 0, time.UTC))
	data, err := json.Marshal(original)
	be.Equal(t, err, nil)
	var got TimeRFC3339
	be.Equal(t, json.Unmarshal(data, &got), nil)
	be.Equal(t, got, original)
}
