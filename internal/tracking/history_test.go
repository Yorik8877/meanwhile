package tracking_test

import (
	"testing"
	"time"
	_ "time/tzdata"

	"github.com/yorik8877/meanwhile/internal/tracking"
)

func TestDailyTotalsSplitIntervalsAtLocalMidnight(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	id := fixture.createTask(t, "Night shift")
	start := time.Date(2026, time.October, 13, 23, 0, 0, 0, time.UTC)
	mustSucceed(t, fixture.tracker.AddInterval(id, start, start.Add(2*time.Hour)))

	totals := fixture.tracker.DailyTotals(tracking.Date{Year: 2026, Month: time.October, Day: 12}, 7)

	if len(totals) != 2 {
		t.Fatalf("days = %+v, want two days", totals)
	}
	if totals[0].Date.Day != 14 || totals[0].Total != time.Hour {
		t.Errorf("first day = %+v, want 14th with 1h", totals[0])
	}
	if totals[1].Date.Day != 13 || totals[1].Total != time.Hour {
		t.Errorf("second day = %+v, want 13th with 1h", totals[1])
	}
}

func TestDailyTotalsOnDaylightSavingDays(t *testing.T) {
	location := berlin(t)
	fixture := newFixture(t, location)
	id := fixture.createTask(t, "Long run")

	springStart := time.Date(2026, time.March, 28, 22, 0, 0, 0, location)
	springEnd := time.Date(2026, time.March, 30, 2, 0, 0, 0, location)
	mustSucceed(t, fixture.tracker.AddInterval(id, springStart, springEnd))
	autumnStart := time.Date(2026, time.October, 24, 22, 0, 0, 0, location)
	autumnEnd := time.Date(2026, time.October, 26, 2, 0, 0, 0, location)
	fixture.clock.moment = time.Date(2026, time.November, 1, 0, 0, 0, 0, location)
	mustSucceed(t, fixture.tracker.AddInterval(id, autumnStart, autumnEnd))

	want := map[tracking.Date]time.Duration{
		{Year: 2026, Month: time.March, Day: 29}:   23 * time.Hour,
		{Year: 2026, Month: time.October, Day: 25}: 25 * time.Hour,
	}
	for date, wantDuration := range want {
		totals := fixture.tracker.DailyTotals(date, 1)
		if len(totals) != 1 || totals[0].Total != wantDuration {
			t.Errorf("%v: totals = %+v, want %v", date, totals, wantDuration)
		}
	}
}

func TestDailyTotalsOrderTasksByDuration(t *testing.T) {
	fixture := newFixture(t, time.UTC)
	short := fixture.createTask(t, "Short")
	long := fixture.createTask(t, "Long")
	morning := time.Date(2026, time.October, 14, 8, 0, 0, 0, time.UTC)
	mustSucceed(t, fixture.tracker.AddInterval(short, morning, morning.Add(10*time.Minute)))
	mustSucceed(t, fixture.tracker.AddInterval(long, morning, morning.Add(time.Hour)))

	totals := fixture.tracker.DailyTotals(tracking.Date{Year: 2026, Month: time.October, Day: 14}, 1)

	if len(totals) != 1 || totals[0].Total != 70*time.Minute {
		t.Fatalf("totals = %+v, want one day with 1h10m", totals)
	}
	if tasks := totals[0].Tasks; tasks[0].Task.ID != long || tasks[1].Task.ID != short {
		t.Errorf("order = %+v, want long then short", tasks)
	}
}

func berlin(t *testing.T) *time.Location {
	t.Helper()
	location, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("load Europe/Berlin: %v", err)
	}
	return location
}
