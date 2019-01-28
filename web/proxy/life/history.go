package life

import (
	"time"
)

type HistoryEvent struct {
	Time   time.Time
	String string
}

func StringHistory(s string) HistoryEvent {
	e := HistoryEvent{}
	e.Time = uniqueTime()
	e.String = s

	return e
}

type History struct {
	events []*HistoryEvent
	max    int
}

func NewHistory() *History {
	h := History{}
	h.max = 10000

	h.Clear()
	return &h
}

func (h *History) Log(e HistoryEvent) {
	h.events = append(h.events, &e)
	if len(h.events) >= h.max {
		h.events = h.events[len(h.events)/2:]
	}
}

func (h *History) Format() string {
	r := ""
	for _, e := range h.events {
		r += e.Time.Format("2006-01-02 15:04:05") + " " + e.String + "\n"
	}

	return r
}

func (h *History) Events() []*HistoryEvent {
	events := make([]*HistoryEvent, 0, len(h.events))
	for _, e := range h.events {
		events = append(events, e)
	}

	return events
}

func (h *History) EventsAfter(t time.Time) []*HistoryEvent {
	events := make([]*HistoryEvent, 0, len(h.events))
	for _, e := range h.events {
		if e.Time.After(t) {
			events = append(events, e)
		}
	}

	return events
}

func (h *History) Clear() {
	h.events = make([]*HistoryEvent, 0, 100)
}
