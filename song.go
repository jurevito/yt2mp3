package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"

	id3 "github.com/bogem/id3v2"
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

func (s *Song) Save(path string) error {
	reader, _, err := client.GetStream(s.Video, FindFormat(s.Video.Formats))
	if err != nil {
		return fmt.Errorf("Could not get video stream from song \"%s - %s\"", s.Artist, s.Title)
	}

	s.Content, err = io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("Could not read video stream from song \"%s - %s\"", s.Artist, s.Title)
	}

	reader.Close()
	fname := fmt.Sprintf("%s%s - %s", path, s.Artist, s.Title)

	err = ioutil.WriteFile(fmt.Sprintf("%s.mp4", fname), s.Content, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("Could not save mp4 file.")
	}

	mp4 := fmt.Sprintf("%s.mp4", fname)
	mp3 := fmt.Sprintf("%s.mp3", fname)

	cmd := exec.Command("ffmpeg", "-y", "-i", mp4, "-vn", mp3)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%v: %s", err, stderr.String())
	}

	if err = os.Remove(mp4); err != nil {
		return fmt.Errorf("Could not remove mp4 file.")
	}

	tag, err := id3.Open(mp3, id3.Options{Parse: true})
	if err != nil {
		return fmt.Errorf("Could not open mp3 file to edit metadata.")
	}
	defer tag.Close()

	tag.SetArtist(s.Artist)
	tag.SetTitle(s.Title)

	if err = tag.Save(); err != nil {
		return fmt.Errorf("Could not save edited metadata.")
	}

	return nil
}

func FindFormat(formats youtube.FormatList) *youtube.Format {
	for _, format := range formats {
		if format.MimeType == "audio/mp4; codecs=\"mp4a.40.2\"" {
			return &format
		}
	}

	return nil
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

func GetSong(client *youtube.Client, link string) (*Song, error) {
	video, err := client.GetVideo(link)
	if err != nil {
		return nil, err
	}

	song := ParseMetadata(video.Title, video.Author)
	song.Video = video

	return song, nil
}

func ParseMetadata(title string, author string) *Song {

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
