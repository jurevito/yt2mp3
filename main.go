package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
	"yt2mp3/title"

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

func main() {

	links, err := readLinks("input.txt")
	if err != nil {
		panic(err)
	}

	client := youtube.Client{Debug: false}

	videos := make([]title.Video, 0, len(links))
	for _, link := range links {
		video, err := client.GetVideo(link)
		if err != nil {
			panic(err)
		}

		parsed := title.Parse(video.Title, video.Author)
		parsed.Format = video.Formats.WithAudioChannels().FindByQuality("tiny")
		videos = append(videos, *parsed)
		// fmt.Printf("%s - %s\n%s\n\n", parsed.Artist, parsed.Title, video.Title)
	}

	sort.Slice(videos, func(i, j int) bool { return videos[i].Reliable < videos[j].Reliable })

	for _, vid := range videos {
		fmt.Printf("%s - %s\n", vid.Artist, vid.Title)
	}

	/*
		// Typically youtube only provides separate streams for video and audio.
		// If you want audio and video combined, take a look a the downloader package.
		format := video.Formats.WithAudioChannels().FindByQuality("tiny")

		reader, _, err := client.GetStream(video, format)
		if err != nil {
			panic(err)
		}

		fmt.Printf("bit rate: %v\n", format.AudioQuality)

		// do something with the reader
		content, err := io.ReadAll(reader)
		if err != nil {
			panic(err)
		}

		reader.Close()

		fileName := fmt.Sprintf("./output/%s.mp3", video.Title)
		err = ioutil.WriteFile(fileName, content, fs.ModePerm)
		if err != nil {
			panic(err)
		}
	*/
}
