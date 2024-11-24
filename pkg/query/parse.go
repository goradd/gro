package query

import (
	"strconv"
	"strings"
	"time"
)

// ParseTime will convert a SQL Date, Time, DateTime, Timestamp or RFC3339 string to a time.Time.
// Various databases express their times differently, and this tries to interpret what is
// attempting to be expressed. It can handle unix time strings that are +- from
// the 1970 epoch, including fractional times up to the microsecond level.
//
// If the SQL date time string does not have timezone information,
// the resulting value will be in UTC time.
// If an error occurs, the returned value will be the zero date.
func ParseTime(s string) (t time.Time) {
	var form string

	if len(s) < 4 {
		return // must at least have some minimal amount of data to start
	}

	// First check for a unix time
	if _, e := strconv.ParseFloat(s, 64); e == nil {
		parts := strings.Split(s, ".")
		var i, f int64
		i, _ = strconv.ParseInt(parts[0], 10, 64)
		if len(parts) > 1 && parts[1] != "" {
			f, _ = strconv.ParseInt(parts[1], 10, 64)
			f = f * pow10(9-len(parts[1]))
		}
		t = time.Unix(i, f).UTC()
		return
	}

	if len(s) > 10 && s[10] == 'T' {
		form = time.RFC3339
	} else {
		var hasDate, hasTime, hasTZ, hasLocale bool
		if strings.Contains(s, "-") {
			hasDate = true
		}
		if strings.Contains(s, ":") {
			hasTime = true

			if strings.LastIndexAny(s, "+-") > strings.LastIndex(s, ":") {
				hasTZ = true
				lastChar := s[len(s)-1]

				if lastChar == 'T' || lastChar == 'C' {
					hasLocale = true
				}
			}
		}
		if hasDate {
			form = "2006-01-02"
			if hasTime {
				form += " 15:04:05"
				if hasTZ {
					form += " -0700"
					if hasLocale {
						form += " MST"
					}
				}
			}
		} else {
			form = "15:04:05"
		}
	}
	t, err := time.Parse(form, s)
	if err == nil {
		t = t.UTC()
	} else {
		z := time.Time{}
		t = z // make sure we return a zero time
	}
	return
}

func pow10(exp int) int64 {
	if exp == 0 {
		return 1
	}
	v := int64(10)
	for i := 1; i < exp; i++ {
		v = v * 10
	}
	return v
}
