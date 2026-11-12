package taskpolicy

const CalendarScope = "read:calendar"

const (
	Start       = "start"
	End         = "end"
	Title       = "title"
	Location    = "location"
	Attendees   = "attendees"
	Description = "description"
)

type Task struct {
	ID                string
	Name              string
	AllowedAttributes []string
}

var catalog = map[string]Task{
	"calendar.freebusy": {
		ID:                "calendar.freebusy",
		Name:              "Find free time",
		AllowedAttributes: []string{Start, End},
	},
	"calendar.reminder": {
		ID:                "calendar.reminder",
		Name:              "Create reminders",
		AllowedAttributes: []string{Start, End, Title},
	},
}

var allAttributes = []string{
	Start,
	End,
	Title,
	Location,
	Attendees,
	Description,
}

func Lookup(taskID string) (Task, bool) {
	task, ok := catalog[taskID]
	if !ok {
		return Task{}, false
	}
	task.AllowedAttributes = append([]string(nil), task.AllowedAttributes...)
	return task, true
}

func AllAttributes() []string {
	return append([]string(nil), allAttributes...)
}
