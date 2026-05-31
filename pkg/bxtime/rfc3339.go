// SPDX-FileCopyrightText: 2026 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

package bxtime

import (
	"encoding/json"
	"time"
)

// TimeRFC3339 is a point in time that marshals to/from RFC3339 strings in JSON.
// TODO: use everywhere.
type TimeRFC3339 time.Time

func (t TimeRFC3339) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Time(t).UTC().Format(time.RFC3339))
}

func (t *TimeRFC3339) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	parsed, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	*t = TimeRFC3339(parsed)
	return nil
}
