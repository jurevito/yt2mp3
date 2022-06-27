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

type Video struct {
	Title    string
	Artist   string
	Video    *youtube.Video
	Content  []byte
	Reliable Reliable
}

func ParseTitle(title string, author string) *Video {

	video := Video{
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
		video.Artist = strings.TrimSpace(titleParts[0])
		video.Title = strings.TrimSpace(titleParts[1])
		video.Reliable = yes

		// Keep only first listed artist.
		regex = regexp.MustCompile(`^[^(&|,)]*[&|,]`)
		if regex.MatchString(video.Artist) {
			video.Artist = regex.FindString(video.Artist)
			video.Artist = strings.TrimSpace(video.Artist[:len(video.Artist)-1])
			video.Reliable = maybe
		}

		if strings.Contains(video.Title, "|") {
			video.Reliable = maybe
		}

	} else {
		video.Title = strings.TrimSpace(titleParts[0])
		video.Reliable = no
	}

	return &video
}

func FindFormat(formats youtube.FormatList) *youtube.Format {
	for _, format := range formats {
		if format.MimeType == "audio/mp4; codecs=\"mp4a.40.2\"" {
			return &format
		}
	}

	return nil
}

func ParseVideos(client *youtube.Client, links []string) ([]Video, error) {

	videos := make([]Video, 0, len(links))
	for _, link := range links {
		video, err := client.GetVideo(link)
		if err != nil {
			return nil, err
		}

		parsed := ParseTitle(video.Title, video.Author)
		parsed.Video = video
		videos = append(videos, *parsed)
	}

	sort.Slice(videos, func(i, j int) bool { return videos[i].Reliable < videos[j].Reliable })
	return videos, nil
}
