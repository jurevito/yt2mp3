package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"
	"yt2mp3/yt2mp3"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
	"github.com/muesli/reflow/indent"
)

var client youtube.Client = youtube.Client{Debug: false}
var nLinks *int
var skip *int
var source string
var output string

var (
	focusedStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	cursorStyle         = focusedStyle.Copy()
	removeStyle         = lipgloss.NewStyle()
	helpStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	yesStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render
	maybeStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render
	noStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
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
		skip:        make([]bool, len(links)),
		links:       links,
		songs:       make([]yt2mp3.Song, 0, len(links)),
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

const (
	padding  = 2
	maxWidth = 80
)

type View int

const (
	fetch View = iota
	edit
	finish
)

type fetchMsg *yt2mp3.Song
type errorMsg error
type downloadMsg struct{}

type model struct {
	skipCount int
	skip      []bool

	links []string
	songs []yt2mp3.Song

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

	err      error
	view     int
	quitting bool
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, fetchCmd(m.links[0]))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	switch m.view {
	case int(fetch):
		return updateFetch(msg, m)
	case int(edit):
		return updateEditor(msg, m)
	case int(finish):
		return updateFinish(msg, m)
	default:
		return updateEditor(msg, m) // TODO
	}
}

func updateFetch(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.fetchBar.Width = msg.Width - padding*2 - 4
		if m.fetchBar.Width > maxWidth {
			m.fetchBar.Width = maxWidth
		}
		return m, nil
	case fetchMsg:
		m.songs = append(m.songs, *(*yt2mp3.Song)(msg))
		m.fetchCount += 1
		m.fetchPercent = float64(m.fetchCount) / float64(len(m.links))

		if m.fetchCount < len(m.links) {
			return m, fetchCmd(m.links[m.fetchCount])
		}

		sort.Slice(m.songs, func(i, j int) bool {
			return m.songs[i].Reliable < m.songs[j].Reliable
		})

		m.view = int(edit)
		m.inputs[0].SetValue(m.songs[0].Title)
		m.inputs[1].SetValue(m.songs[0].Artist)

		return m, nil
	case errorMsg:
		m.view = int(finish)
		m.err = error(msg)
		return m, m.timer.Init()
	default:
		return m, nil
	}
}

func updateEditor(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab", "up", "down":
			s := msg.String()

			// Cycle indexes
			if s == "up" || s == "shift+tab" {
				m.focusIndx--
			} else {
				m.focusIndx++
			}

			if m.focusIndx >= len(m.inputs) {
				m.focusIndx = 0
			} else if m.focusIndx < 0 {
				m.focusIndx = len(m.inputs) - 1
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndx {
					// Set focused state
					cmds[i] = m.inputs[i].Focus()
					m.inputs[i].PromptStyle = focusedStyle
					m.inputs[i].TextStyle = focusedStyle
					continue
				}
				// Remove focused state
				m.inputs[i].Blur()
				m.inputs[i].PromptStyle = removeStyle
				m.inputs[i].TextStyle = removeStyle
			}

			return m, tea.Batch(cmds...)
		case "ctrl+r":
			m.inputs[0].SetValue(m.songs[m.editIndx].Title)
			m.inputs[1].SetValue(m.songs[m.editIndx].Artist)
			m.inputs[0].CursorEnd()
			m.inputs[1].CursorEnd()
			return m, nil
		case "ctrl+s":
			if m.editIndx < len(m.songs) {
				m.skip[m.editIndx] = true
				m.skipCount += 1

				m.downloadCount += 1
				m.downloadPercent = float64(m.downloadCount) / float64(len(m.songs))

				m.editIndx += 1
				m.editPercent = float64(m.editIndx) / float64(len(m.songs))
			}

			if m.downloadCount == len(m.songs) {
				m.view = int(finish)
				return m, m.timer.Init()
			}

			if m.editIndx < len(m.songs) {
				m.inputs[0].SetValue(m.songs[m.editIndx].Title)
				m.inputs[1].SetValue(m.songs[m.editIndx].Artist)

				m.inputs[0].CursorEnd()
				m.inputs[1].CursorEnd()
			}

			return m, nil
		case "enter":
			var cmd tea.Cmd
			if m.editIndx < len(m.songs) {
				m.songs[m.editIndx].Title = m.inputs[0].Value()
				m.songs[m.editIndx].Artist = m.inputs[1].Value()
				cmd = downloadCmd(&m.songs[m.editIndx])

				m.editIndx += 1
				m.editPercent = float64(m.editIndx) / float64(len(m.songs))
			}

			if m.editIndx < len(m.songs) {
				m.inputs[0].SetValue(m.songs[m.editIndx].Title)
				m.inputs[1].SetValue(m.songs[m.editIndx].Artist)

				m.inputs[0].CursorEnd()
				m.inputs[1].CursorEnd()
			}

			return m, cmd
		}
	case tea.WindowSizeMsg:
		m.fetchBar.Width = msg.Width - padding*2 - 4
		if m.fetchBar.Width > maxWidth {
			m.fetchBar.Width = maxWidth
		}
		return m, nil
	case downloadMsg:

		m.downloadCount += 1
		m.downloadPercent = float64(m.downloadCount) / float64(len(m.songs))

		if m.downloadCount == len(m.songs) {
			m.view = int(finish)
			return m, m.timer.Init()
		}

		return m, nil
	case errorMsg:
		m.view = int(finish)
		return m, m.timer.Init()
	}

	// Handle character input and blinking
	cmd := m.updateInputs(msg)
	return m, cmd
}

