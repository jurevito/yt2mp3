package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
	"time"
	"yt2mp3/yt2mp3"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kkdai/youtube/v2"
	"github.com/muesli/reflow/indent"
)

var client youtube.Client = youtube.Client{Debug: false}
var nLinks *int
var source string
var output string

func main() {

	nLinks = flag.Int("n_links", 0, "Only download first given number of youtube links.")
	flag.Parse()

	source = flag.Arg(0)
	output = flag.Arg(1)

	links, err := getLinks(&client, source, *nLinks)
	if err != nil {
		log.Println(err)
	}

	fetchBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	downloadBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))
	editBar := progress.New(progress.WithScaledGradient("#FF7CCB", "#FDFF8C"))

	m := model{
		links:           links,
		songs:           make([]yt2mp3.Song, 0, len(links)),
		fetchBar:        fetchBar,
		downloadBar:     downloadBar,
		editBar:         editBar,
		fetchPercent:    0,
		downloadPercent: 0,
		editPercent:     0,
		fetchIndx:       0,
		editIndx:        0,
		downloadIndx:    0,
		fetched:         false,
		quitting:        false,
	}
	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		fmt.Println("Could not start program:", err)
	}
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

const (
	padding  = 2
	maxWidth = 80
)

type tickMsg time.Time
type fetchMsg *yt2mp3.Song
type downloadMsg struct{ error }

func (e downloadMsg) Error() string { return e.error.Error() }

type model struct {
	links []string
	songs []yt2mp3.Song

	fetchBar    progress.Model
	downloadBar progress.Model
	editBar     progress.Model

	fetchPercent    float64
	downloadPercent float64
	editPercent     float64

	fetchIndx    int
	editIndx     int
	downloadIndx int

	fetched  bool
	quitting bool
}

func (m model) Init() tea.Cmd {
	return fetchCmd(m.links[m.fetchIndx])
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		k := msg.String()
		if k == "q" || k == "esc" || k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}

	if !m.fetched {
		return updateFetch(msg, m)
	}

	return updateEditor(msg, m)
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
		m.fetchIndx += 1
		m.fetchPercent = float64(m.fetchIndx) / float64(len(m.links))

		if m.fetchIndx < len(m.links) {
			return m, fetchCmd(m.links[m.fetchIndx])
		}

		m.fetched = true
		return m, downloadCmd(&m.songs[0])
	default:
		return m, nil
	}
}

func updateEditor(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.fetchBar.Width = msg.Width - padding*2 - 4
		if m.fetchBar.Width > maxWidth {
			m.fetchBar.Width = maxWidth
		}
		return m, nil
	case downloadMsg:
		m.downloadIndx += 1
		m.downloadPercent = float64(m.downloadIndx) / float64(len(m.songs))

		if m.downloadIndx < m.editIndx {
			return m, downloadCmd(&m.songs[m.downloadIndx])
		}

		m.quitting = (m.downloadIndx == len(m.songs))

		return m, nil
	default:
		return m, nil
	}
}

// The main view, which just calls the appropriate sub-view
func (m model) View() string {
	var s string
	if m.quitting {
		return "\n  See you later!\n\n"
	}

	if !m.fetched {
		s = fetchView(m)
	} else {
		s = editorView(m)
	}
	return indent.String("\n"+s+"\n\n", 2)
}

func fetchView(m model) string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + "Fetching video metadata." + "\n\n" +
		pad + m.fetchBar.ViewAs(m.fetchPercent) + "\n\n" +
		pad + helpStyle("Press any key to quit")
}

func editorView(m model) string {
	pad := strings.Repeat(" ", padding)
	return "\n" +
		pad + "Downloading songs." + "\n\n" +
		pad + m.fetchBar.ViewAs(m.downloadPercent) + "\n\n" +
		pad + "Editing metadata." + "\n\n" +
		pad + m.fetchBar.ViewAs(m.editPercent) + "\n\n"
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func fetchCmd(link string) tea.Cmd {
	return func() tea.Msg {
		song, err := yt2mp3.ParseSong(&client, link)
		if err != nil {
			panic(err)
		}

		return fetchMsg(song)
	}
}

func downloadCmd(song *yt2mp3.Song) tea.Cmd {
	return func() tea.Msg {
		err := SaveSong(song, output)
		if err != nil {
			log.Println(err)
		}
		return downloadMsg{err}
	}
}
