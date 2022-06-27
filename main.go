package main

import (
	"fmt"
	"yt2mp3/title"
)

func main() {

	video := title.ParseTitle("Sting - What Could Have Been | Arcane League of Legends | Riot Games Music", "")
	fmt.Printf("%s\n", video.Title)

	/*
		link := "https://www.youtube.com/watch?v=jH1RNk8954Q&list=PL6YgdMS9Bn4GPY9Sm7BJtOwiVYzN-UINM&index=21"

		client := youtube.Client{Debug: true}

		video, err := client.GetVideo(link)
		if err != nil {
			panic(err)
		}

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
