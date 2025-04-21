package fedisearch

import (
	"math/rand"
	"reflect"
	"slices"
	"strings"
	"testing"
)

func TestState_RequestsToMake(t *testing.T) {
	// NOTE(bouncepaw): this very special seed was used to pick
	// test values very precisely. I pray nothing ever breaks,
	// because changing the test values would be a chore.
	rand.Seed(0b0111001101100001011011000110000101101101)

	type fields struct {
		query    string
		seen     map[string]int
		expected map[string]int
		unseen   []string
		ourID    string
	}
	tests := []struct {
		name   string
		fields fields
		want   []Request
	}{
		{
			"A/First page",
			fields{
				query:    "A",
				seen:     nil,
				expected: nil,
				unseen:   []string{"Alice", "Bob", "Charlie", "David"},
				ourID:    "Betulizer",
			},
			[]Request{
				{"v1", "A", 15, 0, "Betulizer", "Alice"},
				{"v1", "A", 15, 0, "Betulizer", "Bob"},
				{"v1", "A", 15, 0, "Betulizer", "Charlie"},
				{"v1", "A", 20, 0, "Betulizer", "David"},
			},
		},
		{
			"A/Second page",
			fields{
				query: "A",
				seen: map[string]int{
					"Alice": 15, "Bob": 15, "Charlie": 15, "David": 20,
				},
				expected: map[string]int{
					"Alice": 77, "Bob": 4, "David": 17, // Charlie depleted
				},
				unseen: nil,
				ourID:  "Betulizer",
			},
			[]Request{ // Returns 66 bookmarks here
				{"v1", "A", 45, 15, "Betulizer", "Alice"},
				{"v1", "A", 4, 15, "Betulizer", "Bob"},
				{"v1", "A", 17, 20, "Betulizer", "David"},
			},
		},
		{
			"A/Third page",
			fields{
				query: "A",
				seen: map[string]int{
					"Alice": 15 + 45, "Bob": 15 + 4, "Charlie": 15, "David": 20 + 17,
				},
				expected: map[string]int{
					"Alice": 17, // Only Alice stands.
				},
				unseen: nil,
				ourID:  "Betulizer",
			},
			[]Request{
				{"v1", "A", 17, 15 + 45, "Betulizer", "Alice"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &State{
				Query:    tt.fields.query,
				Seen:     tt.fields.seen,
				Expected: tt.fields.expected,
				Unseen:   tt.fields.unseen,
				ourID:    tt.fields.ourID,
			}
			reqs := slices.SortedFunc(
				slices.Values(s.RequestsToMake()),
				func(e Request, e2 Request) int {
					return strings.Compare(e.To, e2.To)
				},
			)
			if got := reqs; !reflect.DeepEqual(got, tt.want) {
				t.Errorf("\ngot  %v,\nwant %v", got, tt.want)
			}
		})
	}
}
