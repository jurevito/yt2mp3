package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"yt2mp3/yt2mp3"

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
		links = links[skip:nLinks]
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

func SaveSong(song *yt2mp3.Song, path string) error {
	reader, _, err := client.GetStream(song.Video, yt2mp3.FindFormat(song.Video.Formats))
	if err != nil {
		return fmt.Errorf("Could not get video stream from song \"%s - %s\"", song.Artist, song.Title)
	}

	song.Content, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Could not read video stream from song \"%s - %s\"", song.Artist, song.Title)
	}

	reader.Close()
	fname := fmt.Sprintf("%s%s - %s", path, song.Artist, song.Title)

	err = ioutil.WriteFile(fmt.Sprintf("%s.mp4", fname), song.Content, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not save mp4 file.")
	}

	mp4 := fmt.Sprintf("%s.mp4", fname)
	mp3 := fmt.Sprintf("%s.mp3", fname)

	title := fmt.Sprintf("title=\"%s\"", song.Title)
	artist := fmt.Sprintf("artist=\"%s\"", song.Artist)
	cmd := exec.Command("ffmpeg", "-y", "-i", mp4, "-vn", "-metadata", title, "-metadata", artist, mp3)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	if err = os.Remove(mp4); err != nil {
		return fmt.Errorf("Could not remove mp4 file.")
	}

	return nil
}
