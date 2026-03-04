// Copyright 2026 Bjørn Erik Pedersen
// SPDX-License-Identifier: MIT

package internal

import (
	"fmt"
)

type (
	Months [12]string
	Days   [7]string
)

func (d Days) Slice() []string {
	return d[:]
}

func (m Months) Slice() []string {
	return m[:]
}

type LocaleConfig struct {
	MonthsConfig
	WeekDaysConfig
	NumberConfig
	CurrencyConfig
	DateTimeConfig
	TZNames TimeZoneNames
}

// CurrencyConfig holds locale-specific currency formatting data.
type CurrencyConfig struct {
	// StandardPattern is the CLDR standard currency format pattern (e.g. "#,##0.00 ¤").
	StandardPattern string
	// AccountingPattern is the CLDR accounting format pattern (e.g. "¤ #,##0.00;(¤ #,##0.00)").
	AccountingPattern string
	// CurrencySymbols maps currency code (e.g. "NOK") to its locale symbol (e.g. "kr").
	CurrencySymbols map[string]string
}

type NumberConfig struct {
	Decimal        string
	Group          string
	Minus          string
	Percent        string
	PerMille       string
	Plus           string
	PercentPattern string
}

type DateTimeConfig struct {
	DateFull   string
	DateLong   string
	DateMedium string
	DateShort  string
	TimeFull   string
	TimeLong   string
	TimeMedium string
	TimeShort  string
}

// TimeZoneNames maps metazone name → [standard, daylight] long names.
type TimeZoneNames map[string][2]string

type MonthsConfig struct {
	MonthsAbbreviated Months
	MonthsNarrow      Months
	MonthsWide        Months
}

// TODO1 consider intern.Unique.
type WeekDaysConfig struct {
	WeekDaysAbbreviated Days
	WeekDaysNarrow      Days
	WeekDaysShort       Days
	WeekDaysWide        Days
}

func (d LocaleConfig) String() string {
	return "dateSlices{MonthsAbbreviated: " + fmt.Sprintf("%v", d.MonthsAbbreviated) +
		", MonthsNarrow: " + fmt.Sprintf("%v", d.MonthsNarrow) +
		", MonthsWide: " + fmt.Sprintf("%v", d.MonthsWide) +
		", DaysAbbreviated: " + fmt.Sprintf("%v", d.WeekDaysAbbreviated) +
		", DaysNarrow: " + fmt.Sprintf("%v", d.WeekDaysNarrow) +
		", DaysShort: " + fmt.Sprintf("%v", d.WeekDaysShort) +
		", DaysWide: " + fmt.Sprintf("%v", d.WeekDaysWide) + "}"
}
