package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"yt2mp3/yt2mp3"

	id3 "github.com/bogem/id3v2/v2"
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
	fname := fmt.Sprintf("%s%s - %s", path, song.Artist, song.Title)

	err := ioutil.WriteFile(fmt.Sprintf("%s.mp4", fname), song.Content, fs.ModePerm)
	if err != nil {
		log.Printf("fck1: %v\n", err)
		return err
	}

	mp4 := fmt.Sprintf("%s.mp4", fname)
	mp3 := fmt.Sprintf("%s.mp3", fname)

	cmd := exec.Command("ffmpeg", "-y", "-i", mp4, "-vn", mp3)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	if err = os.Remove(mp4); err != nil {
		return err
	}

	tag, err := id3.Open(mp3, id3.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	tag.SetArtist(song.Artist)
	tag.SetTitle(song.Title)

	if err = tag.Save(); err != nil {
		return err
	}

	return nil
}

func main() {

	client := youtube.Client{Debug: false}

	links, err := fetchPlaylistLinks(&client, "https://www.youtube.com/playlist?list=PL6YgdMS9Bn4FLSnpv368M3s3_cysoeBkT")
	if err != nil {
		log.Println(err)
	}

	songs, err := yt2mp3.ParseSongs(&client, links)
	if err != nil {
		log.Println(err)
	}

	fmt.Printf("Parsed songs and fetched streaming links.\n")

	for _, song := range songs {
		reader, _, err := client.GetStream(song.Video, yt2mp3.FindFormat(song.Video.Formats))
		if err != nil {
			log.Println(err)
		}

		song.Content, err = io.ReadAll(reader)
		if err != nil {
			log.Println(err)
		}

		reader.Close()
		err = SaveSong(&song, "D:/Dokumenti/Documents/Projekti/yt2mp3/output/")
		if err != nil {
			log.Println(err)
		}
	}

	fmt.Printf("Downloaded and saved all %d songs.\n", len(songs))
}
