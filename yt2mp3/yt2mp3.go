package yt2mp3

import (
	"regexp"
	"sort"
	"strings"

	"github.com/kkdai/youtube/v2"
)

type Reliable int

const (
	yes Reliable = iota
	maybe
	no
)

type Song struct {
	Title    string
	Artist   string
	Video    *youtube.Video
	Content  []byte
	Reliable Reliable
}

func ParseTitle(title string, author string) *Song {

	song := Song{
		Title:  title,
		Artist: author,
	}

	// Remove [] and () closed brackets.
	regex := regexp.MustCompile(`[\(\[][^(\)\])]*[\)\]]`)
	for regex.MatchString(title) {
		title = regex.ReplaceAllLiteralString(title, "")
	}

	// Split artist and title by '-'.
	regex = regexp.MustCompile(`[-–]+`)
	titleParts := regex.Split(title, -1)

	if len(titleParts) > 1 {
		song.Artist = strings.TrimSpace(titleParts[0])
		song.Title = strings.TrimSpace(titleParts[1])
		song.Reliable = yes

		// Keep only first listed artist.
		regex = regexp.MustCompile(`^[^(&|,)]*[&|,]`)
		if regex.MatchString(song.Artist) {
			song.Artist = regex.FindString(song.Artist)
			song.Artist = strings.TrimSpace(song.Artist[:len(song.Artist)-1])
			song.Reliable = maybe
		}

		if strings.Contains(song.Title, "|") {
			song.Reliable = maybe
		}

	} else {
		song.Title = strings.TrimSpace(titleParts[0])
		song.Reliable = no
	}

	return &song
}

func FindFormat(formats youtube.FormatList) *youtube.Format {
	for _, format := range formats {
		if format.MimeType == "audio/mp4; codecs=\"mp4a.40.2\"" {
			return &format
		}
	}

	return nil
}

func ParseSongs(client *youtube.Client, links []string) ([]Song, error) {

	songs := make([]Song, 0, len(links))
	for _, link := range links {
		video, err := client.GetVideo(link)
		if err != nil {
			return nil, err
		}

		parsed := ParseTitle(video.Title, video.Author)
		parsed.Video = video
		songs = append(songs, *parsed)
	}

	sort.Slice(songs, func(i, j int) bool { return songs[i].Reliable < songs[j].Reliable })
	return songs, nil
}
