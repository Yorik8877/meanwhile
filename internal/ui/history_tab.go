package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"stopwatch/internal/tracking"
)

const daysPerWeek = 7

type historyTab struct {
	tracker   *tracking.Tracker
	weekStart tracking.Date
	weekLabel *widget.Label
	nextWeek  *widget.Button
	days      *fyne.Container
	content   fyne.CanvasObject
}

func newHistoryTab(tracker *tracking.Tracker) *historyTab {
	tab := &historyTab{tracker: tracker, weekStart: currentWeekStart(), days: container.NewVBox()}

	previousWeek := widget.NewButton(message{"history.previous_week", "← Week"}.String(), func() { tab.moveWeek(-1) })
	tab.nextWeek = widget.NewButton(message{"history.next_week", "Week →"}.String(), func() { tab.moveWeek(1) })
	tab.weekLabel = widget.NewLabel("")
	tab.weekLabel.Alignment = fyne.TextAlignCenter
	hint := widget.NewLabel(message{"history.parallel_hint",
		"Tasks that ran in parallel are counted separately, so a day's total can exceed the time that passed."}.String())
	hint.Wrapping = fyne.TextWrapWord
	hint.Importance = widget.LowImportance

	navigation := container.NewBorder(nil, nil, previousWeek, tab.nextWeek, tab.weekLabel)
	readableHint := container.NewThemeOverride(hint, contrastTheme{})
	tab.content = container.NewBorder(container.NewVBox(navigation, readableHint), nil, nil, nil, container.NewVScroll(tab.days))
	tab.refresh()
	return tab
}

func (tab *historyTab) moveWeek(weeks int) {
	tab.weekStart = tab.weekStart.AddDays(weeks * daysPerWeek)
	tab.refresh()
}

func (tab *historyTab) refresh() {
	weekEnd := tab.weekStart.AddDays(daysPerWeek - 1)
	tab.weekLabel.SetText(formatShortDate(tab.weekStart) + " – " + formatShortDate(weekEnd))
	if tab.weekStart == currentWeekStart() {
		tab.nextWeek.Disable()
	} else {
		tab.nextWeek.Enable()
	}

	tab.days.RemoveAll()
	totals := tab.tracker.DailyTotals(tab.weekStart, daysPerWeek)
	if len(totals) == 0 {
		tab.days.Add(widget.NewLabel(message{"history.empty", "No recorded time this week."}.String()))
		return
	}
	for _, day := range totals {
		tab.days.Add(dayView(day))
	}
}

func dayView(day tracking.DayTotal) fyne.CanvasObject {
	heading := widget.NewLabelWithStyle(formatDayHeading(day), fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	view := container.NewVBox(heading)
	for _, taskDuration := range day.Tasks {
		name := widget.NewLabel(taskDuration.Task.Name)
		name.Truncation = fyne.TextTruncateEllipsis
		duration := widget.NewLabel(formatDuration(taskDuration.Duration))
		view.Add(container.NewBorder(nil, nil, nil, duration, name))
	}
	return view
}

func currentWeekStart() tracking.Date {
	return tracking.DateOf(time.Now()).StartOfWeek()
}
