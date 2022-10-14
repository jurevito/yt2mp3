package main

import (
	"fmt"
	"strings"

	"github.com/muesli/reflow/indent"
)

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
