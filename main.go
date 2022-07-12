package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"yt2mp3/yt2mp3"

	"github.com/kkdai/youtube/v2"
)

func main() {

	nLinks := flag.Int("n_links", 0, "Only download first given number of youtube links.")
	flag.Parse()

	source := flag.Arg(0)
	output := flag.Arg(1)

	client := youtube.Client{Debug: false}

	links, err := getLinks(&client, source, *nLinks)
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
		err = SaveSong(&song, output)
		if err != nil {
			log.Println(err)
		}
	}

	fmt.Printf("Downloaded and saved all %d songs.\n", len(songs))
}
