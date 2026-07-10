package main

type Event struct {
	ID     EventID        `json:"id"`
	Epoch  EpochID        `json:"epoch"`
	Kind   string         `json:"kind"`
	Fields map[string]any `json:"fields"`
}

type Journal struct {
	next   EventID
	events []Event
}

func NewJournal() *Journal {
	return &Journal{next: 1, events: make([]Event, 0)}
}

func (journal *Journal) Append(epoch EpochID, kind string, fields map[string]any) Event {
	if fields == nil {
		fields = map[string]any{}
	}
	event := Event{
		ID:     journal.next,
		Epoch:  epoch,
		Kind:   kind,
		Fields: fields,
	}
	journal.next++
	journal.events = append(journal.events, event)
	return event
}

func (journal *Journal) Events() []Event {
	out := make([]Event, len(journal.events))
	copy(out, journal.events)
	return out
}

func (journal *Journal) CountKind(kind string) int {
	count := 0
	for _, event := range journal.events {
		if event.Kind == kind {
			count++
		}
	}
	return count
}

func (journal *Journal) Last() (Event, bool) {
	if len(journal.events) == 0 {
		return Event{}, false
	}
	return journal.events[len(journal.events)-1], true
}

func eventFields(pairs ...any) map[string]any {
	out := make(map[string]any)
	for i := 0; i+1 < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			continue
		}
		out[key] = pairs[i+1]
	}
	return out
}
