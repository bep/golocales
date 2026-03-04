// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

//go:generate go run main.go
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"maps"
	"os"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/bep/golocales/internal"
	"golang.org/x/text/unicode/cldr"
)

func main() {
	// Load CLDR data to get metazone mapping.
	var decoder cldr.Decoder
	decoder.SetDirFilter("main", "supplemental")
	cldrData, err := decoder.DecodePath("data/core")
	if err != nil {
		log.Fatal(err)
	}
	metazones := internal.BuildMetazoneMap(cldrData)

	locales := make(map[string]internal.LocaleConfig)
	for k, v := range internal.Generate("data/core") {
		locales[strings.ToLower(k)] = v
	}

	data := buildGenData(locales, metazones)

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		log.Fatal(err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatal(err)
	}

	if err := os.WriteFile("../golocales.autogen.go", formatted, 0644); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Generated golocales.autogen.go: %d locales, %d months, %d days, %d strings, %d currency override sets, %d tz name sets, %d metazones\n",
		len(data.Locales), len(data.Months), len(data.Days), len(data.Strings), len(data.CurrencyOverrides), len(data.TZNames), len(data.Metazones))
}

// genData holds all the interned data for code generation.
type genData struct {
	Months            []internal.Months
	Days              []internal.Days
	Strings           []string
	CurrencyDefaults  []kvPair      // sorted by key
	CurrencyOverrides [][]kvPair    // each sorted by key
	Metazones         []kvPair      // sorted by key
	TZNames           [][2][]kvPair // each pair of slices sorted by key
	Locales           []localeEntry
}

type kvPair struct {
	Key, Value string
}

type localeEntry struct {
	Name              string
	MonthsAbbr        int
	MonthsNarrow      int
	MonthsWide        int
	DaysAbbr          int
	DaysNarrow        int
	DaysShort         int
	DaysWide          int
	Decimal           int
	Group             int
	Minus             int
	Percent           int
	PerMille          int
	Plus              int
	PercentPattern    int
	DateFull          int
	DateLong          int
	DateMedium        int
	DateShort         int
	TimeFull          int
	TimeLong          int
	TimeMedium        int
	TimeShort         int
	StdCurrPattern    int
	AcctCurrPattern   int
	CurrencyOverrides int
	TZNames           int
}

// intern is a generic interning helper.
type intern[T comparable] struct {
	values []T
	index  map[T]int
}

func newIntern[T comparable]() *intern[T] {
	return &intern[T]{index: make(map[T]int)}
}

func (in *intern[T]) add(v T) int {
	if idx, ok := in.index[v]; ok {
		return idx
	}
	idx := len(in.values)
	in.values = append(in.values, v)
	in.index[v] = idx
	return idx
}

