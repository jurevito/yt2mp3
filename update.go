package main

import (
	"sort"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

func filter[A any](v []A, m []bool) []A {
	r := make([]A, 0, len(v))
	for i := range v {
		if m[i] {
			r = append(r, v[i])
		}
	}

	return r
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
		m.fetched[m.fetchCount] = true
		m.songs = append(m.songs, *(*Song)(msg))
		m.fetchCount += 1
		m.fetchPercent = float64(m.fetchCount) / float64(len(m.links))

		if m.fetchCount < len(m.links) {
			return m, fetchCmd(m.links[m.fetchCount])
		}

		m.links = filter(m.links, m.fetched)
		m.songs = filter(m.songs, m.fetched)

		sort.Slice(m.songs, func(i, j int) bool {
			return m.songs[i].Reliable < m.songs[j].Reliable
		})

		m.view = int(edit)
		m.inputs[0].SetValue(m.songs[0].Title)
		m.inputs[1].SetValue(m.songs[0].Artist)

		return m, nil
	case errorMsg:
		m.failedFetch += 1
		m.songs = append(m.songs, Song{})
		m.fetchCount += 1
		m.fetchPercent = float64(m.fetchCount) / float64(len(m.links))

		if m.fetchCount < len(m.links) {
			return m, fetchCmd(m.links[m.fetchCount])
		}

		m.links = filter(m.links, m.fetched)
		m.songs = filter(m.songs, m.fetched)

		sort.Slice(m.songs, func(i, j int) bool {
			return m.songs[i].Reliable < m.songs[j].Reliable
		})

		m.view = int(edit)
		m.inputs[0].SetValue(m.songs[0].Title)
		m.inputs[1].SetValue(m.songs[0].Artist)

		return m, nil
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
