package storefront

import (
	"testing"
	"time"
)

func TestNormalizeReleaseDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      *StoreReleaseDate
		precision  ReleaseDatePrecision
		exact      string
		start      string
		end        string
		year       int
		month      int
		quarter    int
		comingSoon bool
		raw        string
	}{
		{name: "nil", precision: ReleaseDatePrecisionNone},
		{name: "empty coming soon", value: &StoreReleaseDate{ComingSoon: true}, precision: ReleaseDatePrecisionTBA, comingSoon: true},
		{name: "empty released", value: &StoreReleaseDate{}, precision: ReleaseDatePrecisionNone},
		{name: "day short first", value: &StoreReleaseDate{Date: "16 Nov, 2009"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11, raw: "16 Nov, 2009"},
		{name: "day short month first", value: &StoreReleaseDate{Date: "Nov 16, 2009"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "day month case", value: &StoreReleaseDate{Date: "nOv 16, 2009"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "day full first", value: &StoreReleaseDate{Date: "16 November, 2009"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "day full month first", value: &StoreReleaseDate{Date: "November 16, 2009"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "iso", value: &StoreReleaseDate{Date: "2009-11-16"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "dotted", value: &StoreReleaseDate{Date: "2009.11.16"}, precision: ReleaseDatePrecisionDay, exact: "2009-11-16", start: "2009-11-16", end: "2009-11-16", year: 2009, month: 11},
		{name: "month full", value: &StoreReleaseDate{ComingSoon: true, Date: "November 2026"}, precision: ReleaseDatePrecisionMonth, start: "2026-11-01", end: "2026-11-30", year: 2026, month: 11, comingSoon: true},
		{name: "month short leap", value: &StoreReleaseDate{Date: "Feb 2028"}, precision: ReleaseDatePrecisionMonth, start: "2028-02-01", end: "2028-02-29", year: 2028, month: 2},
		{name: "month short non leap", value: &StoreReleaseDate{Date: "Feb 2027"}, precision: ReleaseDatePrecisionMonth, start: "2027-02-01", end: "2027-02-28", year: 2027, month: 2},
		{name: "q1", value: &StoreReleaseDate{Date: "Q1 2026"}, precision: ReleaseDatePrecisionQuarter, start: "2026-01-01", end: "2026-03-31", year: 2026, quarter: 1},
		{name: "q2", value: &StoreReleaseDate{Date: "Q2 2026"}, precision: ReleaseDatePrecisionQuarter, start: "2026-04-01", end: "2026-06-30", year: 2026, quarter: 2},
		{name: "q3 case", value: &StoreReleaseDate{ComingSoon: true, Date: "q3 2026"}, precision: ReleaseDatePrecisionQuarter, start: "2026-07-01", end: "2026-09-30", year: 2026, quarter: 3, comingSoon: true},
		{name: "q4", value: &StoreReleaseDate{Date: "Q4 2026"}, precision: ReleaseDatePrecisionQuarter, start: "2026-10-01", end: "2026-12-31", year: 2026, quarter: 4},
		{name: "year", value: &StoreReleaseDate{Date: "2026"}, precision: ReleaseDatePrecisionYear, start: "2026-01-01", end: "2026-12-31", year: 2026},
		{name: "coming soon", value: &StoreReleaseDate{ComingSoon: true, Date: "Coming Soon"}, precision: ReleaseDatePrecisionTBA, comingSoon: true},
		{name: "tba case and trim", value: &StoreReleaseDate{Date: "  tBa  "}, precision: ReleaseDatePrecisionTBA, raw: "tBa"},
		{name: "announced", value: &StoreReleaseDate{Date: "To Be Announced"}, precision: ReleaseDatePrecisionTBA},
		{name: "unknown", value: &StoreReleaseDate{ComingSoon: true, Date: "  Late 2027  "}, precision: ReleaseDatePrecisionUnknown, comingSoon: true, raw: "Late 2027"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := NormalizeReleaseDate(tt.value)
			if got.Precision != tt.precision || got.Year != tt.year || got.Month != tt.month || got.Quarter != tt.quarter || got.ComingSoon != tt.comingSoon {
				t.Fatalf("NormalizeReleaseDate() = %#v", got)
			}
			if tt.raw != "" && got.RawText != tt.raw {
				t.Fatalf("RawText = %q, want %q", got.RawText, tt.raw)
			}
			assertReleaseDate(t, "ExactDate", got.ExactDate, tt.exact)
			assertReleaseDate(t, "RangeStart", got.RangeStart, tt.start)
			assertReleaseDate(t, "RangeEnd", got.RangeEnd, tt.end)
		})
	}
}

func TestReleaseDatePrecisionConstants(t *testing.T) {
	t.Parallel()
	want := map[ReleaseDatePrecision]string{
		ReleaseDatePrecisionDay: "day", ReleaseDatePrecisionMonth: "month",
		ReleaseDatePrecisionQuarter: "quarter", ReleaseDatePrecisionYear: "year",
		ReleaseDatePrecisionTBA: "tba", ReleaseDatePrecisionNone: "none",
		ReleaseDatePrecisionUnknown: "unknown",
	}
	for value, text := range want {
		if string(value) != text {
			t.Fatalf("precision %q = %q", text, value)
		}
	}
}

func assertReleaseDate(t *testing.T, name string, got *time.Time, want string) {
	t.Helper()
	if want == "" {
		if got != nil {
			t.Fatalf("%s = %s, want nil", name, got)
		}
		return
	}
	if got == nil || got.Location() != time.UTC || got.Format(time.DateOnly) != want {
		t.Fatalf("%s = %v, want %s at UTC midnight", name, got, want)
	}
}
