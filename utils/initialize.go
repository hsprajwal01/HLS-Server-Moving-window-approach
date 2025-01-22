package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/grafov/m3u8"
)

// InitializePlaylists creates media playlists for each resolution and stores them in the global cache.
func InitializePlaylists() {
	for _, res := range resolutions {
		playlist, err := m3u8.NewMediaPlaylist(uint(movingWindowSize), uint(movingWindowSize))
		if err != nil {
			log.Printf("[ERROR] Failed to create playlist for %s: %v", res.Folder, err)
			continue
		}
		globalPlaylists[res.Folder] = playlist
		log.Printf("[INFO] Initialized playlist for resolution %s", res.Folder)
	}
}

// InitializeAdContentCache caches the available ad segments for each resolution.
func InitializeAdContentCache() {
	for _, res := range resolutions {
		segmentDir := filepath.Join(adSegmentFolder, res.Folder)
		segments, err := os.ReadDir(segmentDir)
		if err != nil {
			log.Printf("[ERROR] Failed to read ad segments from %s: %v", segmentDir, err)
			continue
		}

		var segmentNames []string
		for _, seg := range segments {
			if strings.HasSuffix(seg.Name(), ".ts") {
				segmentNames = append(segmentNames, seg.Name())
			}
		}

		sortSegments(segmentNames) // Sort the ad segments
		adSegmentCache[res.Folder] = segmentNames
		log.Printf("[INFO] Cached %d ad segments for resolution %s", len(segmentNames), res.Folder)
	}
}

// parseInputPlaylist parses the input playlist, identifies content segments, and detects SCTE markers.
func parseInputPlaylist() {
	file, err := os.Open(inputPlaylistPath)
	if err != nil {
		log.Fatalf("[FATAL] Failed to open input playlist: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var segmentIndex int
	var currentSCTEDuration int
	var scteMarkerActive bool

	for scanner.Scan() {
		line := scanner.Text()

		// Detect SCTE markers
		if strings.HasPrefix(line, "#EXT-X-CUE-OUT") {
			currentSCTEDuration = parseSCTEDuration(line)
			scteMarkerActive = true
			log.Printf("[INFO] Detected SCTE CUE-OUT at segment %d with duration %d seconds", segmentIndex, currentSCTEDuration)
		} else if strings.HasPrefix(line, "#EXTINF") {
			// The next line should contain the segment name
			if scanner.Scan() {
				segmentName := scanner.Text()
				contentSegmentCache = append(contentSegmentCache, segmentName)

				if scteMarkerActive {
					// Mark the segment for SCTE ad duration
					scteMarkers[segmentIndex] = currentSCTEDuration
					scteMarkerActive = false // Reset after marking the segment
				}

				segmentIndex++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("[FATAL] Error reading input playlist: %v", err)
	}

	totalSegments = segmentIndex

	log.Printf("[INFO] Parsed %d total segments", segmentIndex)
	log.Printf("[INFO] Detected SCTE markers: %v", scteMarkers)
	log.Printf("[INFO] Cached content segments: %v", contentSegmentCache)
}

// parseSCTEDuration extracts the SCTE duration from the tag string.
// Returns 0 if the format is invalid.
func parseSCTEDuration(tag string) int {
	parts := strings.Split(tag, "DURATION=")
	if len(parts) < 2 {
		log.Printf("[WARNING] Invalid SCTE tag format: %s", tag)
		return 0
	}
	durationStr := strings.Split(parts[1], ",")[0]
	var duration int
	fmt.Sscanf(durationStr, "%d", &duration)
	return duration
}

// Initialize performs all necessary initialization steps for the service.
func Initialize() {
	log.Println("[INFO] Starting initialization process...")
	parseInputPlaylist()
	InitializePlaylists()
	InitializeAdContentCache()
	generateInitialManifests()
	log.Println("[INFO] Initialization process completed.")
}
