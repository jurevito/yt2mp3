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

	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) > 0 {
			filtered = append(filtered, strings.TrimSpace(line))
		}
	}

	return filtered, nil
}

func SaveMusic(video *yt2mp3.Video, path string) error {
	fileName := fmt.Sprintf("%s%s - %s.mp3", path, video.Artist, video.Title)

	err := ioutil.WriteFile(fileName, video.Content, fs.ModePerm)
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

	videos, err := yt2mp3.ParseVideos(&client, links)
	if err != nil {
		panic(err)
	}

	for _, video := range videos {
		reader, _, err := client.GetStream(video.Video, yt2mp3.FindFormat(video.Video.Formats))
		if err != nil {
			panic(err)
		}

		video.Content, err = io.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		reader.Close()
		err = SaveMusic(&video, "./output/")
		if err != nil {
			panic(err)
		}
	}
}
