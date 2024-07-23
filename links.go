package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/kkdai/youtube/v2"
)

func getLinks(client *youtube.Client, source string, nLinks int, skip int) (links []string, err error) {
	isPlaylist := strings.Contains(source, "https://www.youtube.com/playlist")

	if isPlaylist {
		links, err = fetchPlaylistLinks(client, source)
	} else {
		links, err = readLinks(source)
	}

	if err != nil {
		return
	}

	if nLinks != 0 {
		links = links[skip : nLinks+skip]
	}

	return
}

func readLinks(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	links := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 {
			links = append(links, strings.TrimSpace(line))
		}
	}

	return links, nil
}

func fetchPlaylistLinks(client *youtube.Client, link string) ([]string, error) {
	playlist, err := client.GetPlaylist(link)
	if err != nil {
		return nil, err
	}

	links := make([]string, 0, len(playlist.Videos))
	for _, video := range playlist.Videos {
		links = append(links, fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID))
	}

	return links, nil
}
