// localizing all of the functions required to construct the user interface

package itashi

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const SPRING_EQUINOX = 81
const SUMMER_SOLSTICE = 173
const AUTUMN_EQUINOX = 265
const WINTER_SOLSTICE = 356

var Quarters = []int{
	SPRING_EQUINOX,
	SUMMER_SOLSTICE,
	AUTUMN_EQUINOX,
	WINTER_SOLSTICE,
}

const HEADER_TEMPLATE = `
{{.Date}}                                 {{.Season}}, {{.DaysToQuarter}} days until the next {{.QuarterType}}.
{{.DayOfWeek}}
{{.Time}} {{.Meridiem}} ({{.TtEod.Hours}}H, {{.TtEod.Minutes}}M -> EoD, {{.TtSun.Hours}}H, {{.TtSun.Minutes}}M -> {{.SunCycle}})
`

const TASK_ITEM = `
+------------------------------------
|  Task ID: {{.Id}} 
|  Title: {{.Title}}|
|  {{.Desc}}|
|  Due: {{.Due}}
|  Priority: {{.Priority}}
|  Done: {{.Done}}
|
+---------------------------
`

const TIME_TO_TEMPLATE = `{{.Hours}}H, {{.Minutes}}M`

// TODO: put all templates in their own file

type HeaderData struct {
	Date          string
	Season        string
	DaysToQuarter int
	QuarterType   string
	DayOfWeek     string
	Time          string
	Meridiem      string
	TtEod         TimeToSunShift
	TtSun         TimeToSunShift
	SunCycle      string
}

type TimeToSunShift struct {
	Hours   int
	Minutes int
}

type UserDetails interface {
	daysToQuarter(day int) int
	getQuarterType(day int) string
	getSeason(day int) string
	getTime(ts time.Time) string
	getDate(ts time.Time) string
	getMeridiem(ts time.Time) string
	getTimeToEod(ts time.Time) TimeToSunShift
	getTimeToSunShift(ts time.Time) TimeToSunShift
	getSunCycle(ts time.Time) string
}

type UserImplementation struct{}

func (u UserImplementation) daysToQuarter(day int) int {
	season := u.getSeason(day)
	if season == "Spring" {
		return SUMMER_SOLSTICE - day
	}

	return 1

}

/*
Return the quarter (solstice/equinox). We have to remember that we are returning
the NEXT season type, i.e. if its currently spring, the next quarter (summer) will have a solstice

	    :param day: the numerical day of the year
		:returns: either solstice, or equinox
*/
func (u UserImplementation) getQuarterType(day int) string {
	season := u.getSeason(day)
	if season == "Winter" {
		return "Equinox"
	}
	if season == "Summer" {
		return "Equinox"
	}
	return "Solstice"
}
func (u UserImplementation) getSeason(day int) string {
	if day > 365 {
		return "[REDACTED]"
	}
	if day < 0 {
		return "[REDACTED]"
	}
	if day > 0 && day < SPRING_EQUINOX {
		return "Winter"
	}
	if day > SPRING_EQUINOX && day < SUMMER_SOLSTICE {
		return "Spring"
	}
	if day > SUMMER_SOLSTICE && day < AUTUMN_EQUINOX {
		return "Summer"
	}
	if day > AUTUMN_EQUINOX && day < WINTER_SOLSTICE {
		return "Autumn"
	}
	if day > WINTER_SOLSTICE && day < 365 {
		return "Winter"
	}
	return "idk bruh"

}
func (u UserImplementation) getMeridiem(ts time.Time) string {
	if ts.Hour() < 12 {
		return "AM"
	}
	if ts.Hour() >= 12 {
		return "PM"
	}
	return "idk bruh"
}
func (u UserImplementation) getTimeToEod(ts time.Time) TimeToSunShift {
	if ts.Hour() > 17 {
		return TimeToSunShift{Hours: 0, Minutes: 0}
	}
	out := time.Date(ts.Year(), ts.Month(), ts.Day(), 17, 0, ts.Second(), ts.Nanosecond(), ts.Location())
	dur := time.Until(out)
	hours := dur.Minutes() / 60
	hours = math.Floor(hours)
	minutes := dur.Minutes() - (hours * 60)

	return TimeToSunShift{Hours: int(hours), Minutes: int(minutes)}
}

func (u UserImplementation) getTimeToSunShift(ts time.Time) TimeToSunShift {
	return TimeToSunShift{}
}

func (u UserImplementation) getSunCycle(ts time.Time) string {
	return "☼"
}
func (u UserImplementation) getTime(ts time.Time) string {
	var hour int
	if ts.Hour() == 0 {
		hour = 12
	} else {
		hour = ts.Hour()
	}
	return fmt.Sprintf("%v:%v", hour, ts.Minute())
}
func (u UserImplementation) getDate(ts time.Time) string {
	return fmt.Sprintf("%s %v, %v", ts.Month().String(), ts.Day(), ts.Year())
}

// Format the header string with a template
func getHeader(ud UserDetails, tmpl *template.Template) string {
	rn := time.Now()
	header := HeaderData{
		Date:          ud.getDate(rn),
		Season:        ud.getSeason(rn.YearDay()),
		DaysToQuarter: ud.daysToQuarter(rn.YearDay()),
		QuarterType:   ud.getQuarterType(rn.YearDay()),
		DayOfWeek:     rn.Weekday().String(),
		Time:          ud.getTime(rn),
		Meridiem:      ud.getMeridiem(rn),
		TtEod:         ud.getTimeToEod(rn),
		TtSun:         ud.getTimeToSunShift(rn),
		SunCycle:      ud.getSunCycle(rn),
	}
	var bw bytes.Buffer
	err := tmpl.Execute(&bw, header)
	if err != nil {
		log.Fatal("There was an issue parsing the header. sorry, ", err)
	}
	return bw.String()

}

type model struct {
	choices  []string
	cursor   int
	selected map[int]struct{}
}

func InitialModel() model {
	shelf := NewFilesystemShelf(GetDefualtSave())
	return model{
		choices:  GetTaskNames(shelf.GetAll()),
		selected: make(map[int]struct{}),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			fmt.Print("\033[H\033[2J")
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	// The header
	tmpl, err := template.New("header").Parse(HEADER_TEMPLATE)
	if err != nil {
		log.Fatal("Couldnt parse the header template.. sorry. ", err)
	}
	shelf := NewFilesystemShelf(GetDefualtSave())

	s := getHeader(UserImplementation{}, tmpl)

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Is this choice selected?
		var taskrender string
		checked := " " // not selected
		if _, ok := m.selected[i]; ok {
			for x := range shelf.Tasks {
				if shelf.Tasks[x].Title == choice {
					taskrender = shelf.RenderTask(shelf.Tasks[x])
				}
			}
			checked = "x" // selected!

		}

		// Render the row
		s += fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice)
		s += taskrender
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

/*
Add task to the shelf
*/
func AddTaskPrompt(shelf TaskShelf) {
	var title string
	var desc string
	var priority string
	var due time.Time

	var reader *bufio.Reader
	reader = bufio.NewReader(os.Stdout)
	fmt.Print("Enter Task Title: ")
	title, _ = reader.ReadString('\n')
	fmt.Print("Task description: ")
	desc, _ = reader.ReadString('\n')
	fmt.Print("Priority: ")
	priority, _ = reader.ReadString('\n')
	priorityInt, err := strconv.Atoi(strings.TrimSpace(priority))
	if err != nil {
		fmt.Print("We couldnt parse the priority value given. :(\n")
	}

	shelf.AddTask(title, desc, priorityInt, due)
}
