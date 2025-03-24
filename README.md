# Go-Cards

Go-Cards is a terminal-based quiz application that allows users to answer multiple-choice questions interactively. It supports customizable JSON question files, provides immediate feedback with color-coded responses, and features a keyboard-driven TUI interface.

## Prerequisites

- Go: Install Go (version 1.16 or higher).

## Installation

### `go install`:

```bash
go install github.com/mashfeii/go-cards@main
```

### Manual Build:

- Clone/download the main.go file.
- Build and run:

```bash
go build -o go-cards main.go
./go-cards -path=cards.json
```

## Usage

### Question File Structure

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

- **Navigation**:
  `Up`/`Down` arrows or `j`/`k` to move between options.
- **Submit Answer**:
  `Enter` or `Space`.
- **Exit**:
  `Ctrl+C` or `Esc`.
- **Color Indicators**:
  - 🟢 Green: Correct answer (shown after submission).
  - 🔴 Red: Incorrectly selected option (marked with a red icon).
  - ⚪ Gray: Previously wrong selections.

## Troubleshooting

- **Installation issues**:
  - Ensure `GOPATH/bin` is in your `PATH` to run `go-cards` globally.
  - Run `go mod tidy` to install dependencies (`tcell`, `color`).
- **JSON errors**:
  - Validate your JSON file structure using tools like `JSONLint`.
- **Key bindings not working**:
  - Use `Enter` instead of `Space`.
  - Ensure your terminal emulator supports keyboard inputs (e.g., not some web-based terminals).
- **No colors displayed**:
  - Enable ANSI color support in your terminal settings.
