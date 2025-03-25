package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
)

const (
	file_search    = "file_search"
	selecting      = "selecting"
	show_incorrect = "showing_incorrect"
	show_correct   = "showing_correct"
)

type Question struct {
	Title   string   `json:"question"`
	Options []string `json:"options"`
	Correct int      `json:"correct"`
}

type TimeoutEvent struct {
	tcell.EventTime
}

type App struct {
	CurrentQuestion int
	Questions       []Question

	SelectedIndex int
	State         string
	WrongSelected map[int]struct{}

	Screen tcell.Screen

	QuestionsFrom   []QuestionFile
	QuestionsChosen map[int]struct{}
}

type QuestionFile struct {
	Path     string
	Filename string
}

// NOTE: recursive search for json questions file
func fileSearch(path string) ([]QuestionFile, error) {
	result := []QuestionFile{}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		// NOTE: skip hidden files
		if entry.Name()[0] == '.' {
			continue
		}

		// NOTE: if directory, search inside
		if entry.IsDir() {
			innerPath := filepath.Join(path, entry.Name())

			innerResult, err := fileSearch(innerPath)
			if err != nil {
				log.Printf("searching in directory %s: %+v", innerPath, err)
			}

			result = append(result, innerResult...)

			continue
		}

		// NOTE: only json files
		if !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		currentFilePath := filepath.Join(path, entry.Name())
		result = append(result, QuestionFile{
			Path:     currentFilePath,
			Filename: entry.Name(),
		})
	}

	return result, nil
}

func (a *App) parseQuestions(path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("opening questions file: %w", err)
	}

	questions := []Question{}

	err = json.Unmarshal(file, &questions)
	if err != nil {
		return fmt.Errorf("unable to decode file: %w", err)
	}

	a.Questions = append(a.Questions, questions...)

	return nil
}

