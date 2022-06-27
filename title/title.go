package title

import (
	"fmt"
	"regexp"
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
	Format   youtube.Format
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

	fmt.Printf("%s\n", title)

	// Split artist and title by '-'.
	regex = regexp.MustCompile(`[-|–]+`)
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

	} else {
		video.Title = strings.TrimSpace(titleParts[0])
		video.Reliable = no
	}

	return &video
}
