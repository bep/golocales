// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

package golocales

import (
	"strings"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
)

func TestTranslator(t *testing.T) {
	c := qt.New(t)
	tr := New("nn")
	c.Assert(tr, qt.Not(qt.IsNil))

	// Date formatting.
	loc, err := time.LoadLocation("America/Toronto")
	c.Assert(err, qt.IsNil)
	fixed := time.FixedZone("OTHER", -4)
	timeInLoc1 := time.Date(2016, 0o2, 0o3, 9, 5, 1, 0, loc)
	timeInLoc2 := time.Date(2016, 0o2, 0o3, 20, 5, 1, 0, loc)
	timeInFixed := time.Date(2016, 0o2, 0o3, 20, 5, 1, 0, fixed)

	c.Assert(tr.FormatTimeFull(timeInLoc1), qt.Equals, "09:05:01 normaltid for den nordamerikanske austkysten")
	c.Assert(tr.FormatTimeFull(timeInFixed), qt.Equals, "20:05:01 OTHER")
	c.Assert(tr.FormatTimeLong(timeInLoc1), qt.Equals, "09:05:01 EST")
	c.Assert(tr.FormatTimeLong(timeInLoc2), qt.Equals, "20:05:01 EST")
	c.Assert(tr.FormatTimeMedium(timeInLoc1), qt.Equals, "09:05:01")
	c.Assert(tr.FormatTimeMedium(timeInLoc2), qt.Equals, "20:05:01")
	c.Assert(tr.FormatTimeShort(timeInLoc1), qt.Equals, "09:05")
	c.Assert(tr.FormatTimeShort(timeInLoc2), qt.Equals, "20:05")

	c.Assert(tr.FormatDateFull(timeInLoc1), qt.Equals, "onsdag 3. februar 2016")
	c.Assert(tr.FormatDateLong(timeInLoc1), qt.Equals, "3. februar 2016")
	c.Assert(tr.FormatDateMedium(timeInLoc1), qt.Equals, "3. feb. 2016")
	c.Assert(tr.FormatDateShort(timeInLoc1), qt.Equals, "03.02.16")

	// Week days.
	c.Assert(tr.WeekdaysAbbreviated(), qt.DeepEquals, []string{"sø.", "må.", "ty.", "on.", "to.", "fr.", "la."})
	c.Assert(tr.WeekdaysWide(), qt.DeepEquals, []string{"søndag", "måndag", "tysdag", "onsdag", "torsdag", "fredag", "laurdag"})

	// Months.
	c.Assert(tr.MonthsAbbreviated(), qt.DeepEquals, []string{"jan.", "feb.", "mars", "apr.", "mai", "juni", "juli", "aug.", "sep.", "okt.", "nov.", "des."})
	c.Assert(tr.MonthsWide(), qt.DeepEquals, []string{"januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember"})

	// Number formatting.
	pi := 3.14159
	c.Assert(tr.FormatNumber(pi, 3), qt.Equals, "3,142")

	c.Assert(tr.FormatNumber(1123456.5643, 2), qt.Equals, "1\u00a0123\u00a0456,56")
	c.Assert(tr.FormatAccounting(550.5643, 2, "USD"), qt.Equals, "550,56\u00a0USD")
	c.Assert(tr.FormatAccounting(123.567, 2, "NOK"), qt.Equals, "123,57\u00a0kr")
	c.Assert(tr.FormatAccounting(123.567, 2, "NoK"), qt.Equals, "123,57\u00a0kr")
	c.Assert(tr.FormatPercent(15, 0), qt.Equals, "15\u00a0%")
	c.Assert(tr.FormatCurrency(1123456.5643, 2, "USD"), qt.Equals, "1\u00a0123\u00a0456,56\u00a0USD")
	c.Assert(tr.FormatCurrency(1123456.5643, 2, "NOK"), qt.Equals, "1\u00a0123\u00a0456,56\u00a0kr")
	c.Assert(tr.FormatCurrency(1123456.5643, 2, "nok"), qt.Equals, "1\u00a0123\u00a0456,56\u00a0kr")
}

