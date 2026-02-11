package audit

import (
	"sync"
	"time"
)

// Event is a single audited gateway action.
type Event struct {
	Timestamp     time.Time
	RequestID     string
	Team          string
	Model         string
	Status        string
	DenyReason    string
	RedactedInput string
	CostUSD       float64
	LatencyMS     int64
}

// Store is an in-memory bounded audit log.
type Store struct {
	mu        sync.Mutex
	maxEvents int
	events    []Event
}

func NewStore(maxEvents int) *Store {
	if maxEvents <= 0 {
		maxEvents = 1000
	}
	return &Store{maxEvents: maxEvents, events: make([]Event, 0, maxEvents)}
}

func (s *Store) Add(ev Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = append(s.events, ev)
	if len(s.events) > s.maxEvents {
		start := len(s.events) - s.maxEvents
		s.events = append([]Event(nil), s.events[start:]...)
	}
}

func (s *Store) List(team string, limit int) []Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	out := make([]Event, 0, limit)
	for i := len(s.events) - 1; i >= 0 && len(out) < limit; i-- {
		ev := s.events[i]
		if team != "" && ev.Team != team {
			continue
		}
		out = append(out, ev)
	}
	return out
}
