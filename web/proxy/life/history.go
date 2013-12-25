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
	e.Time = time.Now()
	e.String = s

	return e
}

type History struct {
	events []HistoryEvent
}

func NewHistory() *History {
	h := History{}
	h.events = make([]HistoryEvent, 0, 100)
	return &h
}

func (h *History) Log(e HistoryEvent) {
	h.events = append(h.events, e)
}

func (h *History) Format() string {
	r := ""
	for _, e := range h.events {
		r += e.Time.Format("2006-01-02 15:04:05") + " " + e.String + "\n"
	}

	return r
}