// sortedKVPairs converts a map to a sorted slice of kvPair.
func sortedKVPairs(m map[string]string) []kvPair {
	if len(m) == 0 {
		return nil
	}
	pairs := make([]kvPair, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kvPair{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool { return pairs[i].Key < pairs[j].Key })
	return pairs
}

// mapKey returns a stable string key for deduplication.
func mapKey(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := slices.Sorted(maps.Keys(m))
	var b strings.Builder
	for _, k := range keys {
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(m[k])
		b.WriteByte(';')
	}
	return b.String()
}

// tzNamesKey returns a stable string key for a TimeZoneNames map.
func tzNamesKey(m internal.TimeZoneNames) string {
	if len(m) == 0 {
		return ""
	}
	keys := slices.Sorted(maps.Keys(m))
	var b strings.Builder
	for _, k := range keys {
		v := m[k]
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(v[0])
		b.WriteByte('/')
		b.WriteString(v[1])
		b.WriteByte(';')
	}
	return b.String()
}

func buildGenData(locales map[string]internal.LocaleConfig, metazones map[string]string) genData {
	months := newIntern[internal.Months]()
	days := newIntern[internal.Days]()
	strs := newIntern[string]()

	// Pre-add empty string at index 0.
	strs.add("")

	// Collect currency defaults from root locale.
	var currencyDefaults map[string]string
	if root, ok := locales["root"]; ok {
		currencyDefaults = root.CurrencySymbols
	}
	if currencyDefaults == nil {
		currencyDefaults = make(map[string]string)
	}

	// Build currency overrides and intern them.
	overrideIndex := make(map[string]int)
	var overrides [][]kvPair
	overrides = append(overrides, nil) // index 0 = no overrides
	overrideIndex[""] = 0

	internOverride := func(m map[string]string) int {
		override := make(map[string]string)
		for code, sym := range m {
			if defSym, ok := currencyDefaults[code]; !ok || defSym != sym {
				override[code] = sym
			}
		}
		if len(override) == 0 {
			return 0
		}
		key := mapKey(override)
		if idx, ok := overrideIndex[key]; ok {
			return idx
		}
		idx := len(overrides)
		overrides = append(overrides, sortedKVPairs(override))
		overrideIndex[key] = idx
		return idx
	}

	// Intern timezone name maps.
	tzIndex := make(map[string]int)
	var tzNames [][2][]kvPair
	tzNames = append(tzNames, [2][]kvPair{}) // index 0 = no tz names
	tzIndex[""] = 0

	internTZNames := func(m internal.TimeZoneNames) int {
		if len(m) == 0 {
			return 0
		}
		key := tzNamesKey(m)
		if idx, ok := tzIndex[key]; ok {
			return idx
		}
		// Split into standard and daylight sorted slices.
		stdMap := make(map[string]string, len(m))
		dstMap := make(map[string]string, len(m))
		for k, v := range m {
			if v[0] != "" {
				stdMap[k] = v[0]
			}
			if v[1] != "" {
				dstMap[k] = v[1]
			}
		}
		idx := len(tzNames)
		tzNames = append(tzNames, [2][]kvPair{sortedKVPairs(stdMap), sortedKVPairs(dstMap)})
		tzIndex[key] = idx
		return idx
	}

	// Sort locale names for deterministic output.
	names := slices.Sorted(maps.Keys(locales))

	var entries []localeEntry
	for _, name := range names {
		cfg := locales[name]
		e := localeEntry{
			Name:            name,
			MonthsAbbr:      months.add(cfg.MonthsAbbreviated),
			MonthsNarrow:    months.add(cfg.MonthsNarrow),
			MonthsWide:      months.add(cfg.MonthsWide),
			DaysAbbr:        days.add(cfg.WeekDaysAbbreviated),
			DaysNarrow:      days.add(cfg.WeekDaysNarrow),
			DaysShort:       days.add(cfg.WeekDaysShort),
			DaysWide:        days.add(cfg.WeekDaysWide),
			Decimal:         strs.add(cfg.Decimal),
			Group:           strs.add(cfg.Group),
			Minus:           strs.add(cfg.Minus),
			Percent:         strs.add(cfg.Percent),
			PerMille:        strs.add(cfg.PerMille),
			Plus:            strs.add(cfg.Plus),
			PercentPattern:  strs.add(cfg.PercentPattern),
			DateFull:        strs.add(cfg.DateFull),
			DateLong:        strs.add(cfg.DateLong),
			DateMedium:      strs.add(cfg.DateMedium),
			DateShort:       strs.add(cfg.DateShort),
			TimeFull:        strs.add(cfg.TimeFull),
			TimeLong:        strs.add(cfg.TimeLong),
			TimeMedium:      strs.add(cfg.TimeMedium),
			TimeShort:       strs.add(cfg.TimeShort),
			StdCurrPattern:  strs.add(cfg.StandardPattern),
			AcctCurrPattern: strs.add(cfg.AccountingPattern),
		}
		e.CurrencyOverrides = internOverride(cfg.CurrencySymbols)
		e.TZNames = internTZNames(cfg.TZNames)
		entries = append(entries, e)
	}

	return genData{
		Months:            months.values,
		Days:              days.values,
		Strings:           strs.values,
		CurrencyDefaults:  sortedKVPairs(currencyDefaults),
		CurrencyOverrides: overrides,
		Metazones:         sortedKVPairs(metazones),
		TZNames:           tzNames,
		Locales:           entries,
	}
}

var funcMap = template.FuncMap{
	"gostr": func(s string) string { return fmt.Sprintf("%q", s) },
}

var tmpl = template.Must(template.New("autogen").Funcs(funcMap).Parse(`// Code generated by gen/main.go; DO NOT EDIT.
package golocales

// kv is a key-value pair used in sorted slices for map-free lookups.
type kv struct{ k, v string }

// localeEntry holds a locale name and its compact config as indices into interned tables.
type localeEntry struct {
	name                                           string
	monthsAbbr, monthsNarrow, monthsWide           uint16
	daysAbbr, daysNarrow, daysShort, daysWide      uint16
	decimal, group, minus, percent, perMille, plus  uint16
	percentPattern                                  uint16
	dateFull, dateLong, dateMedium, dateShort        uint16
	timeFull, timeLong, timeMedium, timeShort        uint16
	stdCurrencyPattern, acctCurrencyPattern         uint16
	currencyOverrides                               uint16
	tzNames                                         uint16
}

type months [12]string
type days [7]string

var _months = [...]months{
{{- range $i, $m := .Months }}
	{ {{- range $j, $s := $m }}{{ if $j }}, {{ end }}{{ gostr $s }}{{ end -}} },
{{- end }}
}

var _days = [...]days{
{{- range $i, $d := .Days }}
	{ {{- range $j, $s := $d }}{{ if $j }}, {{ end }}{{ gostr $s }}{{ end -}} },
{{- end }}
}

var _strings = [...]string{
{{- range $i, $s := .Strings }}
	{{ gostr $s }},
{{- end }}
}

// Sorted slices for binary search lookups (no map allocation at init).

var _currencyDefaults = [...]kv{
{{- range .CurrencyDefaults }}
	{ {{ gostr .Key }}, {{ gostr .Value }} },
{{- end }}
}

var _currencyOverrides = [...][]kv{
{{- range $i, $pairs := .CurrencyOverrides }}
{{- if $pairs }}
	{ {{- range $pairs }}{ {{ gostr .Key }}, {{ gostr .Value }} }, {{ end -}} },
{{- else }}
	nil,
{{- end }}
{{- end }}
}

var _metazones = [...]kv{
{{- range .Metazones }}
	{ {{ gostr .Key }}, {{ gostr .Value }} },
{{- end }}
}

var _tzNames = [...][2][]kv{
{{- range $i, $pair := .TZNames }}
{{- if or (index $pair 0) (index $pair 1) }}
	{
		{ {{- range index $pair 0 }}{ {{ gostr .Key }}, {{ gostr .Value }} }, {{ end -}} },
		{ {{- range index $pair 1 }}{ {{ gostr .Key }}, {{ gostr .Value }} }, {{ end -}} },
	},
{{- else }}
	{nil, nil},
{{- end }}
{{- end }}
}

var _locales = [...]localeEntry{
{{- range .Locales }}
	{ {{ gostr .Name }}, {{ .MonthsAbbr }}, {{ .MonthsNarrow }}, {{ .MonthsWide }}, {{ .DaysAbbr }}, {{ .DaysNarrow }}, {{ .DaysShort }}, {{ .DaysWide }}, {{ .Decimal }}, {{ .Group }}, {{ .Minus }}, {{ .Percent }}, {{ .PerMille }}, {{ .Plus }}, {{ .PercentPattern }}, {{ .DateFull }}, {{ .DateLong }}, {{ .DateMedium }}, {{ .DateShort }}, {{ .TimeFull }}, {{ .TimeLong }}, {{ .TimeMedium }}, {{ .TimeShort }}, {{ .StdCurrPattern }}, {{ .AcctCurrPattern }}, {{ .CurrencyOverrides }}, {{ .TZNames }} },
{{- end }}
}
`))
