package storefront

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ReleaseDatePrecision describes the precision represented by a normalized
// Storefront release date.
type ReleaseDatePrecision string

const (
	// ReleaseDatePrecisionDay represents an exact calendar date.
	ReleaseDatePrecisionDay ReleaseDatePrecision = "day"
	// ReleaseDatePrecisionMonth represents a calendar month.
	ReleaseDatePrecisionMonth ReleaseDatePrecision = "month"
	// ReleaseDatePrecisionQuarter represents one calendar quarter.
	ReleaseDatePrecisionQuarter ReleaseDatePrecision = "quarter"
	// ReleaseDatePrecisionYear represents one calendar year.
	ReleaseDatePrecisionYear ReleaseDatePrecision = "year"
	// ReleaseDatePrecisionTBA represents a forthcoming but unspecified date.
	ReleaseDatePrecisionTBA ReleaseDatePrecision = "tba"
	// ReleaseDatePrecisionNone represents an absent release date.
	ReleaseDatePrecisionNone ReleaseDatePrecision = "none"
	// ReleaseDatePrecisionUnknown represents non-empty text the normalizer does not recognize.
	ReleaseDatePrecisionUnknown ReleaseDatePrecision = "unknown"
)

// NormalizedReleaseDate is a calculable representation of Storefront release
// date text. Range endpoints are inclusive calendar dates at UTC midnight.
type NormalizedReleaseDate struct {
	ComingSoon bool                 `json:"coming_soon"`
	Precision  ReleaseDatePrecision `json:"precision"`
	ExactDate  *time.Time           `json:"exact_date,omitempty"`
	Year       int                  `json:"year,omitempty"`
	Month      int                  `json:"month,omitempty"`
	Quarter    int                  `json:"quarter,omitempty"`
	RangeStart *time.Time           `json:"range_start,omitempty"`
	RangeEnd   *time.Time           `json:"range_end,omitempty"`
	RawText    string               `json:"raw_text"`
}

var (
	releaseQuarterPattern = regexp.MustCompile(`(?i)^q([1-4])\s+([0-9]{4})$`)
	releaseYearPattern    = regexp.MustCompile(`^[0-9]{4}$`)
)

// NormalizeReleaseDate converts Steam's English Storefront release-date text
// into a stable typed result. Unrecognized text is preserved with unknown
// precision and never causes the surrounding Storefront response to fail.
func NormalizeReleaseDate(value *StoreReleaseDate) NormalizedReleaseDate {
	if value == nil {
		return NormalizedReleaseDate{Precision: ReleaseDatePrecisionNone}
	}

	raw := strings.TrimSpace(value.Date)
	result := NormalizedReleaseDate{
		ComingSoon: value.ComingSoon,
		RawText:    raw,
	}
	if raw == "" {
		if value.ComingSoon {
			result.Precision = ReleaseDatePrecisionTBA
		} else {
			result.Precision = ReleaseDatePrecisionNone
		}
		return result
	}

	switch strings.ToLower(raw) {
	case "coming soon", "tba", "to be announced":
		result.Precision = ReleaseDatePrecisionTBA
		return result
	}

	if matches := releaseQuarterPattern.FindStringSubmatch(raw); matches != nil {
		quarter, _ := strconv.Atoi(matches[1])
		year, _ := strconv.Atoi(matches[2])
		month := time.Month((quarter-1)*3 + 1)
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 3, -1)
		result.Precision = ReleaseDatePrecisionQuarter
		result.Year = year
		result.Quarter = quarter
		result.RangeStart = &start
		result.RangeEnd = &end
		return result
	}

	if releaseYearPattern.MatchString(raw) {
		year, _ := strconv.Atoi(raw)
		start := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, time.December, 31, 0, 0, 0, 0, time.UTC)
		result.Precision = ReleaseDatePrecisionYear
		result.Year = year
		result.RangeStart = &start
		result.RangeEnd = &end
		return result
	}

	for _, layout := range []string{
		"2 Jan, 2006",
		"Jan 2, 2006",
		"2 January, 2006",
		"January 2, 2006",
		"2006-01-02",
		"2006.01.02",
	} {
		if parsed, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			date := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
			result.Precision = ReleaseDatePrecisionDay
			result.ExactDate = &date
			result.Year = date.Year()
			result.Month = int(date.Month())
			result.RangeStart = &date
			result.RangeEnd = &date
			return result
		}
	}

	for _, layout := range []string{"January 2006", "Jan 2006"} {
		if parsed, err := time.ParseInLocation(layout, raw, time.UTC); err == nil {
			start := time.Date(parsed.Year(), parsed.Month(), 1, 0, 0, 0, 0, time.UTC)
			end := start.AddDate(0, 1, -1)
			result.Precision = ReleaseDatePrecisionMonth
			result.Year = start.Year()
			result.Month = int(start.Month())
			result.RangeStart = &start
			result.RangeEnd = &end
			return result
		}
	}

	result.Precision = ReleaseDatePrecisionUnknown
	return result
}
