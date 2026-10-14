package ui

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

var (
	errBadClock    = errors.New("time must look like HH:MM")
	errMissingDate = errors.New("date is missing")
)

var clockPattern = regexp.MustCompile(`^([01]?[0-9]|2[0-3]):([0-5][0-9])$`)

// newIntervalSpan is the default length of a manually added interval: the
// form proposes the hour that just ended.
const newIntervalSpan = time.Hour

// intervalForm holds the date and time fields of one interval. A running
// interval has no end, so its end fields stay empty and disabled.
type intervalForm struct {
	startDate  *widget.DateEntry
	startClock *widget.Entry
	endDate    *widget.DateEntry
	endClock   *widget.Entry
	errorLabel *widget.Label
	running    bool
	content    fyne.CanvasObject
}

func newIntervalForm(interval tracking.Interval, now time.Time) *intervalForm {
	start, end := proposedBounds(interval, now)
	form := &intervalForm{
		startDate:  dateEntry(start),
		startClock: clockEntry(start),
		endDate:    widget.NewDateEntry(),
		endClock:   widget.NewEntry(),
		errorLabel: widget.NewLabel(""),
		running:    interval.ID != 0 && interval.IsRunning(),
	}
	if form.running {
		form.endDate.Disable()
		form.endClock.Disable()
	} else {
		form.endDate.SetDate(&end)
		form.endClock.SetText(formatClock(end))
	}
	form.errorLabel.Importance = widget.DangerImportance
	form.errorLabel.Wrapping = fyne.TextWrapWord
	form.errorLabel.Hide()

	fields := widget.NewForm(
		widget.NewFormItem(message{"interval_form.start_date", "Start date"}.String(), form.startDate),
		widget.NewFormItem(message{"interval_form.start_time", "Start time"}.String(), form.startClock),
		widget.NewFormItem(message{"interval_form.end_date", "End date"}.String(), form.endDate),
		widget.NewFormItem(message{"interval_form.end_time", "End time"}.String(), form.endClock),
	)
	form.content = container.NewVBox(fields, form.errorLabel)
	return form
}

// moments returns the bounds typed by the user. original supplies the exact
// moments for fields the user did not change.
func (form *intervalForm) moments(original tracking.Interval) (start, end time.Time, err error) {
	start, err = composeMoment(form.startDate.Date, form.startClock.Text, original.StartedAt, time.Local)
	if err != nil || form.running {
		return start, time.Time{}, err
	}
	end, err = composeMoment(form.endDate.Date, form.endClock.Text, original.StoppedAt, time.Local)
	return start, end, err
}

func (form *intervalForm) showError(err error) {
	form.errorLabel.SetText(errorText(err))
	form.errorLabel.Show()
}

// composeMoment builds a moment from a picked date and an HH:MM text. If the
// result is the minute that original falls into, original is returned
// unchanged, so that untouched fields keep their seconds.
func composeMoment(date *time.Time, clockText string, original time.Time, location *time.Location) (time.Time, error) {
	if date == nil {
		return time.Time{}, errMissingDate
	}
	matches := clockPattern.FindStringSubmatch(strings.TrimSpace(clockText))
	if matches == nil {
		return time.Time{}, errBadClock
	}
	hour, _ := strconv.Atoi(matches[1])
	minute, _ := strconv.Atoi(matches[2])

	year, month, day := date.Date()
	moment := time.Date(year, month, day, hour, minute, 0, 0, location)
	if !original.IsZero() && original.Truncate(time.Minute).Equal(moment) {
		return original, nil
	}
	return moment, nil
}

func proposedBounds(interval tracking.Interval, now time.Time) (start, end time.Time) {
	if interval.ID == 0 {
		end = now.Truncate(time.Minute)
		return end.Add(-newIntervalSpan), end
	}
	return interval.StartedAt, interval.EndAt(now)
}

func dateEntry(moment time.Time) *widget.DateEntry {
	entry := widget.NewDateEntry()
	localMoment := moment.In(time.Local)
	entry.SetDate(&localMoment)
	return entry
}

func clockEntry(moment time.Time) *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder("HH:MM")
	entry.SetText(formatClock(moment.In(time.Local)))
	return entry
}
