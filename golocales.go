// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

package golocales

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"time"
)

// New returns a new Translator for the given locale,
// or nil if the locale is not supported.
func New(locale string) Translator {
	// Normalize locale.
	locale = strings.ToLower(locale)
	locale = strings.ReplaceAll(locale, "-", "_")

	i, ok := slices.BinarySearchFunc(_locales[:], locale, func(e localeEntry, target string) int {
		return cmp.Compare(e.name, target)
	})
	if !ok {
		return nil
	}
	return &translatorImpl{data: &_locales[i]}
}

type Translator interface {
	// WeekdaysWide returns the full names of the week days, starting with Sunday.
	WeekdaysWide() []string
	// WeekdaysAbbreviated returns the abbreviated names of the week days, starting with Sunday.
	WeekdaysAbbreviated() []string

	// MonthsWide returns the full names of the months, starting with January.
	MonthsWide() []string
	// MonthsAbbreviated returns the abbreviated names of the months, starting with January.
	MonthsAbbreviated() []string

	// FormatDateFull formats the given time using the full date pattern for the locale.
	FormatDateFull(t time.Time) string
	// FormatDateLong formats the given time using the long date pattern for the locale.
	FormatDateLong(t time.Time) string
	// FormatDateMedium formats the given time using the medium date pattern for the locale.
	FormatDateMedium(t time.Time) string
	// FormatDateShort formats the given time using the short date pattern for the locale.
	FormatDateShort(t time.Time) string
	// FormatTimeFull formats the given time using the full time pattern for the locale.
	FormatTimeFull(t time.Time) string
	// FormatTimeLong formats the given time using the long time pattern for the locale.
	FormatTimeLong(t time.Time) string
	// FormatTimeMedium formats the given time using the medium time pattern for the locale.
	FormatTimeMedium(t time.Time) string
	// FormatTimeShort formats the given time using the short time pattern for the locale.
	FormatTimeShort(t time.Time) string

	// FormatNumber formats the given number with the specified number of decimal places.
	FormatNumber(num float64, p int) string
	// FormatAccounting formats the given number as an accounting value with the specified number of decimal places and currency.
	FormatAccounting(num float64, p int, currency string) string
	// FormatPercent formats the given number as a percentage with the specified number of decimal places.
	FormatPercent(num float64, p int) string
	// FormatCurrency formats the given number as a currency value with the specified number of decimal places and currency.
	FormatCurrency(num float64, p int, currency string) string
}

type translatorImpl struct {
	data *localeEntry
}

func (t *translatorImpl) WeekdaysWide() []string {
	d := _days[t.data.daysWide]
	return d[:]
}

func (t *translatorImpl) WeekdaysAbbreviated() []string {
	d := _days[t.data.daysAbbr]
	return d[:]
}

func (t *translatorImpl) MonthsWide() []string {
	m := _months[t.data.monthsWide]
	return m[:]
}

func (t *translatorImpl) MonthsAbbreviated() []string {
	m := _months[t.data.monthsAbbr]
	return m[:]
}

// Date/time formatting.

func (t *translatorImpl) FormatDateFull(tm time.Time) string {
	return t.formatPattern(_strings[t.data.dateFull], tm)
}

func (t *translatorImpl) FormatDateLong(tm time.Time) string {
	return t.formatPattern(_strings[t.data.dateLong], tm)
}

func (t *translatorImpl) FormatDateMedium(tm time.Time) string {
	return t.formatPattern(_strings[t.data.dateMedium], tm)
}

func (t *translatorImpl) FormatDateShort(tm time.Time) string {
	return t.formatPattern(_strings[t.data.dateShort], tm)
}

func (t *translatorImpl) FormatTimeFull(tm time.Time) string {
	return t.formatPattern(_strings[t.data.timeFull], tm)
}

func (t *translatorImpl) FormatTimeLong(tm time.Time) string {
	return t.formatPattern(_strings[t.data.timeLong], tm)
}

func (t *translatorImpl) FormatTimeMedium(tm time.Time) string {
	return t.formatPattern(_strings[t.data.timeMedium], tm)
}

