package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"sort"
	"strings"
	"time"

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

type fetchMsg *Song
type errorMsg error
type downloadMsg int
type saveMsg int

type model struct {
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
		m.songs = append(m.songs, *(*Song)(msg))
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
				m.skipCount += 1
				m.downloadPercent = float64(m.downloadCount+m.failedCount+m.skipCount) / float64(len(m.songs))

				m.editIndx += 1
				m.editPercent = float64(m.editIndx) / float64(len(m.songs))
			}

			if m.downloadCount+m.failedCount+m.skipCount == len(m.songs) {
				skipped := make([]string, 0, len(m.links))
				for i := range m.links {
					if !m.downloaded[i] {
						skipped = append(skipped, m.links[i])
					}
				}
				return m, saveCmd(skipped)
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
				cmd = downloadCmd(&m.songs[m.editIndx], m.editIndx)

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
		m.downloadPercent = float64(m.downloadCount+m.failedCount+m.skipCount) / float64(len(m.songs))
		m.downloaded[int(msg)] = true

		if m.downloadCount+m.failedCount+m.skipCount == len(m.songs) {
			skipped := make([]string, 0, len(m.links))
			for i := range m.links {
				if !m.downloaded[i] {
					skipped = append(skipped, m.links[i])
				}
			}
			return m, saveCmd(skipped)
		}

		return m, nil
	case errorMsg:
		m.failedCount += 1
		m.downloadPercent = float64(m.downloadCount+m.failedCount+m.skipCount) / float64(len(m.songs))

		if m.downloadCount+m.failedCount+m.skipCount == len(m.songs) {
			skipped := make([]string, 0, len(m.links))
			for i := range m.links {
				if !m.downloaded[i] {
					skipped = append(skipped, m.links[i])
				}
			}
			return m, saveCmd(skipped)
		}

		return m, nil
	case saveMsg:
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
	var b strings.Builder
	b.WriteString("\n")

	if m.quitting {
		b.WriteString("  See you later!")
		return indent.String("\n"+b.String()+"\n\n", 2)
	}

	switch m.view {
	case int(fetch):
		b.WriteString(fetchView(m))
	case int(edit):
		b.WriteString(editorView(m))
	case int(finish):
		b.WriteString(finishView(m))
	}

	return indent.String("\n"+b.String()+"\n\n", 2)
}

func fetchView(m model) string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(barTextStyle("Fetching video metadata.") + "\n")
	b.WriteString(m.fetchBar.ViewAs(m.fetchPercent) + "\n\n")
	b.WriteString(helpStyle("Press ") + keyStyle("q") + helpStyle(" button to quit."))

	return b.String()
}

func editorView(m model) string {
	var b strings.Builder

	if m.editIndx < len(m.songs) {
		// Original title and author of a video.
		b.WriteString(helpStyle("Title : "))
		b.WriteString(songStyle.Render((m.songs[m.editIndx].Video.Title)) + "\n")
		b.WriteString(helpStyle("Author: "))
		b.WriteString(songStyle.Render(m.songs[m.editIndx].Video.Author) + "\n\n")

		for i := range m.inputs {
			b.WriteString(m.inputs[i].View() + "\n")
		}

		b.WriteString("\n")
	}

	// Statistics tab with current progress.
	failed := failedStyle.Render(fmt.Sprintf("%2d", m.failedCount))
	skipped := skippedStyle.Render(fmt.Sprintf("%2d", m.skipCount))
	downloaded := downloadedStyle.Render(fmt.Sprintf("%2d", m.downloadCount))
	b.WriteString(fmt.Sprintf("%s failed • %s skipped • %s downloaded\n", failed, skipped, downloaded))

	// Render progress bars.
	b.WriteString("\n")
	b.WriteString(barTextStyle("Downloading songs.") + "\n")
	b.WriteString(m.fetchBar.ViewAs(m.downloadPercent) + "\n\n")
	b.WriteString(barTextStyle("Editing metadata.") + "\n")
	b.WriteString(m.fetchBar.ViewAs(m.editPercent) + "\n\n")

	// Render help.
	b.WriteString(helpStyle("ctrl+r reset • ctrl+s skip • enter confirm • ↑/↓ move • q quit"))
	b.WriteString("\n")

	return b.String()
}

func finishView(m model) string {
	var b strings.Builder

	// Statistics tab with current progress.
	failed := failedStyle.Render(fmt.Sprintf("%2d", m.failedCount))
	skipped := skippedStyle.Render(fmt.Sprintf("%2d", m.skipCount))
	downloaded := downloadedStyle.Render(fmt.Sprintf("%2d", m.downloadCount))
	b.WriteString(fmt.Sprintf("%s failed • %s skipped • %s downloaded\n", failed, skipped, downloaded))

	b.WriteString(helpStyle(fmt.Sprintf("Exiting in %s\n", m.timer.View())))
	b.WriteString(fmt.Sprint(m.err))

	return b.String()
}

func fetchCmd(link string) tea.Cmd {
	return func() tea.Msg {
		song, err := GetSong(&client, link)
		if err != nil {
			return errorMsg(err)
		}

		return fetchMsg(song)
	}
}

func downloadCmd(song *Song, index int) tea.Cmd {
	return func() tea.Msg {

		max, min := 5, 1
		delay := rand.Intn(max-min) + min
		err := retry(5, time.Duration(delay)*time.Second, func() error {
			return song.Save(output)
		})

		if err != nil {
			return errorMsg(err)
		}
		return downloadMsg(index)
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

func saveCmd(skipped []string) tea.Cmd {
	return func() tea.Msg {
		path := output + "failed.txt"
		b := []byte(strings.Join(skipped, "\n"))
		err := ioutil.WriteFile(path, b, 0644)
		if err != nil {
			return errorMsg(err)
		}

		return saveMsg(len(skipped))
	}
}

func reliabilityColor(r Reliable) lipgloss.Color {
	switch r {
	case Maybe:
		return lipgloss.Color("3")
	case No:
		return lipgloss.Color("9")
	default:
		return lipgloss.Color("2")
	}
}
