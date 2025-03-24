# Go-Cards

Go-Cards is a terminal-based quiz application that allows users to answer multiple-choice questions interactively. It supports customizable JSON question files, provides immediate feedback with color-coded responses, and features a keyboard-driven TUI interface.

## Prerequisites

- Go: Install Go (version 1.16 or higher).

## Installation

### Manual Build:

- Clone/download the main.go file.
- Build and run:

```bash
go build -o go-cards main.go
./go-cards -path=cards.json
```

## Usage

Question File Structure
Create a JSON file (e.g., `cards.json`) with this format:

```json
[
  {
    "question": "What is Go?",
    "options": ["Programming Language", "Animal", "Game"],
    "correct": 0 // Index of the correct option (starts at 0)
  }
]
```

## Running the App

Use the default `cards.json`:

```bash
go-cards
```

Specify a custom file:

```bash
go-cards -path=my_questions.json
```

### Controls

- __Navigation__:
  `Up`/`Down` arrows or `j`/`k` to move between options.
- __Submit Answer__:
  `Enter` or `Space`.
- __Exit__:
  `Ctrl+C` or `Esc`.
- __Color Indicators__:
  - 🟢 Green: Correct answer (shown after submission).
  - 🔴 Red: Incorrectly selected option (marked with a red icon).
  - ⚪ Gray: Previously wrong selections.

## Troubleshooting

- __Installation issues__:
  - Ensure `GOPATH/bin` is in your `PATH` to run `go-cards` globally.
  - Run `go mod tidy` to install dependencies (`tcell`, `color`).
- __JSON errors__:
  - Validate your JSON file structure using tools like `JSONLint`.
- __Key bindings not working__:
  - Use `Enter` instead of `Space`.
  - Ensure your terminal emulator supports keyboard inputs (e.g., not some web-based terminals).
- __No colors displayed__:
  - Enable ANSI color support in your terminal settings.