func (t *translatorImpl) FormatTimeShort(tm time.Time) string {
	return t.formatPattern(_strings[t.data.timeShort], tm)
}

// formatPattern interprets a CLDR date/time pattern and formats the given time.
func (t *translatorImpl) formatPattern(pattern string, tm time.Time) string {
	var buf strings.Builder
	i := 0
	for i < len(pattern) {
		c := pattern[i]

		// Quoted literal text.
		if c == '\'' {
			i++
			if i < len(pattern) && pattern[i] == '\'' {
				buf.WriteByte('\'')
				i++
				continue
			}
			for i < len(pattern) && pattern[i] != '\'' {
				buf.WriteByte(pattern[i])
				i++
			}
			if i < len(pattern) {
				i++ // skip closing quote
			}
			continue
		}

		// Pattern letter: consume run of same character.
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' {
			start := i
			for i < len(pattern) && pattern[i] == c {
				i++
			}
			n := i - start
			t.formatField(&buf, c, n, tm)
			continue
		}

		// Literal character.
		buf.WriteByte(c)
		i++
	}
	return buf.String()
}

func (t *translatorImpl) formatField(buf *strings.Builder, c byte, n int, tm time.Time) {
	switch c {
	case 'y': // year
		year := tm.Year()
		if n == 2 {
			fmt.Fprintf(buf, "%02d", year%100)
		} else {
			fmt.Fprintf(buf, "%d", year)
		}
	case 'M': // month
		month := tm.Month()
		switch {
		case n >= 4: // MMMM = wide month name
			m := _months[t.data.monthsWide]
			buf.WriteString(m[month-1])
		case n == 3: // MMM = abbreviated month name
			m := _months[t.data.monthsAbbr]
			buf.WriteString(m[month-1])
		case n == 2: // MM = zero-padded
			fmt.Fprintf(buf, "%02d", month)
		default: // M = no padding
			fmt.Fprintf(buf, "%d", month)
		}
	case 'd': // day of month
		day := tm.Day()
		if n >= 2 {
			fmt.Fprintf(buf, "%02d", day)
		} else {
			fmt.Fprintf(buf, "%d", day)
		}
	case 'E': // weekday
		wd := tm.Weekday() // Sunday=0
		if n >= 4 {        // EEEE = wide
			d := _days[t.data.daysWide]
			buf.WriteString(d[wd])
		} else { // EEE = abbreviated
			d := _days[t.data.daysAbbr]
			buf.WriteString(d[wd])
		}
	case 'H': // hour 24h
		hour := tm.Hour()
		if n >= 2 {
			fmt.Fprintf(buf, "%02d", hour)
		} else {
			fmt.Fprintf(buf, "%d", hour)
		}
	case 'h': // hour 12h
		hour := tm.Hour() % 12
		if hour == 0 {
			hour = 12
		}
		if n >= 2 {
			fmt.Fprintf(buf, "%02d", hour)
		} else {
			fmt.Fprintf(buf, "%d", hour)
		}
	case 'm': // minute
		min := tm.Minute()
		if n >= 2 {
			fmt.Fprintf(buf, "%02d", min)
		} else {
			fmt.Fprintf(buf, "%d", min)
		}
	case 's': // second
		sec := tm.Second()
		if n >= 2 {
			fmt.Fprintf(buf, "%02d", sec)
		} else {
			fmt.Fprintf(buf, "%d", sec)
		}
	case 'a': // AM/PM
		if tm.Hour() < 12 {
			// This is a choice based mostly on how Hugo's old library did it,
			// but this is a common convention in digital writing.
			buf.WriteString("am")
		} else {
			buf.WriteString("pm")
		}
	case 'z': // timezone
		if n >= 4 { // zzzz = long timezone name
			buf.WriteString(t.longTimezoneName(tm))
		} else { // z = short timezone abbreviation
			name, _ := tm.Zone()
			buf.WriteString(name)
		}
	default:
		// Unknown pattern letter: output as-is.
		for range n {
			buf.WriteByte(c)
		}
	}
}

