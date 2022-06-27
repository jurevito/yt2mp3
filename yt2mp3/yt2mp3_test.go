package yt2mp3

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testVideo struct {
	videoTitle  string
	videoAuthor string
	title       string
	artist      string
	reliable    Reliable
}

func TestParseTitle(t *testing.T) {
	testVideos := []testVideo{
		{
			"Ava Max - Kings & Queens [Official Music Video]",
			"Ava Max",
			"Kings & Queens",
			"Ava Max",
			yes,
		},
		{
			"Jim Croce - Dont Mess Around With Jim (Remaster Best Quality)",
			"remastermusic",
			"Dont Mess Around With Jim",
			"Jim Croce",
			yes,
		},
		{
			"Tomislav Ivčić - Večeras je naša fešta",
			"Torcida1950Solin",
			"Večeras je naša fešta",
			"Tomislav Ivčić",
			yes,
		},
		{
			"Našoj Ljubavi Je Kraj",
			"Oliver Dragojević Crorec Official",
			"Našoj Ljubavi Je Kraj",
			"Oliver Dragojević Crorec Official",
			no,
		},
		{
			"Mausberg (Feat. DJ Quik) - Get Nekkid - HQ",
			"Camiousse",
			"Get Nekkid",
			"Mausberg",
			yes,
		},
		{
			"Siddharta - Ledena (official video) - Album Infra",
			"Nika Records",
			"Ledena",
			"Siddharta",
			yes,
		},
		{
			"Dino Merlin - Kad si rekla da me voliš (Official Audio) [2000]",
			"Dino Merlin",
			"Kad si rekla da me voliš",
			"Dino Merlin",
			yes,
		},
		{
			"Steve Angello & Laidback Luke Feat. Robin S – Show Me Love (Official HD Video) [2009]",
			"Ministry of Sound",
			"Show Me Love",
			"Steve Angello",
			maybe,
		},
		{
			"Logic - Wu Tang Forever ft. Wu Tang Clan (Official Audio)",
			"Visionary Music Group",
			"Wu Tang Forever ft. Wu Tang Clan",
			"Logic",
			yes,
		},
		{
			"Sting - What Could Have Been | Arcane League of Legends | Riot Games Music",
			"Riot Games Music",
			"What Could Have Been | Arcane League of Legends | Riot Games Music",
			"Sting",
			maybe,
		},
		{
			"TECHNO MIX 2021 | DJD3",
			"DJD3",
			"TECHNO MIX 2021 | DJD3",
			"DJD3",
			no,
		},
	}
	//

	for _, song := range testVideos {
		video := ParseTitle(song.videoTitle, song.videoAuthor)
		require.Equal(t, song.title, video.Title)
		require.Equal(t, song.artist, video.Artist)
		require.Equal(t, song.reliable, video.Reliable)
	}
}
