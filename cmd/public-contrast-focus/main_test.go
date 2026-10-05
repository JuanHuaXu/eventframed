package main

import "testing"

func TestFocusQuery(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Which adopted decision, rather than the proposed draft?", "Which adopted decision?"},
		{"Which proposed draft, rather than the adopted decision?", "Which proposed draft?"},
		{"What happened before the decision?", "What happened before the decision?"},
		{"What happened after the decision?", "What happened after the decision?"},
		{"Which A, rather than B, rather than C?", "Which A, rather than B, rather than C?"},
		{", rather than B?", ", rather than B?"},
		{"A, rather than ", "A, rather than "},
		{"", ""},
	}
	for _, c := range cases {
		if got := focusQuery(c.in); got != c.want {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
}
