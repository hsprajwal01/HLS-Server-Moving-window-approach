package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/grafov/m3u8"
)

// logRequest logs details of an HTTP request for debugging and monitoring purposes.
func logRequest(r *http.Request) {
	log.Printf("[INFO] HTTP Request: Method=%s, URL=%s, RemoteAddr=%s, Headers=%v",
		r.Method, r.URL.Path, r.RemoteAddr, r.Header)
}

// sortSegments sorts a list of segment filenames based on their segment number.
func sortSegments(segments []string) {
	sort.Slice(segments, func(i, j int) bool {
		return extractSegmentNumber(segments[i]) < extractSegmentNumber(segments[j])
	})
}

// extractSegmentNumber extracts and returns the numerical part of a segment filename.
// If the filename format is invalid, it returns 0.
func extractSegmentNumber(segment string) int {
	parts := strings.Split(segment, "_")
	if len(parts) < 2 {
		log.Printf("[WARNING] Invalid segment format: %s", segment)
		return 0
	}
	number := strings.TrimSuffix(parts[1], ".ts")
	num, err := strconv.Atoi(number)
	if err != nil {
		log.Printf("[ERROR] Failed to extract segment number from %s: %v", segment, err)
		return 0
	}
	return num
}

// placeAdbreak determines whether to insert an ad at a given segment index.
// Returns a flag indicating if an ad should be added, the number of ad segments, and the total ad duration.
func placeAdbreak(nextSegmentIndex int) (bool, int, float64) {
	adDuration, shouldAdd := scteMarkers[nextSegmentIndex] // Retrieve ad schedule details
	if shouldAdd {
		numAdSegments := adDuration / 5 // Assuming each ad segment is 5 seconds
		log.Printf("[INFO] Ad scheduled for segment index %d: Duration=%d seconds", nextSegmentIndex, adDuration)
		return true, numAdSegments, float64(adDuration)
	}
	return false, 0, 0
}

// insertAd handles the insertion of ad segments into the media playlist.
func insertAd(folder, manifest string, adInserted, numAdsToInsert int, adDuration float64, sequenceNumber int) {
	// Retrieve cached ad segments
	segments, cached := adSegmentCache[folder]
	if !cached || len(segments) == 0 {
		log.Printf("[WARNING] No cached ad segments for folder %s. Skipping ad insertion.", folder)
		return
	}

	// Retrieve or create the media playlist
	playlist, exists := globalPlaylists[folder]
	if !exists {
		maxSegments := uint(movingWindowSize)
		var err error
		playlist, err = m3u8.NewMediaPlaylist(maxSegments, maxSegments)
		if err != nil {
			log.Printf("[ERROR] Failed to create M3U8 media playlist for folder %s: %v", folder, err)
			return
		}
		globalPlaylists[folder] = playlist
		log.Printf("[INFO] New media playlist created for folder %s", folder)
	}

	// Adjust playlist to maintain sliding window
	if playlist.Count() >= movingWindowSize {
		updatePlaylistWindow(playlist)
	}

	// Insert ad segments into the playlist
	if adInserted < numAdsToInsert {
		newAdSegment := segments[adInserted]
		uri := fmt.Sprintf("/vod/adv/ad2/%s/%s", folder, newAdSegment)
		duration := 5.0 // Each ad segment is 5 seconds
		playlist.Append(uri, duration, "")

		log.Printf("[INFO] Ad segment inserted: %s into folder %s", newAdSegment, folder)
	}

	// Update discontinuity markers
	updateDiscontinuityMarkers(playlist, adInserted, numAdsToInsert)

	// Save the updated playlist to a file
	savePlaylistToFile(playlist, manifest, sequenceNumber)
}

// updatePlaylistWindow adjusts the playlist to maintain a sliding window of segments.
func updatePlaylistWindow(playlist *m3u8.MediaPlaylist) {
	newPlaylist, _ := m3u8.NewMediaPlaylist(uint(movingWindowSize), uint(movingWindowSize))
	for i := 1; i < int(playlist.Count()); i++ {
		seg := playlist.Segments[i]
		if seg != nil {
			newPlaylist.Append(seg.URI, seg.Duration, "")
			if seg.Discontinuity {
				_ = newPlaylist.SetDiscontinuity()
			}
		}
	}
	newPlaylist.DiscontinuitySeq = playlist.DiscontinuitySeq
	newPlaylist.SetDiscontinuityNext(playlist.DiscontinuityNext())
	*playlist = *newPlaylist
	log.Printf("[INFO] Playlist window updated. Old segments removed.")
}

// updateDiscontinuityMarkers manages discontinuity markers during ad insertion.
func updateDiscontinuityMarkers(playlist *m3u8.MediaPlaylist, adInserted, numAdsToInsert int) {
	if adInserted == 0 {
		playlist.SetDiscontinuity()
		log.Printf("[INFO] Discontinuity marker set at the start of ad insertion.")
	}
	if adInserted == numAdsToInsert-1 {
		playlist.SetDiscontinuityNext(true)
		log.Printf("[INFO] Discontinuity marker set at the end of ad insertion.")
	}
}

// savePlaylistToFile writes the updated playlist to the specified file path.
func savePlaylistToFile(playlist *m3u8.MediaPlaylist, manifest string, sequenceNumber int) {
	playlist.SeqNo = uint64(sequenceNumber + 1)
	manifestPath := filepath.Join(manifestFileFolder, manifest)
	file, err := os.Create(manifestPath)
	if err != nil {
		log.Printf("[ERROR] Failed to create manifest file at %s: %v", manifestPath, err)
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(playlist.Encode().String()))
	if err != nil {
		log.Printf("[ERROR] Failed to write M3U8 playlist to file %s: %v", manifestPath, err)
		return
	}

	log.Printf("[INFO] Manifest file %s updated successfully.", manifestPath)
}

// ServeMasterManifest serves the master playlist manifest to clients.
func ServeMasterManifest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	http.ServeFile(w, r, masterManifestFilePath)
	log.Printf("[INFO] Master manifest served: %s", masterManifestFilePath)
}

// ServeSegmentOrManifest serves either a segment or manifest file based on the request URL.
func ServeSegmentOrManifest(w http.ResponseWriter, r *http.Request) {
	logRequest(r)
	if strings.HasSuffix(r.URL.Path, ".m3u8") {
		manifestPath := filepath.Join(manifestFileFolder, strings.TrimPrefix(r.URL.Path, "/vod/"))
		http.ServeFile(w, r, manifestPath)
		log.Printf("[INFO] Manifest served: %s", manifestPath)
	} else {
		segmentPath := filepath.Join(contentSegmentFolder, strings.TrimPrefix(r.URL.Path, "/vod/"))
		http.ServeFile(w, r, segmentPath)
		log.Printf("[INFO] Segment served: %s", segmentPath)
	}
}
