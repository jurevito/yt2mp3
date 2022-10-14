package main

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

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