func (t *translatorImpl) longTimezoneName(tm time.Time) string {
	loc := tm.Location().String()
	meta, ok := kvLookup(_metazones[:], loc)
	if !ok {
		name, _ := tm.Zone()
		return name
	}
	tzn := _tzNames[t.data.tzNames]
	idx := 0
	if tm.IsDST() {
		idx = 1
	}
	if name, ok := kvLookup(tzn[idx], meta); ok {
		return name
	}
	name, _ := tm.Zone()
	return name
}

func (t *translatorImpl) currencySymbol(code string) string {
	code = strings.ToUpper(code)
	overrides := _currencyOverrides[t.data.currencyOverrides]
	if s, ok := kvLookup(overrides, code); ok {
		return s
	}
	if s, ok := kvLookup(_currencyDefaults[:], code); ok {
		return s
	}
	return code
}

func (t *translatorImpl) FormatAccounting(num float64, p int, currency string) string {
	symbol := t.currencySymbol(currency)
	formatted := t.FormatNumber(num, p)

	pattern := _strings[t.data.stdCurrencyPattern]
	if pattern == "" {
		return formatted + " " + symbol
	}

	// Handle positive/negative subpatterns (separated by ";").
	positive, _, _ := strings.Cut(pattern, ";")

	// Replace ¤ with symbol and number pattern with formatted number.
	result := strings.Replace(positive, "¤", symbol, 1)
	numPattern := strings.TrimSpace(strings.Replace(positive, "¤", "", 1))
	result = strings.Replace(result, numPattern, formatted, 1)

	return result
}

func (t *translatorImpl) FormatPercent(num float64, p int) string {
	formatted := t.FormatNumber(num, p)
	pattern := _strings[t.data.percentPattern]
	if pattern == "" {
		return formatted + "%"
	}
	percentSym := _strings[t.data.percent]
	result := strings.Replace(pattern, "%", percentSym, 1)
	numPattern := strings.TrimSpace(strings.Replace(pattern, "%", "", 1))
	result = strings.Replace(result, numPattern, formatted, 1)
	return result
}

func (t *translatorImpl) FormatCurrency(num float64, p int, currency string) string {
	symbol := t.currencySymbol(currency)
	formatted := t.FormatNumber(num, p)

	pattern := _strings[t.data.stdCurrencyPattern]
	if pattern == "" {
		return formatted + " " + symbol
	}

	positive, _, _ := strings.Cut(pattern, ";")
	result := strings.Replace(positive, "¤", symbol, 1)
	numPattern := strings.TrimSpace(strings.Replace(positive, "¤", "", 1))
	result = strings.Replace(result, numPattern, formatted, 1)
	return result
}

func (t *translatorImpl) FormatNumber(num float64, p int) string {
	negative := num < 0
	if negative {
		num = -num
	}

	decimal := _strings[t.data.decimal]
	group := _strings[t.data.group]
	minus := _strings[t.data.minus]

	// Round to the requested precision.
	if p > 0 {
		shift := math.Pow10(p)
		num = math.Round(num*shift) / shift
	} else {
		num = math.Round(num)
	}
	s := strconv.FormatFloat(num, 'f', p, 64)

	// Split into integer and fractional parts.
	intPart, fracPart, hasFrac := strings.Cut(s, ".")

	// Add grouping separators to the integer part (groups of 3 from the right).
	if group != "" && len(intPart) > 3 {
		var buf strings.Builder
		start := len(intPart) % 3
		if start == 0 {
			start = 3
		}
		buf.WriteString(intPart[:start])
		for i := start; i < len(intPart); i += 3 {
			buf.WriteString(group)
			buf.WriteString(intPart[i : i+3])
		}
		intPart = buf.String()
	}

	var result string
	if hasFrac {
		result = intPart + decimal + fracPart
	} else {
		result = intPart
	}

	if negative {
		result = minus + result
	}
	return result
}

// kvLookup does a binary search on a sorted slice of kv pairs.
func kvLookup(pairs []kv, key string) (string, bool) {
	i, ok := slices.BinarySearchFunc(pairs, key, func(p kv, target string) int {
		return cmp.Compare(p.k, target)
	})
	if ok {
		return pairs[i].v, true
	}
	return "", false
}
