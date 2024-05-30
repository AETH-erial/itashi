package itashi

import "text/template"

type ShelfHome struct {
	Items     []Option
	Tasks     TaskShelf
	TaskTempl *template.Template
}

// Return a list of the options as strings for the UI to render
func (s ShelfHome) OptionList() []string {
	var optnames []string
	for i := range s.Items {
		optnames = append(optnames, s.Items[i].Name)
	}
	return optnames

}

type Option struct {
	Name     string             // the display name in the UI
	Template *template.Template // The template to render in the Render() func

}

// Render the template stored in the Template struct field
func (o Option) Render() string {
	return "This is a placeholder"
}

// Create the task shelf homepage
func GetShelfHome(save string) ShelfHome {
	return ShelfHome{
		Items: GetOptions(),
		Tasks: NewFilesystemShelf(save),
	}
}

// Removing this from GetShelfHome to allow for indirecting the data feed
func GetOptions() []Option {
	var opts []Option
	opts = append(opts, Option{Name: "Add task to your shelf"})
	opts = append(opts, Option{Name: "Edit Task"})
	opts = append(opts, Option{Name: "Move task to done pile"})
	opts = append(opts, Option{Name: "View my shelf"})
	return opts

}
