package affinity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected []string
	}{
		{[]string{"a", "b"}, []string{"b"}, []string{"a"}},
		{[]string{"a", "b"}, []string{"c"}, []string{"a", "b"}},
		{[]string{"a", "b"}, []string{"a", "b"}, nil},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, diff(test.a, test.b))
	}
}

func TestDifference(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected []chunk
	}{
		{
			[]string{},
			[]string{"b", "a"},
			[]chunk{
				{add, "b"},
				{add, "a"},
			},
		},
		{
			[]string{"b", "a"},
			[]string{},
			[]chunk{
				{del, "b"},
				{del, "a"},
			},
		},
		{
			[]string{"a"},
			[]string{"b", "a"},
			[]chunk{
				{add, "b"},
			},
		},
		{
			[]string{"a", "b"},
			[]string{"a"},
			[]chunk{
				{del, "b"},
			},
		},
		{
			[]string{"b", "d", "c"},
			[]string{"d", "e", "f"},
			[]chunk{
				{del, "b"},
				{del, "c"},
				{add, "e"},
				{add, "f"},
			},
		},
	}

	for _, test := range tests {
		assert.Equal(t, test.expected, difference(test.a, test.b))
	}
}
