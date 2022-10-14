package main

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

func TestRemoveSpacialChars(t *testing.T) {

	output := []string{"What is that", ""}
	input := []string{"What \"is\" that?", ":,?*\\/<>"}

	for i := 0; i < len(output); i++ {
		require.Equal(t, output[i], RemoveSpecialChars(input[i]))
	}
}

func TestParseTitle(t *testing.T) {
	testVideos := []testVideo{
		{
			"Ava Max - Kings & Queens [Official Music Video]",
			"Ava Max",
			"Kings & Queens",
			"Ava Max",
			Yes,
		},
		{
			"Jim Croce - Dont Mess Around With Jim (Remaster Best Quality)",
			"remastermusic",
			"Dont Mess Around With Jim",
			"Jim Croce",
			Yes,
		},
		{
			"Tomislav Ivčić - Večeras je naša fešta",
			"Torcida1950Solin",
			"Večeras je naša fešta",
			"Tomislav Ivčić",
			Yes,
		},
		{
			"Našoj Ljubavi Je Kraj",
			"Oliver Dragojević Crorec Official",
			"Našoj Ljubavi Je Kraj",
			"Oliver Dragojević Crorec Official",
			No,
		},
		{
			"Mausberg (Feat. DJ Quik) - Get Nekkid - HQ",
			"Camiousse",
			"Get Nekkid",
			"Mausberg",
			Yes,
		},
		{
			"Siddharta - Ledena (official video) - Album Infra",
			"Nika Records",
			"Ledena",
			"Siddharta",
			Yes,
		},
		{
			"DiNo Merlin - Kad si rekla da me voliš (Official Audio) [2000]",
			"DiNo Merlin",
			"Kad si rekla da me voliš",
			"DiNo Merlin",
			Yes,
		},
		{
			"Steve Angello & Laidback Luke Feat. Robin S – Show Me Love (Official HD Video) [2009]",
			"Ministry of Sound",
			"Show Me Love",
			"Steve Angello",
			Maybe,
		},
		{
			"Logic - Wu Tang Forever ft. Wu Tang Clan (Official Audio)",
			"Visionary Music Group",
			"Wu Tang Forever ft. Wu Tang Clan",
			"Logic",
			Yes,
		},
		{
			"Sting - What Could Have Been | Arcane League of Legends | Riot Games Music",
			"Riot Games Music",
			"What Could Have Been Arcane League of Legends Riot Games Music",
			"Sting",
			Maybe,
		},
		{
			"TECHNO MIX 2021 | DJD3",
			"DJD3",
			"TECHNO MIX 2021 DJD3",
			"DJD3",
			No,
		},
		{
			"OFFICIAL Somewhere over the Rainbow - Israel IZ Kamakawiwoʻole",
			"Mountain Apple Company Inc",
			"Israel IZ Kamakawiwoʻole",
			"OFFICIAL Somewhere over the Rainbow",
			Yes,
		},
		{
			`Wu-Tang Clan - C.R.E.A.M. (Official HD Video)`,
			`Wu-Tang Clan`,
			`C.R.E.A.M.`,
			`Wu-Tang Clan`,
			Yes,
		},
		{
			`Calle`,
			`El Mola - Topic`,
			`Calle`,
			`El Mola`,
			No,
		},
	}

	for _, song := range testVideos {
		video := ParseMetadata(song.videoTitle, song.videoAuthor)
		require.Equal(t, song.title, video.Title)
		require.Equal(t, song.artist, video.Artist)
		require.Equal(t, song.reliable, video.Reliable)
	}
}
