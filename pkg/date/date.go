package date

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Format(t time.Time) string {
	return t.Format("2006-01-02")
}

func Parse(s string) (*time.Time, error) {
	if strings.Contains(s, "T") {
		s = s[:strings.Index(s, "T")]
	}

	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid date: %s", s)
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid year: %s", parts[0])
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid month: %s", parts[1])
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid day: %s", parts[2])
	}

	t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	return &t, nil

}
