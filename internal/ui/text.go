// Package ui is the Fyne interface: the main window with its tabs, the
// interval editor and the tray icon with its menu.
package ui

import (
	"embed"
	"errors"
	"fmt"
	"time"

	"fyne.io/fyne/v2/lang"

	"stopwatch/internal/tracking"
)

//go:embed translations
var translations embed.FS

// LoadTranslations must run before any window is created. The language is
// picked from the system locale; English is the fallback.
func LoadTranslations() error {
	return lang.AddTranslationsFS(translations, "translations")
}

type message struct {
	key      string
	fallback string
}

func (text message) String() string {
	return lang.X(text.key, text.fallback)
}

func (text message) With(data map[string]any) string {
	return lang.X(text.key, text.fallback, data)
}

var errorMessages = []struct {
	target error
	text   message
}{
	{tracking.ErrEmptyName, message{"error.empty_name", "Enter a task name."}},
	{tracking.ErrTaskNotFound, message{"error.task_not_found", "The task no longer exists."}},
	{tracking.ErrIntervalNotFound, message{"error.interval_not_found", "The interval no longer exists."}},
	{tracking.ErrMissingEnd, message{"error.missing_end", "Enter the end of the interval."}},
	{tracking.ErrEndBeforeStart, message{"error.end_before_start", "The end must be later than the start."}},
	{tracking.ErrIntervalInFuture, message{"error.in_future", "The interval cannot end in the future."}},
	{tracking.ErrIntervalsOverlap, message{"error.overlap", "The interval overlaps another interval of this task."}},
	{tracking.ErrRunningIntervalEnd, message{"error.running_end", "A running interval has no end."}},
	{errBadClock, message{"interval_form.time_format", "Enter time as HH:MM."}},
	{errMissingDate, message{"interval_form.date_missing", "Pick a date."}},
}

// errorText turns domain errors into a translated sentence; other errors,
// such as a failed database write, keep their technical text.
func errorText(err error) string {
	for _, known := range errorMessages {
		if errors.Is(err, known.target) {
			return known.text.String()
		}
	}
	return message{"error.generic", "Could not save the change: {{.Error}}"}.With(map[string]any{"Error": err.Error()})
}

func formatDuration(duration time.Duration) string {
	duration = max(duration, 0)
	hours := duration / time.Hour
	minutes := duration % time.Hour / time.Minute
	seconds := duration % time.Minute / time.Second
	return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
}

func formatClock(moment time.Time) string {
	return moment.Format("15:04")
}

var weekdayNames = map[time.Weekday]message{
	time.Monday:    {"weekday.monday", "Mon"},
	time.Tuesday:   {"weekday.tuesday", "Tue"},
	time.Wednesday: {"weekday.wednesday", "Wed"},
	time.Thursday:  {"weekday.thursday", "Thu"},
	time.Friday:    {"weekday.friday", "Fri"},
	time.Saturday:  {"weekday.saturday", "Sat"},
	time.Sunday:    {"weekday.sunday", "Sun"},
}

var monthNames = map[time.Month]message{
	time.January:   {"month.january", "Jan"},
	time.February:  {"month.february", "Feb"},
	time.March:     {"month.march", "Mar"},
	time.April:     {"month.april", "Apr"},
	time.May:       {"month.may", "May"},
	time.June:      {"month.june", "Jun"},
	time.July:      {"month.july", "Jul"},
	time.August:    {"month.august", "Aug"},
	time.September: {"month.september", "Sep"},
	time.October:   {"month.october", "Oct"},
	time.November:  {"month.november", "Nov"},
	time.December:  {"month.december", "Dec"},
}

// formatShortDate gives "Oct 13" in English and "13 окт" in Russian.
func formatShortDate(date tracking.Date) string {
	return message{"date.short", "{{.Month}} {{.Day}}"}.With(map[string]any{
		"Day":   date.Day,
		"Month": monthNames[date.Month].String(),
	})
}

func formatDayHeading(day tracking.DayTotal) string {
	return message{"history.day_heading", "{{.Weekday}}, {{.Date}} — {{.Duration}}"}.With(map[string]any{
		"Weekday":  weekdayNames[day.Date.Weekday()].String(),
		"Date":     formatShortDate(day.Date),
		"Duration": formatDuration(day.Total),
	})
}