func TestNew(t *testing.T) {
	c := qt.New(t)

	checkOne := func(locale string) {
		tr := New(locale)
		c.Assert(tr, qt.Not(qt.IsNil))
	}

	for _, locale := range []string{"en", "de", "fr", "zh", "no", "nb", "nb_NO", "nn", "nn_NO"} {
		checkOne(locale)
		checkOne(strings.ToUpper(locale))
		checkOne(strings.ReplaceAll(locale, "_", "-"))
	}

	tr := New("nonexistent")
	c.Assert(tr, qt.IsNil)
}

func TestTranslatorQuotedLiterals(t *testing.T) {
	c := qt.New(t)
	// fr_CA uses patterns like: HH 'h' mm 'min' ss 's'
	tr := New("fr_CA")
	c.Assert(tr, qt.Not(qt.IsNil))

	loc, err := time.LoadLocation("America/Toronto")
	c.Assert(err, qt.IsNil)
	tm := time.Date(2024, 3, 15, 14, 30, 45, 0, loc)

	// TimeMedium: HH 'h' mm 'min' ss 's' → "14 h 30 min 45 s"
	c.Assert(tr.FormatTimeMedium(tm), qt.Equals, "14 h 30 min 45 s")

	// TimeShort: HH 'h' mm → "14 h 30"
	c.Assert(tr.FormatTimeShort(tm), qt.Equals, "14 h 30")
}

func TestTranslatorEnglish12h(t *testing.T) {
	c := qt.New(t)
	tr := New("en")
	c.Assert(tr, qt.Not(qt.IsNil))

	loc, err := time.LoadLocation("America/New_York")
	c.Assert(err, qt.IsNil)

	// Morning time (AM).
	morning := time.Date(2024, 7, 4, 9, 5, 1, 0, loc)
	// Afternoon time (PM).
	afternoon := time.Date(2024, 7, 4, 15, 30, 45, 0, loc)
	// Midnight (12 AM).
	midnight := time.Date(2024, 1, 1, 0, 0, 0, 0, loc)

	// TimeFull uses h:mm:ss a zzzz pattern — covers 'h' (12h), 'a' (AM/PM), 'zzzz' (long tz).
	c.Assert(tr.FormatTimeFull(morning), qt.Equals, "9:05:01\u202fam Eastern Daylight Time")
	c.Assert(tr.FormatTimeFull(afternoon), qt.Equals, "3:30:45\u202fpm Eastern Daylight Time")

	// Midnight: h=12, AM.
	c.Assert(tr.FormatTimeFull(midnight), qt.Equals, "12:00:00\u202fam Eastern Standard Time")

	// DateFull covers EEEE (wide weekday) and MMMM (wide month).
	c.Assert(tr.FormatDateFull(morning), qt.Equals, "Thursday, July 4, 2024")

	// DateShort covers single-digit M and d.
	c.Assert(tr.FormatDateShort(morning), qt.Equals, "7/4/24")
}

func TestFormatNumberEdgeCases(t *testing.T) {
	c := qt.New(t)
	tr := New("en")
	c.Assert(tr, qt.Not(qt.IsNil))

	// Negative number.
	c.Assert(tr.FormatNumber(-1234.56, 2), qt.Equals, "-1,234.56")

	// Small number (no grouping needed).
	c.Assert(tr.FormatNumber(42, 0), qt.Equals, "42")

	// Zero precision.
	c.Assert(tr.FormatNumber(1234567, 0), qt.Equals, "1,234,567")

	// Zero.
	c.Assert(tr.FormatNumber(0, 2), qt.Equals, "0.00")
}