func (m *model) updateInputs(msg tea.Msg) tea.Cmd {
	var cmds = make([]tea.Cmd, len(m.inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func updateFinish(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timer.TickMsg:
		var cmd tea.Cmd
		m.timer, cmd = m.timer.Update(msg)
		return m, cmd
	case timer.TimeoutMsg:
		m.quitting = true
		return m, tea.Quit
	case errorMsg:
		m.err = error(msg)
		return m, nil
	}

	return m, nil
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.quitting {
		return "\n  See you later!\n\n"
	}

	switch m.view {
	case int(fetch):
		s = fetchView(m)
	case int(edit):
		s = editorView(m)
	case int(finish):
		s = finishView(m)
	}

	return indent.String("\n"+s+"\n\n", 2)
}

func fetchView(m model) string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + "Fetching video metadata." + "\n" +
		pad + m.fetchBar.ViewAs(m.fetchPercent) + "\n\n" +
		pad + helpStyle("Press q button to quit.")
}

func editorView(m model) string {
	var b strings.Builder

	if m.editIndx < len(m.songs) {
		// Render title and author of a video.
		b.WriteString(m.songs[m.editIndx].Video.Title + "\n")
		b.WriteString(m.songs[m.editIndx].Video.Author + "\n\n")

		// Render reliability of title parsing.
		switch m.songs[m.editIndx].Reliable {
		case yt2mp3.Yes:
			b.WriteString(yesStyle("[ ✓ ]"))
		case yt2mp3.Maybe:
			b.WriteString(maybeStyle("[ ? ]"))
		case yt2mp3.No:
			b.WriteString(noStyle("[ ! ]"))
		}
		b.WriteString("\n")

		// Render text inputs.
		for i := range m.inputs {
			b.WriteString(m.inputs[i].View() + "\n")
		}
	}

	// Render progress bars.
	pad := strings.Repeat(" ", padding)
	b.WriteString("\n")
	b.WriteString(pad + "Downloading songs." + "\n")
	b.WriteString(pad + m.fetchBar.ViewAs(m.downloadPercent) + "\n\n")
	b.WriteString(pad + "Editing metadata." + "\n")
	b.WriteString(pad + m.fetchBar.ViewAs(m.editPercent) + "\n\n")

	// Render help.
	b.WriteString(helpStyle("ctrl+r reset • ctrl+s skip • enter confirm • ↑/↓ move • q quit"))
	b.WriteString("\n")

	return b.String()
}

func finishView(m model) string {
	pad := strings.Repeat(" ", padding)
	var b strings.Builder
	b.WriteString(pad + focusedStyle.Render("FINISH VIEW!\n"))
	b.WriteString(pad + m.timer.View() + "\n")
	b.WriteString(fmt.Sprint(m.err))

	return b.String()
}

func fetchCmd(link string) tea.Cmd {
	return func() tea.Msg {
		song, err := yt2mp3.ParseSong(&client, link)
		if err != nil {
			return errorMsg(err)
		}

		return fetchMsg(song)
	}
}

func downloadCmd(song *yt2mp3.Song) tea.Cmd {
	return func() tea.Msg {

		min := 1
		max := 5
		delay := rand.Intn(max-min) + min
		err := retry(5, time.Duration(delay)*time.Second, func() error {
			return SaveSong(song, output)
		})

		if err != nil {
			return errorMsg(err)
		}
		return downloadMsg{}
	}
}

func retry(attempts int, sleep time.Duration, f func() error) error {
	if err := f(); err != nil {
		if attempts--; attempts > 0 {
			time.Sleep(sleep)
			return retry(attempts, 4*sleep, f)
		}
		return err
	}
	return nil
}
