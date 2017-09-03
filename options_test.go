package gorrion

import "testing"

func TestParseOpt(t *testing.T) {
	var cases = []struct {
		str string
		opt Option
	}{
		{"keyValues", OptKeyValues},
		{"values", OptValues},
		{"count", OptCount},
		{"unique", OptUnique},
		{"append", OptAppend},
		{"", OptInvalid},
		{"x", OptInvalid},
	}
	for _, c := range cases {
		o := ParseOpt(c.str)
		if o != c.opt {
			t.Errorf("wanted %#v, got %#v\n", c.opt, o)
		}
	}
}

func TestOption_String(t *testing.T) {
	var cases = []struct {
		str string
		opt Option
	}{
		{"keyValues", OptKeyValues},
		{"values", OptValues},
		{"count", OptCount},
		{"unique", OptUnique},
		{"append", OptAppend},
		{"invalidOption", OptInvalid},
	}
	for _, c := range cases {
		str := c.opt.String()
		if str != c.str {
			t.Errorf("wanted %#v, got %#v\n", c.str, str)
		}
	}
}

func TestOptionSet_GetSet(t *testing.T) {
	for i := OptMinValue; i < OptMaxValue; i++ {
		var os OptionSet

		os.Set(i)

		for j := OptMinValue; j < i; j++ {
			if os.Get(j) {
				t.Fatalf("wanted %s to be false", j)
			}
		}

		for j := i + 1; j < OptMaxValue; j++ {
			if os.Get(j) {
				t.Fatalf("wanted %s to be false", j)
			}
		}

		if !os.Get(i) {
			t.Errorf("wanted %s to be true", i)
		}
	}
}

func TestParseOptSet(t *testing.T) {
	var cases = map[string]OptionSet{
		"keyValues":                           {OptKeyValues: true},
		"values":                              {OptValues: true},
		"count":                               {OptCount: true},
		"unique":                              {OptUnique: true},
		"append":                              {OptAppend: true},
		"append, unique":                      {OptAppend: true, OptUnique: true},
		"unique, append":                      {OptAppend: true, OptUnique: true},
		"unique, count, append":               {OptAppend: true, OptUnique: true, OptCount: true},
		"values, values, values, values":      {OptValues: true},
		" values,           keyValues       ": {OptValues: true, OptKeyValues: true},
	}
	for str, optS := range cases {
		if got := ParseOptSet(str); got != optS {
			t.Errorf("wanted %s, got %s (%q)", optS, got, str)
		}
	}
}

func TestOptionSet_String(t *testing.T) {
	var cases = map[OptionSet]string{
		{OptKeyValues: true}:                               "[keyValues]",
		{OptValues: true}:                                  "[values]",
		{OptCount: true}:                                   "[count]",
		{OptUnique: true}:                                  "[unique]",
		{OptAppend: true}:                                  "[append]",
		{OptAppend: true, OptUnique: true}:                 "[append unique]",
		{OptAppend: true, OptUnique: true}:                 "[append unique]",
		{OptAppend: true, OptUnique: true, OptCount: true}: "[append count unique]",
		{OptValues: true, OptKeyValues: true}:              "[keyValues values]",
	}

	for optS, str := range cases {
		if got := optS.String(); got != str {
			t.Errorf("wanted %s, got %s (%s)", str, got, optS)
		}
	}
}
