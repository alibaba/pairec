package utils

import "time"

//copy from https://github.com/Knetic/govaluate/blob/0580e9b47a69125afa0e4ebd1cf93c49eb5a43ec/parsing.go#L258

// TryParseTime tries to parse a string into a time.Time object.
func TryParseTime(candidate string) (time.Time, bool) {

	var ret time.Time
	var found bool

	timeFormats := [...]string{
		"2006-01-02 15:04:05", // RFC 3339 with seconds
		"2006-01-02 15:04",    // RFC 3339 with minutes
		"2006-01-02",          // RFC 3339
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02 15:04:05-07:00",          // RFC 3339 with seconds and timezone
		"2006-01-02T15Z0700",                 // ISO8601 with hour
		"2006-01-02T15:04Z0700",              // ISO8601 with minutes
		"2006-01-02T15:04:05Z0700",           // ISO8601 with seconds
		"2006-01-02T15:04:05.999999999Z0700", // ISO8601 with nanoseconds
		time.ANSIC,
		time.UnixDate,
		time.RubyDate,
		time.Kitchen,
	}

	for _, format := range timeFormats {

		ret, found = tryParseExactTime(candidate, format)
		if found {
			return ret, true
		}
	}

	return time.Now(), false
}

func tryParseExactTime(candidate string, format string) (time.Time, bool) {

	var ret time.Time
	var err error

	ret, err = time.Parse(format, candidate)
	if err != nil {
		return time.Now(), false
	}

	return ret, true
}
