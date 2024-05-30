package main

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	itashi "git.aetherial.dev/aeth/itashi/pkg"
	tea "github.com/charmbracelet/bubbletea"
)

const ADD_TASK = "add"
const DEL_TASK = "del"
const DONE_TASK = "done"
const TIDY_TASK = "tidy"

//go:embed banner.txt
var banner []byte

func main() {
	args := os.Args
	if len(args) >= 2 {
		action := args[1]
		shelf := itashi.NewFilesystemShelf(itashi.GetDefualtSave())
		switch action {
		case ADD_TASK:
			itashi.AddTaskPrompt(shelf)
			os.Exit(0)
		case DEL_TASK:
			id, err := strconv.Atoi(args[2])
			if err != nil {
				log.Fatal("ID passed was not a valid integer: ", err)
			}
			shelf.DeleteTask(id)
			os.Exit(0)
		case DONE_TASK:
			var taskName string
			id, err := strconv.Atoi(args[2])
			if err != nil {
				log.Fatal("ID passed was not a valid integer: ", err)
			}
			taskName = shelf.MarkDone(id)
			if taskName == "" {
				fmt.Printf("No task was indexed with ID: %v\n", id)
				os.Exit(0)
			}
			fmt.Printf("よくできた! Good job! Task '%s' was marked complete!\n", strings.TrimSuffix(taskName, "\n"))
			os.Exit(0)
		case TIDY_TASK:
			fmt.Printf("Shelf tidied, removed %v completed tasks.\n", shelf.Clean())
			os.Exit(0)

		}
	}
	fmt.Printf("%+v\n", string(banner))

	p := tea.NewProgram(itashi.InitialModel())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
