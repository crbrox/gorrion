package gorrion

import (
	"fmt"
	"sort"
	"strings"
)

type Option int

const (
	OptMinValue         = OptKeyValues
	OptKeyValues Option = iota
	OptValues
	OptCount
	OptUnique
	OptAppend
	OptMaxValue
	OptInvalid = OptMaxValue
)

func (o Option) String() string {
	switch o {
	case OptKeyValues:
		return "keyValues"
	case OptValues:
		return "values"
	case OptCount:
		return "count"
	case OptUnique:
		return "unique"
	case OptAppend:
		return "append"
	default:
		return "invalidOption"
	}
}
func ParseOpt(s string) Option {
	switch s {
	case "keyValues":
		return OptKeyValues
	case "values":
		return OptValues
	case "count":
		return OptCount
	case "unique":
		return OptUnique
	case "append":
		return OptAppend
	default:
		return OptInvalid
	}
}

type OptionSet [OptMaxValue]bool

func (os *OptionSet) Set(o Option) {
	os[o] = true
}

func (os OptionSet) Get(o Option) bool {
	return os[o]
}

func (os OptionSet) String() string {
	ss := []string{}
	for opt, isOn := range os {
		if isOn {
			ss = append(ss, Option(opt).String())
		}
	}
	sort.Strings(ss)
	return fmt.Sprint(ss)
}

func ParseOptSet(str string) OptionSet {
	optSet := OptionSet{}
	ss := strings.Split(str, ",")
	for _, s := range ss {
		o := ParseOpt(strings.TrimSpace(s))
		if o != OptInvalid {
			optSet.Set(o)
		}
	}
	return optSet
}
