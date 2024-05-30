package itashi

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const SHELF_LINE_DELIM = "\n----    ----    ----    ----\n"
const SHELF_COL_DELIM = "    "
const TIME_FORMAT = "2006-01-02T15:04:05 -07:00:00"
const FS_SAVE_LOCATION = "./todo.ta"
const SHELF_TEMPLATE = "{{.Id}}    {{.Title}}    {{.Desc}}    {{.Due}}    {{.Done}}    {{.Priority}}"

type Task struct {
	Id       int
	Title    string
	Desc     string
	Due      time.Time
	Done     bool
	Priority int
}

func GetDefualtSave() string {
	return fmt.Sprintf("%s/.config/itashi/todo.ta", os.Getenv("HOME"))
}

// lets make Task implement the tea.Model interface
func (t Task) Init() tea.Cmd {
	return nil

}

func (t Task) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return t, nil
}

func (t Task) View() string {
	return ""
}

type TaskShelf interface {
	// Modify the due date of an existing task
	ModifyDue(id int, due time.Time)
	// Modify the description field of an existing task
	ModifyDesc(id int, desc string)
	// Modify the priority of an existing task
	ModifyPriority(id int, pri int)
	// modify the title of an existing task
	ModifyTitle(id int, title string)
	// delete an existing task from the shelf
	DeleteTask(id int)
	// Mark a task as complete
	MarkDone(id int) string
	// hopefully you dont need to call this! ;)
	ResetDone(id int)
	// Add a task to the shelf
	AddTask(title string, desc string, priority int, due time.Time)
	// Retrieve all tasks in the shelf
	GetAll() []Task
	// Render a task to a task template
	RenderTask(task Task) string
	// Clean the shelf of all completed tasks
	Clean() int
}

/*
Retrieve all the tasks from the designated TaskShelf
*/
func GetTaskList(taskio TaskShelf) []Task {
	return taskio.GetAll()
}

/*
Grab all of the names of the tasks from the TaskShelf

	    :param t: a list of Task structs
		:returns: A list of the task names
*/
func GetTaskNames(t []Task) []string {
	var taskn []string
	for i := range t {
		taskn = append(taskn, t[i].Title)
	}
	return taskn
}

type FilesystemShelf struct {
	SaveLocation string
	Template     *template.Template
	TaskTempl    *template.Template
	Tasks        []Task
}

func (t *FilesystemShelf) RenderTask(task Task) string {
	var bw bytes.Buffer
	err := t.TaskTempl.Execute(&bw, task)
	if err != nil {
		log.Fatal("Had a problem rendering this task.", task, err)
	}
	return bw.String()
}

/*
Create a new filesystem shelf struct to reflect the filesystem shelf

	    :param save: the save location to store the shelf in
		:returns: a pointer to a FilesystemShelf struct
*/
func NewFilesystemShelf(save string) *FilesystemShelf {
	tmpl, err := template.New("task").Parse(SHELF_TEMPLATE)
	if err != nil {
		log.Fatal("Could not parse the shelf template! ", err)
	}
	tasktmpl, err := template.New("task").Parse(TASK_ITEM)
	if err != nil {
		log.Fatal("Couldnt parse task template. ", err)
	}

	shelf := &FilesystemShelf{
		SaveLocation: save,
		Template:     tmpl,
		TaskTempl:    tasktmpl,
		Tasks:        []Task{},
	}
	shelf.Tasks = shelf.GetAll()
	return shelf

}

// Retrieve all the tasks from the filesystem shelf
func (f *FilesystemShelf) GetAll() []Task {
	b, err := os.ReadFile(f.SaveLocation)
	if err != nil {
		log.Fatal(err)
	}
	return parseFilesystemShelf(b)
}

/*
Add a task to the filesystem shelf

	    :param title: the title to give the task
		:param desc: the description to give the task
		:param priority: the priority to give the task
		:param due: the due date for the task
		:returns: Nothing
*/
func (f *FilesystemShelf) AddTask(title string, desc string, priority int, due time.Time) {
	var inc int
	inc = 0
	for i := range f.Tasks {
		if f.Tasks[i].Id > inc {
			inc = f.Tasks[i].Id
		}
	}
	inc += 1
	task := Task{
		Id:       inc,
		Title:    title,
		Desc:     desc,
		Due:      due,
		Done:     false,
		Priority: priority,
	}
	f.Tasks = append(f.Tasks, task)

	err := os.WriteFile(f.SaveLocation, marshalTaskToShelf(f.Tasks, f.Template), os.ModePerm)
	if err != nil {
		log.Fatal("Need to fix later, error writing to fs ", err)
	}
}

