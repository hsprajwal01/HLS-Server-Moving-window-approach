package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/grafov/m3u8"
)

// generateInitialManifests creates initial manifest files for each resolution.
func generateInitialManifests() {
	for _, res := range resolutions {
		playlist, exists := globalPlaylists[res.Folder]
		if !exists {
			log.Printf("[WARNING] Playlist for resolution %s not found. Skipping manifest generation.", res.Folder)
			continue
		}

		// Populate the playlist with initial segments
		for i := 0; i < movingWindowSize && i < len(contentSegmentCache); i++ {
			segmentName := contentSegmentCache[i]
			uri := fmt.Sprintf("/vod/%s/%s", res.Folder, segmentName)
			duration := 10.0
			err := playlist.Append(uri, duration, "")
			if err != nil {
				log.Printf("[ERROR] Failed to append segment %s to playlist for resolution %s: %v", segmentName, res.Folder, err)
			}
		}

		writeManifest(res.Manifest, playlist)
		log.Printf("[INFO] Initial manifest generated for resolution %s.", res.Folder)
	}
}

// writeManifest saves the M3U8 playlist to the specified file.
func writeManifest(manifestName string, playlist *m3u8.MediaPlaylist) {
	manifestPath := filepath.Join(manifestFileFolder, manifestName)
	file, err := os.Create(manifestPath)
	if err != nil {
		log.Printf("[ERROR] Failed to create manifest file %s: %v", manifestName, err)
		return
	}
	defer file.Close()

	_, err = file.Write([]byte(playlist.Encode().String()))
	if err != nil {
		log.Printf("[ERROR] Failed to write M3U8 playlist to file %s: %v", manifestName, err)
		return
	}
	log.Printf("[INFO] Manifest %s generated successfully.", manifestName)
}

// UpdateManifestsPeriodically updates the manifests at regular intervals.
func UpdateManifestsPeriodically() {
	ticker := time.NewTicker(time.Duration(updateInterval) * time.Second)
	defer ticker.Stop()
	sequenceNumber := 0
	nextContentSegmentIndex := movingWindowSize
	addInserted := 0

	// Allow system to stabilize before starting updates
	time.Sleep(5 * time.Second)

	for range ticker.C {
		lock.Lock()

		shouldAddAdBreaks, totalAdSegmentsToAdd, adDuration := placeAdbreak(nextContentSegmentIndex % totalSegments)

		if shouldAddAdBreaks && addInserted < totalAdSegmentsToAdd {
			for _, res := range resolutions {
				insertAd(res.Folder, res.Manifest, addInserted, totalAdSegmentsToAdd, adDuration, sequenceNumber)
			}
			addInserted++
		} else {
			for _, res := range resolutions {
				addInserted = 0
				updateResolutionManifest(res.Folder, res.Manifest, sequenceNumber, nextContentSegmentIndex)
			}
			nextContentSegmentIndex++
		}
		sequenceNumber++

		lock.Unlock()
	}
}

// updateResolutionManifest updates the manifest for a specific resolution.
func updateResolutionManifest(folder, manifest string, sequenceNumber, nextSegmentIndex int) {
	segments := contentSegmentCache
	if len(segments) == 0 {
		log.Printf("[WARNING] No cached segments available for %s. Skipping update.", folder)
		return
	}

	// Ensure nextSegmentIndex loops within the available segments
	nextSegmentIndex %= totalSegments

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
	}

	// Adjust playlist for discontinuity and sliding window logic
	if playlist.Count() > 0 && playlist.Segments[0].Discontinuity {
		playlist.DiscontinuitySeq++
	}
	if playlist.Count() >= movingWindowSize {
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
		globalPlaylists[folder] = newPlaylist
		playlist = newPlaylist
		log.Printf("[INFO] Sliding window applied to playlist for resolution %s.", folder)
	}

	// Add the next content segment to the playlist
	newSegment := contentSegmentCache[nextSegmentIndex]
	uri := fmt.Sprintf("/vod/%s/%s", folder, newSegment)
	playlist.Append(uri, 10.0, "") // Assuming 10 seconds per segment

	// Update discontinuity markers
	if playlist.DiscontinuityNext() {
		playlist.SetDiscontinuity()
		playlist.SetDiscontinuityNext(false)
	}

	// Handle discontinuity for the first segment
	if nextSegmentIndex == 0 {
		if err := playlist.SetDiscontinuity(); err != nil {
			log.Printf("[ERROR] Failed to set discontinuity for resolution %s: %v", folder, err)
		} else {
			log.Printf("[INFO] Discontinuity added before segment: %s for resolution %s", newSegment, folder)
		}
	}

	// Save the updated playlist
	playlist.SeqNo = uint64(sequenceNumber + 1)
	writeManifest(manifest, playlist)
}

// updateMasterManifest generates and updates the master manifest file.
func updateMasterManifest() {
	masterContent := "#EXTM3U\n"
	for _, res := range resolutions {
		masterContent += fmt.Sprintf(
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n/vod/%s\n",
			res.Bandwidth, res.Resolution, res.Manifest,
		)
	}
	err := os.WriteFile(masterManifestFilePath, []byte(masterContent), 0644)
	if err != nil {
		log.Printf("[ERROR] Failed to write master manifest: %v", err)
		return
	}
	log.Printf("[INFO] Master manifest updated successfully.")
}
