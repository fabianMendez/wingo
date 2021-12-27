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

func Parse(s string) (t time.Time, err error) {
	if strings.Contains(s, "T") {
		s = s[:strings.Index(s, "T")]
	}

	parts := strings.Split(s, "-")
	if len(parts) != 3 {
		err = fmt.Errorf("invalid date: %s", s)
		return
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		err = fmt.Errorf("invalid year: %s", parts[0])
		return
	}
	month, err := strconv.Atoi(parts[1])
	if err != nil {
		err = fmt.Errorf("invalid month: %s", parts[1])
		return
	}
	day, err := strconv.Atoi(parts[2])
	if err != nil {
		err = fmt.Errorf("invalid day: %s", parts[2])
		return
	}

	t = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	return
}

func MustParse(s string) time.Time {
	t, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}
