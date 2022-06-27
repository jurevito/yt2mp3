package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"strings"
	"yt2mp3/yt2mp3"

	"github.com/kkdai/youtube/v2"
)

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

func SaveSong(song *yt2mp3.Song, path string) error {
	fileName := fmt.Sprintf("%s%s - %s.mp3", path, song.Artist, song.Title)

	err := ioutil.WriteFile(fileName, song.Content, fs.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func main() {

	links, err := readLinks("input.txt")
	if err != nil {
		panic(err)
	}

	client := youtube.Client{Debug: false}

	songs, err := yt2mp3.ParseSongs(&client, links)
	if err != nil {
		panic(err)
	}

	for _, song := range songs {
		reader, _, err := client.GetStream(song.Video, yt2mp3.FindFormat(song.Video.Formats))
		if err != nil {
			panic(err)
		}

		song.Content, err = io.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		reader.Close()
		err = SaveSong(&song, "./output/")
		if err != nil {
			panic(err)
		}
	}
}
