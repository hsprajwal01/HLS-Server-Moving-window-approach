# HLS Server (Moving window approach)

Overview

 This application is an HLS server utilizing a linear moving window approach. It parses the input playlist to identify SCTE markers and dynamically inserts advertisements to match the specified duration of time indicated by these markers.

Dynamic Ad Insertion: Supports SCTE markers to insert ads dynamically into video streams.

Manifest Management: Generates and updates HLS manifests for various resolutions.

Multi-Resolution Support: Serves content in multiple resolutions (360p, 480p, 720p) with appropriate bitrate settings.

Input Playlist Parsing: Parses the input playlist, identifies SCTE markers, and matches ad segments to the specified durations.

## Directory Structure

```plaintext
app/
├── adv/                       # Advertisement segments
│   ├── ad1/                   # ad1 segments present here
│   ├── ad2/                   # ad2 segments present here
│   └── ad3/                   # ad3 segments present here
├── cmd/                       # Command-line commands
│   ├── root.go                # Root command definition
│   └── start.go               # "start" command implementation
├
├── manifests/                 # HLS manifest files for master and each varients
│   ├── master.m3u8            # Master playlist
│   ├── 360p.m3u8              # 360p playlist
│   ├── 480p.m3u8              # 480p playlist
│   └── 720p.m3u8              # 720p playlist
|
├── segments/                  # Video content segments for each resolution
│   ├── 360p/
│   ├── 480p/
│   └── 720p/
├── utils/                     # Utility functions
│   ├── constants.go           # Constants for the application
│   ├── initialize.go          # Initialization logic
│   ├── manifests-upd.go       # Manifest update logic
│   └── utils.go
├── input-playlist.m3u8        # input playlist
├── go.mod                     # Go module dependencies
├── go.sum                     # Go dependency checksums
└── main.go                    # Entry point of the application
```

### HTTP Server
- Serves video segments and manifests via HTTP.
- Default port: `8084` (configurable with the `--port` flag).

## Getting Started

### Prerequisites
- [Go](https://golang.org/) (version 1.16 or later)


### Installation

1. Clone the repository:
   ```bash
   git clone <repository_url>
   cd app
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

### Running the Application

1. **Start the Server**:
   ```bash
   go run main.go start --port 8084
   ```
or 

   ```bash
   ./hls-moving-window-app start --port 8080
   ```

2. **Access the Server**:
   - Master playlist: `http://localhost:8084/vod/master.m3u8`
   - Segments and other resources: `http://localhost:8084/vod/<resource>`


## Code Overview

### Main Components

1. **cmd/root.go**:
   - Defines the root command for the CLI.
   - Initializes the command structure and subcommands.

2. **cmd/start.go**:
   - Implements the "start" command.
   - Starts the HTTP server and initializes utilities.

3. **utils/initialize.go**:
   - Parses input playlists.
   - Initializes media and ad playlists.

4. **utils/manifests-upd.go**:
   - Handles the logic for updating manifests periodically.

5. **utils/constants.go**:
   - Stores configuration constants like file paths and update intervals.

### Key Constants
- `contentSegmentFolder`: Path to content segments.
- `adSegmentFolder`: Path to ad segments.
- `manifestFileFolder`: Path to manifest files.
- `movingWindowSize`: Number of segments retained in the playlist.
- `updateInterval`: Interval for manifest updates (in seconds).

## Customization

- Add new resolutions by modifying the `resolutions` array in `utils/constants.go`.
- Extend ad content by adding new segments in the `adv/` folder.
- The input-playlist.m3u8 serves as the primary input, containing SCTE markers that define ad insertion points and their durations.
- Segments for each resolution are stored in the segments/ folder, where transcoded video segments for 360p, 480p, and 720p resolutions can be found.



