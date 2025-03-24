package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
)

const (
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
	Screen          tcell.Screen
	SelectedIndex   int
	State           string
	WrongSelected   map[int]struct{}
}

func NewApp(path string) (*App, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("opening questions file: %w", err)
	}

	var questions []Question

	err = json.Unmarshal(file, &questions)
	if err != nil {
		return nil, fmt.Errorf("unable to decode file: %w", err)
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
		Questions:       questions,
		Screen:          screen,
		SelectedIndex:   0,
		State:           selecting,
		WrongSelected:   make(map[int]struct{}),
	}, nil
}

func (a *App) postEvent(duration time.Duration) {
	time.AfterFunc(duration, func() {
		event := &TimeoutEvent{}
		event.SetEventNow()
		_ = a.Screen.PostEvent(event)
	})
}

func (a *App) renderScreen() {
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

	if a.State == show_incorrect {
		col = 0
		row++
		textStyle := tcell.StyleDefault.Foreground(tcell.ColorRed).Bold(true)
		message := "Your choice is incorrect"

		for _, r := range message {
			a.Screen.SetContent(col, row, r, nil, textStyle)
			col++
		}
	}

	a.Screen.Show()
}

func (a *App) handleSelection() {
	if a.SelectedIndex == a.Questions[a.CurrentQuestion].Correct {
		a.State = show_correct
		a.postEvent(2 * time.Second)
	} else {
		a.State = show_incorrect
		a.postEvent(2 * time.Second)
	}
}

func (a *App) handleKeyDown(event *tcell.EventKey) {
	optionsLength := len(a.Questions[a.CurrentQuestion].Options)

	switch event.Key() {
	case tcell.KeyUp:
		a.SelectedIndex = (a.SelectedIndex - 1 + optionsLength) % optionsLength
	case tcell.KeyDown:
		a.SelectedIndex = (a.SelectedIndex + 1) % optionsLength
	case tcell.KeyEnter, tcell.KeyBS:
		a.handleSelection()
	case tcell.KeyRune:
		switch event.Rune() {
		case 'j':
			a.SelectedIndex = (a.SelectedIndex + 1) % optionsLength
		case 'k':
			a.SelectedIndex = (a.SelectedIndex - 1 + optionsLength) % optionsLength
		}
	}
}

func (a *App) Run() {
	quit := func() {
		a.Screen.Fini()
		os.Exit(0)
	}

	for {
		if a.CurrentQuestion == len(a.Questions) {
			break
		}

		a.renderScreen()

		event := a.Screen.PollEvent()

		switch event := event.(type) {
		case *tcell.EventKey:
			if event.Key() == tcell.KeyCtrlC || event.Key() == tcell.KeyEscape {
				quit()
			}

			if a.State == selecting {
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
}

func main() {
	path := flag.String("path", "cards.json", "path to the file with questions")

	app, err := NewApp(*path)
	if err != nil {
		log.Fatal(err)
	}

	app.Run()
}
