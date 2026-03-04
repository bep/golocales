[![Tests on Linux, MacOS and Windows](https://github.com/bep/golocales/workflows/Test/badge.svg)](https://github.com/bep/golocales/actions?query=workflow:Test)
[![Go Report Card](https://goreportcard.com/badge/github.com/bep/golocales)](https://goreportcard.com/report/github.com/bep/golocales)
[![GoDoc](https://godoc.org/github.com/bep/golocales?status.svg)](https://godoc.org/github.com/bep/golocales)

A narrow API for locale aware date and number fornmatting using CLDR 48.1:

```go
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
```

Implemented to replace some old and unmaintained libraries used for this in [Hugo](https://github.com/gohugoio/hugo/). Hugo will most likely soonish move over to use the `text/**` packages, so I'm not sure how long lived this package will be, either.