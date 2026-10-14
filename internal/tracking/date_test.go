package tracking_test

import (
	"testing"
	"time"

	"stopwatch/internal/tracking"
)

func TestStartOfWeekIsMonday(t *testing.T) {
	cases := map[tracking.Date]tracking.Date{
		{Year: 2026, Month: time.October, Day: 14}: {Year: 2026, Month: time.October, Day: 12},
		{Year: 2026, Month: time.October, Day: 12}: {Year: 2026, Month: time.October, Day: 12},
		{Year: 2026, Month: time.October, Day: 18}: {Year: 2026, Month: time.October, Day: 12},
		{Year: 2026, Month: time.January, Day: 1}:  {Year: 2025, Month: time.December, Day: 29},
	}
	for date, wantMonday := range cases {
		if monday := date.StartOfWeek(); monday != wantMonday {
			t.Errorf("StartOfWeek(%v) = %v, want %v", date, monday, wantMonday)
		}
	}
}