func TestFormatNumberRounding(t *testing.T) {
	c := qt.New(t)
	tr := New("en")
	c.Assert(tr, qt.Not(qt.IsNil))

	// FormatNumber rounding.
	c.Assert(tr.FormatNumber(3.14159, 3), qt.Equals, "3.142")
	c.Assert(tr.FormatNumber(2.5, 0), qt.Equals, "3")
	c.Assert(tr.FormatNumber(1.005, 2), qt.Equals, "1.00") // IEEE 754: 1.005 is actually 1.00499... in float64.
	c.Assert(tr.FormatNumber(9.9999, 2), qt.Equals, "10.00")

	// FormatCurrency rounding.
	c.Assert(tr.FormatCurrency(123.567, 2, "USD"), qt.Equals, "$123.57")
	c.Assert(tr.FormatCurrency(99.995, 2, "USD"), qt.Equals, "$100.00")

	// FormatAccounting rounding.
	c.Assert(tr.FormatAccounting(123.567, 2, "USD"), qt.Equals, "$123.57")

	// FormatPercent rounding.
	c.Assert(tr.FormatPercent(33.456, 1), qt.Equals, "33.5%")
	c.Assert(tr.FormatPercent(99.99, 0), qt.Equals, "100%")
}

func TestFormatCurrencyFallbackToCode(t *testing.T) {
	c := qt.New(t)
	tr := New("en")
	c.Assert(tr, qt.Not(qt.IsNil))

	// Use a currency code that has no symbol override and no default symbol.
	result := tr.FormatCurrency(100, 2, "XYZ")
	c.Assert(result, qt.Contains, "XYZ")
	c.Assert(result, qt.Contains, "100.00")
}

func TestFormatEmptyPatternFallbacks(t *testing.T) {
	c := qt.New(t)
	// "aa" locale has empty currency/percent patterns.
	tr := New("aa")
	c.Assert(tr, qt.Not(qt.IsNil))

	// Currency with empty pattern falls back to "num symbol".
	result := tr.FormatCurrency(100, 2, "USD")
	c.Assert(result, qt.Contains, "100")

	// Accounting with empty pattern.
	result = tr.FormatAccounting(100, 2, "USD")
	c.Assert(result, qt.Contains, "100")

	// Percent with empty pattern.
	result = tr.FormatPercent(50, 0)
	c.Assert(result, qt.Contains, "50")
	c.Assert(result, qt.Contains, "%")
}

func TestFormatAccountingAndPercent(t *testing.T) {
	c := qt.New(t)
	tr := New("en")
	c.Assert(tr, qt.Not(qt.IsNil))

	// Accounting format.
	result := tr.FormatAccounting(1234.56, 2, "USD")
	c.Assert(result, qt.Contains, "$")
	c.Assert(result, qt.Contains, "1,234.56")

	// Percent format.
	pct := tr.FormatPercent(75.5, 1)
	c.Assert(pct, qt.Contains, "75.5")
	c.Assert(pct, qt.Contains, "%")
}

func BenchmarkAll(b *testing.B) {
	tr := New("en")
	loc, _ := time.LoadLocation("America/New_York")
	tm := time.Date(2024, 7, 4, 15, 30, 45, 0, loc)

	b.Run("New", func(b *testing.B) {
		for range b.N {
			New("en")
		}
	})
	b.Run("FormatDateFull", func(b *testing.B) {
		for range b.N {
			tr.FormatDateFull(tm)
		}
	})
	b.Run("FormatDateShort", func(b *testing.B) {
		for range b.N {
			tr.FormatDateShort(tm)
		}
	})
	b.Run("FormatTimeFull", func(b *testing.B) {
		for range b.N {
			tr.FormatTimeFull(tm)
		}
	})
	b.Run("FormatTimeShort", func(b *testing.B) {
		for range b.N {
			tr.FormatTimeShort(tm)
		}
	})
	b.Run("FormatNumber", func(b *testing.B) {
		for range b.N {
			tr.FormatNumber(1234567.89, 2)
		}
	})
	b.Run("FormatCurrency", func(b *testing.B) {
		for range b.N {
			tr.FormatCurrency(1234567.89, 2, "USD")
		}
	})
	b.Run("FormatPercent", func(b *testing.B) {
		for range b.N {
			tr.FormatPercent(75.5, 1)
		}
	})
	b.Run("WeekdaysWide", func(b *testing.B) {
		for range b.N {
			tr.WeekdaysWide()
		}
	})
	b.Run("MonthsWide", func(b *testing.B) {
		for range b.N {
			tr.MonthsWide()
		}
	})
}
