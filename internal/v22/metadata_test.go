package v22

import "testing"

func TestExtractMetadataTable(t *testing.T) {
	// table test for extractMetadataKey
	var tests = []struct {
		input string
		want  string
	}{
		{"metadata[foo]", "foo"},
		{"metadata[bar]", "bar"},
		{"metadata[foobar]", "foobar"},
		{"metadata[foo bar]", "foo bar"},
		{"metadata[foo-bar]", "foo-bar"},
		{"metadata[foo_bar]", "foo_bar"},
		{"metadata[foo.bar]", "foo.bar"},
		{"metadata[foo:bar]", "foo:bar"},
		{"metadata[foo,bar]", "foo,bar"},
		{"metadata[foo;bar]", "foo;bar"},
		{"metadata[foo!bar]", "foo!bar"},
		{"metadata[foo?bar]", "foo?bar"},
		{"metadata[ ]", ""},
		{"metadata[]", ""},
		{"", ""},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			if got := extractMetadataKey(test.input); got != test.want {
				t.Errorf("extractMetadataKey(%q) = %q, want %q", test.input, got, test.want)
			}
		})

	}
}
