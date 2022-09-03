package yt2mp3

import (
	"regexp"
	"strings"

	"github.com/kkdai/youtube/v2"
	"golang.org/x/exp/slices"
)

type Reliable int

const (
	Yes Reliable = iota
	Maybe
	No
)

type Song struct {
	Title    string
	Artist   string
	Video    *youtube.Video
	Content  []byte
	Reliable Reliable
}

func RemoveSpecialChars(s string) string {

	sb := []byte(s)
	special := []byte("<>,:\"/\\|?*")

	n := 0
	for i := 0; i < len(sb); i++ {
		if !slices.Contains(special, s[i]) {
			sb[n] = sb[i]
			n++
		}
	}

	return string(sb[:n])
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
	regex = regexp.MustCompile(`( [-]+ | [–]+ )`)
	titleParts := regex.Split(title, -1)
	artistParts := regex.Split(author, -1)

	if len(titleParts) > 1 {
		song.Artist = strings.TrimSpace(titleParts[0])
		song.Title = strings.TrimSpace(titleParts[1])
		song.Reliable = Yes

		// Keep only first listed artist.
		regex = regexp.MustCompile(`^[^(&|,)]*[&|,]`)
		if regex.MatchString(song.Artist) {
			song.Artist = strings.TrimSpace(regex.FindString(song.Artist))
			song.Artist = strings.TrimSpace(song.Artist[:len(song.Artist)-1])
			song.Reliable = Maybe
		}

		if strings.Contains(song.Title, "|") {
			song.Reliable = Maybe
		}

	} else {
		song.Artist = strings.TrimSpace(artistParts[0])
		song.Title = strings.TrimSpace(titleParts[0])
		song.Reliable = No
	}

	// Remove special characters for file saving.
	song.Artist = RemoveSpecialChars(song.Artist)
	song.Title = RemoveSpecialChars(song.Title)

	// Remove redundant spaces.
	regex = regexp.MustCompile(` {2,}`)
	song.Artist = regex.ReplaceAllLiteralString(song.Artist, " ")
	song.Title = regex.ReplaceAllLiteralString(song.Title, " ")

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

func ParseSong(client *youtube.Client, link string) (*Song, error) {
	video, err := client.GetVideo(link)
	if err != nil {
		panic(err)
	}

	song := ParseTitle(video.Title, video.Author)
	song.Video = video

	return song, nil
}