func NewApp() (*App, error) {
	questionFiles, err := fileSearch("./")
	if err != nil || len(questionFiles) == 0 {
		return nil, fmt.Errorf("unable to search for question files: %w", err)
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, fmt.Errorf("unable to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return nil, fmt.Errorf("unable to initialize screen: %w", err)
	}

	return &App{
		CurrentQuestion: 0,
		Screen:          screen,
		SelectedIndex:   0,
		State:           file_search,
		Questions:       []Question{},
		WrongSelected:   make(map[int]struct{}),
		QuestionsFrom:   questionFiles,
		QuestionsChosen: make(map[int]struct{}),
	}, nil
}

func (a *App) postEvent(duration time.Duration) {
	time.AfterFunc(duration, func() {
		event := &TimeoutEvent{}
		event.SetEventNow()
		_ = a.Screen.PostEvent(event)
	})
}

func (a *App) drawQuestions() {
	a.Screen.Clear()

	row, col := 0, 0

	// NOTE: drawing title for the question
	titleStyle := tcell.StyleDefault.Bold(true)
	for _, r := range a.Questions[a.CurrentQuestion].Title {
		a.Screen.SetContent(col, row, r, nil, titleStyle)
		col++
	}

	// NOTE: title bigger gap
	row += 2

	for idx, option := range a.Questions[a.CurrentQuestion].Options {
		// NOTE: drawing checkbox
		checkbox := "\uf096"
		checkboxStyle := tcell.StyleDefault

		// NOTE: state: showing correct option
		if a.State == show_correct && idx == a.Questions[a.CurrentQuestion].Correct {
			checkbox = "\uf14a"
			checkboxStyle = checkboxStyle.Foreground(tcell.ColorGreen)
		}

		// NOTE: wrong selected option
		if _, ok := a.WrongSelected[idx]; ok {
			checkbox = "\uf2d3"
			checkboxStyle = checkboxStyle.Foreground(tcell.ColorRed)
		}

		col := 0
		for _, r := range checkbox {
			a.Screen.SetContent(col, row, r, nil, checkboxStyle)
			col++
		}

		// NOTE: space before text
		col++

		textStyle := tcell.StyleDefault
		if a.State == selecting && idx == a.SelectedIndex {
			textStyle = textStyle.Reverse(true)
		}

		if a.State == show_correct && idx == a.Questions[a.CurrentQuestion].Correct {
			textStyle = textStyle.Foreground(tcell.ColorGreen)
		}

		if _, ok := a.WrongSelected[idx]; ok {
			textStyle = textStyle.Foreground(tcell.ColorGray)
		}

		for _, r := range option {
			a.Screen.SetContent(col, row, r, nil, textStyle)
			col++
		}

		row++
	}

	col = 0
	row++

	var message string
	var textStyle tcell.Style

	switch a.State {
	case show_correct:
		message = "Your choice is correct"
		textStyle = tcell.StyleDefault.Foreground(tcell.ColorGreen).Bold(true)
	case show_incorrect:
		message = "Your choice is incorrect"
		textStyle = tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
	default:
		message = "Press `j`/`k` or `Up`/`Down` to navigate\n`Enter` to select"
		textStyle = tcell.StyleDefault
	}

	monoFlag := false

	for _, r := range message {
		if r == '`' {
			monoFlag = !monoFlag
			continue
		}

		if monoFlag {
			textStyle = textStyle.Italic(true)
		} else {
			textStyle = textStyle.Italic(false)
		}

		if r == '\n' {
			col = 0
			row++
			continue
		}

		a.Screen.SetContent(col, row, r, nil, textStyle)
		col++
	}

	a.Screen.Show()
}

func (a *App) drawConfiguraiton() {
	a.Screen.Clear()

	row, col := 0, 0

	// NOTE: drawing title for the question
	titleStyle := tcell.StyleDefault.Bold(true)
	message := "Select file(s) to load questions from"
	for _, r := range message {
		a.Screen.SetContent(col, row, r, nil, titleStyle)
		col++
	}

	// NOTE: title bigger gap
	row += 2

	for idx, option := range a.QuestionsFrom {
		// NOTE: drawing checkbox
		checkbox := "\uf096"
		checkboxStyle := tcell.StyleDefault

		// NOTE: state: showing correct option
		if _, ok := a.QuestionsChosen[idx]; ok {
			checkbox = "\uf14a"
			checkboxStyle = checkboxStyle.Foreground(tcell.ColorGreen)
		}

		col := 0
		for _, r := range checkbox {
			a.Screen.SetContent(col, row, r, nil, checkboxStyle)
			col++
		}

		// NOTE: space before text
		col++

		textStyle := tcell.StyleDefault
		if idx == a.SelectedIndex {
			textStyle = textStyle.Reverse(true)
		}

		for _, r := range option.Filename {
			a.Screen.SetContent(col, row, r, nil, textStyle)
			col++
		}

		row++
	}

	col = 0
	row++

	message = "Press `j`/`k` or `Up`/`Down` to navigate\n`Enter` to select\n`G` to start the quiz"
	style := tcell.StyleDefault
	monoFlag := false

	for _, r := range message {
		if r == '`' {
			monoFlag = !monoFlag
			continue
		}

		if monoFlag {
			style = style.Italic(true)
		} else {
			style = style.Italic(false)
		}

		if r == '\n' {
			col = 0
			row++
			continue
		}

		a.Screen.SetContent(col, row, r, nil, style)
		col++
	}

	a.Screen.Show()
}

func (a *App) handleSelection() {
	switch a.State {
	case file_search:
		for idx := range a.QuestionsChosen {
			if err := a.parseQuestions(a.QuestionsFrom[idx].Path); err != nil {
				a.Screen.Fini()
				log.Fatalf("error parsing file: %s", a.QuestionsFrom[idx].Path)
			}
		}
		a.State = selecting
		a.SelectedIndex = 0
	case selecting:
		if a.SelectedIndex == a.Questions[a.CurrentQuestion].Correct {
			a.State = show_correct
			a.postEvent(1 * time.Second)
    } else if _, ok := a.WrongSelected[a.SelectedIndex]; !ok {
			a.State = show_incorrect
			a.postEvent(1 * time.Second)
		}
	}
}

func (a *App) handleKeyDown(event *tcell.EventKey) {
	optionsLength := len(a.QuestionsFrom)

	if a.State != file_search {
		optionsLength = len(a.Questions[a.CurrentQuestion].Options)
	}

	switch event.Key() {
	case tcell.KeyUp:
		a.SelectedIndex = (a.SelectedIndex - 1 + optionsLength) % optionsLength
	case tcell.KeyDown:
		a.SelectedIndex = (a.SelectedIndex + 1) % optionsLength
	case tcell.KeyEnter, tcell.KeyBS:
		switch a.State {
		case file_search:
			if _, ok := a.QuestionsChosen[a.SelectedIndex]; ok {
				delete(a.QuestionsChosen, a.SelectedIndex)
			} else {
				a.QuestionsChosen[a.SelectedIndex] = struct{}{}
			}
		default:
			a.handleSelection()
		}
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			a.SelectedIndex = (a.SelectedIndex + 1) % optionsLength
		case 'k':
			a.SelectedIndex = (a.SelectedIndex - 1 + optionsLength) % optionsLength
		case 'G':
			a.handleSelection()
		}
	}
}

func (a *App) Run() {
	for {
		if a.State != file_search && a.CurrentQuestion > len(a.Questions) {
			break
		}

		if a.State == file_search {
			a.drawConfiguraiton()
		} else {
			a.drawQuestions()
		}

		event := a.Screen.PollEvent()

		switch event := event.(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC || event.Key() == tcell.KeyEscape {
				a.shutdown()
			}

			if a.State == selecting || a.State == file_search {
				a.handleKeyDown(event)
			}
		case *TimeoutEvent:
			if a.State == show_correct {
				a.CurrentQuestion++
				a.WrongSelected = make(map[int]struct{})
			}

			if a.State == show_incorrect {
				a.WrongSelected[a.SelectedIndex] = struct{}{}
			}

			a.State = selecting
			a.SelectedIndex = 0
		}
	}

	a.shutdown()
}

func (a *App) shutdown() {
	a.Screen.Fini()

	color := color.New(color.FgGreen, color.Bold)
	color.Println("You have finished the quiz!")

	os.Exit(0)
}

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
