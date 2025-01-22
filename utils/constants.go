package utils

import (
	"sync"

	"github.com/grafov/m3u8"
)

const (
	contentSegmentFolder   = "./segments"
	adSegmentFolder        = "./adv/ad2"
	manifestFileFolder     = "./manifests"
	masterManifestFilePath = "./manifests/master.m3u8"
	movingWindowSize       = 5
	updateInterval         = 1
	inputPlaylistPath      = "./input-playlist.m3u8"
)

var (
	contentSegmentCache []string
	adSegmentCache      = make(map[string][]string)
	lock                sync.Mutex
	globalPlaylists     = make(map[string]*m3u8.MediaPlaylist)
	totalSegments       int
	scteMarkers         = map[int]int{}
)

var resolutions = []struct {
	Folder     string
	Manifest   string
	Bandwidth  int
	Resolution string
}{
	{"360p", "360p.m3u8", 800000, "640x360"},
	{"480p", "480p.m3u8", 1200000, "854x480"},
	{"720p", "720p.m3u8", 3000000, "1280x720"},
}
