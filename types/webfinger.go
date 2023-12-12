package types

import "time"

type WebfingerAcct struct {
	Acct     string
	ActorURL string
	Document []byte
	// Do not set it yourself.
	LastCheckedAt string
}

func (wa WebfingerAcct) CheckTime() time.Time {
	t, err := time.Parse(TimeLayout, wa.LastCheckedAt)
	if err != nil {
		// Whatever
		return time.Now()
	}
	return t
}

func (wa WebfingerAcct) IsStale() bool {
	return wa.CheckTime().Before(time.Now())
}