// Boiler plate so i can implement later
func (f *FilesystemShelf) ModifyDue(id int, due time.Time)  {}
func (f *FilesystemShelf) ModifyDesc(id int, desc string)   {}
func (f *FilesystemShelf) ModifyPriority(id int, pri int)   {}
func (f *FilesystemShelf) ModifyTitle(id int, title string) {}
func (f *FilesystemShelf) DeleteTask(id int) {
	replTasks := []Task{}
	for i := range f.Tasks {
		if f.Tasks[i].Id == id {
			continue
		}
		replTasks = append(replTasks, f.Tasks[i])
	}
	os.WriteFile(f.SaveLocation, marshalTaskToShelf(replTasks, f.Template), os.ModePerm)

}

// Clean the filesystem shelf of all completed tasks
func (f *FilesystemShelf) Clean() int {
	replTasks := []Task{}
	var cleaned int
	cleaned = 0
	for i := range f.Tasks {
		if f.Tasks[i].Done {
			cleaned += 1
			continue
		}
		replTasks = append(replTasks, f.Tasks[i])
	}
	os.WriteFile(f.SaveLocation, marshalTaskToShelf(replTasks, f.Template), os.ModePerm)
	return cleaned
}

/*
Mark task as done and write the shelf to disk. since the Tasks within FilesystemShelf are
values and not pointers, we need to copy the entirety of the shelf over to a new set
and write it, as opposed to just modifying the pointer and then writing.

	    :param id: the ID of the task to mark as done
		:returns: Nothing
*/
func (f *FilesystemShelf) MarkDone(id int) string {
	replTasks := []Task{}
	var taskName string
	for i := range f.Tasks {
		if f.Tasks[i].Id == id {
			replTasks = append(replTasks, Task{
				Id:       f.Tasks[i].Id,
				Title:    f.Tasks[i].Title,
				Desc:     f.Tasks[i].Desc,
				Due:      f.Tasks[i].Due,
				Done:     true,
				Priority: f.Tasks[i].Priority,
			})
			taskName = f.Tasks[i].Title
			continue
		}
		replTasks = append(replTasks, f.Tasks[i])
	}
	os.WriteFile(f.SaveLocation, marshalTaskToShelf(replTasks, f.Template), os.ModePerm)
	return taskName
}

func (f *FilesystemShelf) ResetDone(id int) {}

// private function for parsing the byte stream from the filesystem
func parseFilesystemShelf(data []byte) []Task {
	var filestring string
	filestring = string(data)
	items := strings.Split(filestring, SHELF_LINE_DELIM)
	var shelf []Task
	for i := range items {
		sect := strings.Split(items[i], SHELF_COL_DELIM)
		if len(sect) < 6 {
			continue
		}
		var id int
		var due time.Time
		var done bool
		var pri int
		var err error
		id, err = strconv.Atoi(sect[0])
		due, err = time.Parse(TIME_FORMAT, sect[3])
		done, err = strconv.ParseBool(sect[4])
		pri, err = strconv.Atoi(sect[5])
		if err != nil {
			log.Fatal("Couldnt parse from filesystem shelf", err)
		}

		shelf = append(shelf, Task{
			Id:       id,
			Title:    sect[1],
			Desc:     sect[2],
			Due:      due,
			Done:     done,
			Priority: pri,
		})

	}
	return shelf

}

/*
Helper function to marshal the tasks to the custom format

	     :param tasks: an array of Task structs
		 :returns: a byte array
*/
func marshalTaskToShelf(tasks []Task, templ *template.Template) []byte {
	var bw bytes.Buffer
	for i := range tasks {
		err := templ.Execute(&bw, tasks[i])
		if err != nil {
			log.Fatal("Error parsing data into template: ", err)
		}
		// dynamically allocating, no need for the read delim
		_, err = bw.Write([]byte(SHELF_LINE_DELIM))
		if err != nil {
			log.Fatal("Error parsing data into template: ", err)
		}

	}
	return bw.Bytes()
}
