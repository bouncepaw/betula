package search

import (
	"reflect"
	"testing"
)

type entry struct {
	query string
	// expected:
	clearedQuery string
	includedTags []string
	excludedTags []string
}

func TestFor(t *testing.T) {
	table := []entry{
		{"", "", nil, nil},
		{"kobuna", "kobuna", nil, nil},
		{"#kobuna -#kinoko", "", []string{"kobuna"}, []string{"kinoko"}},
		{"-#kinoko osaka", "osaka", nil, []string{"kinoko"}},
		{"#kobuna #sakana school", "school", []string{"kobuna", "sakana"}, nil},
	}
	for _, entry := range table {
		query, excludedTags := extractWithRegex(entry.query, excludeTagRe)
		query, includedTags := extractWithRegex(query, includeTagRe)

		if query != entry.clearedQuery {
			t.Errorf("Expect %q Got %q", entry.clearedQuery, query)
		}

		if !reflect.DeepEqual(includedTags, entry.includedTags) {
			t.Errorf("Expect %q Got %q", entry.includedTags, includedTags)
		}

		if !reflect.DeepEqual(excludedTags, entry.excludedTags) {
			t.Errorf("Expect %q Got %q", entry.excludedTags, excludedTags)
		}
	}
}
