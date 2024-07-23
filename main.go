package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
)

var client youtube.Client = youtube.Client{}
var nLinks *int
var skip *int
var source string
var output string

var (
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	keyStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("145")).Render
	barTextStyle = lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("0")).Render
	songStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
	removeStyle  = lipgloss.NewStyle()

	failedStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	skippedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("31"))
	downloadedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))

	yesStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render
	maybeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render
	noStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render
)

func main() {

	nLinks = flag.Int("n_links", 0, "Download first given number of youtube links.")
	skip = flag.Int("skip", 0, "Skip first number of youtube links.")
	flag.Parse()

	source = flag.Arg(0)
	output = flag.Arg(1)

	links, err := getLinks(&client, source, *nLinks, *skip)
	if err != nil {
		log.Println(err)
	}

	fetchBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	downloadBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	editBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	m := model{
		fetched:     make([]bool, len(links)),
		downloaded:  make([]bool, len(links)),
		links:       links,
		songs:       make([]Song, 0, len(links)),
		fetchBar:    fetchBar,
		downloadBar: downloadBar,
		editBar:     editBar,
		timer:       timer.NewWithInterval(time.Second*10, time.Second),
		inputs:      make([]textinput.Model, 2),
	}

	for i := range m.inputs {
		input := textinput.New()
		input.CursorStyle = cursorStyle
		input.CharLimit = 64

		switch i {
		case 0:
			input.Placeholder = "Title"
			input.PromptStyle = focusedStyle
			input.TextStyle = focusedStyle
			input.Focus()
		case 1:
			input.Placeholder = "Artist"
		}

		m.inputs[i] = input
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if err := p.Start(); err != nil {
		fmt.Println("Could not start program:", err)
	}
}

type View int

const (
	fetch View = iota
	edit
	finish
)

type fetchMsg *Song
type errorMsg error
type downloadMsg int
type saveMsg int

type model struct {
	failedFetch int
	fetched     []bool

	skipCount  int
	downloaded []bool

	links []string
	songs []Song

	fetchBar    progress.Model
	downloadBar progress.Model
	editBar     progress.Model
	timer       timer.Model

	inputs    []textinput.Model
	focusIndx int

	fetchPercent    float64
	downloadPercent float64
	editPercent     float64

	fetchCount    int
	editIndx      int
	downloadCount int
	failedCount   int

	err      error
	view     int
	quitting bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, fetchCmd(m.links[0]))
}
