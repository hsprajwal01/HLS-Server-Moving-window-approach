package cmd

import (
	"app/utils"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

// Port for the HTTP server, default is set to 8084
var port int

// startCmd represents the "start" command to initialize the HLS streaming server
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the HLS streaming server",
	Long:  `Start the HLS streaming server with dynamic ad insertion and manifest generation.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize required utilities and configurations
		utils.Initialize()

		// Start a background goroutine to update manifests periodically
		go utils.UpdateManifestsPeriodically()

		// Define HTTP handlers for serving manifests and segments
		http.HandleFunc("/vod/master.m3u8", utils.ServeMasterManifest)
		http.HandleFunc("/vod/", utils.ServeSegmentOrManifest)

		// Build the address string and start the server
		address := fmt.Sprintf(":%d", port)
		log.Printf("Server running at http://localhost%s", address)

		// Log and handle any errors from the HTTP server
		log.Fatal(http.ListenAndServe(address, nil))
	},
}

func init() {
	// Define the --port or -p flag for specifying the server port
	startCmd.Flags().IntVarP(&port, "port", "p", 8084, "Port on which the server will run")

	// Add the startCmd to the root command
	rootCmd.AddCommand(startCmd)
}
