package tracking

import "time"

// Date is a calendar day without a time of day or a time zone.
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

// DateOf returns the calendar day of the moment in the moment's own location.
func DateOf(moment time.Time) Date {
	year, month, day := moment.Date()
	return Date{Year: year, Month: month, Day: day}
}

// Start returns local midnight. time.Date handles days that are 23 or 25
// hours long because of daylight saving transitions.
func (date Date) Start(location *time.Location) time.Time {
	return time.Date(date.Year, date.Month, date.Day, 0, 0, 0, 0, location)
}

func (date Date) AddDays(days int) Date {
	return DateOf(date.noonUTC().AddDate(0, 0, days))
}

func (date Date) Weekday() time.Weekday {
	return date.noonUTC().Weekday()
}

// StartOfWeek returns the Monday of the week that contains the date.
func (date Date) StartOfWeek() Date {
	daysSinceMonday := (int(date.Weekday()) + 6) % 7
	return date.AddDays(-daysSinceMonday)
}

func (date Date) noonUTC() time.Time {
	return time.Date(date.Year, date.Month, date.Day, 12, 0, 0, 0, time.UTC)
}
