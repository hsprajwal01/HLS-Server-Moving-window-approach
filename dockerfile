# Use the official Golang image as the base image
FROM golang:1.20

# Set the working directory
WORKDIR /app



COPY ./hls-moving-window-app-docker /app/hls-moving-window-app-docker
COPY ./m3u8-local /app/m3u8-local
COPY ./segments /app/segments
COPY ./input-playlist.m3u8 /app/input-playlist.m3u8
COPY . .


# Expose the port your application runs on
EXPOSE 8082

# Create necessary folders for segments and manifests
RUN mkdir -p /app/manifests

# Set environment variables for folders
ENV SEGMENT_FOLDER="/app/segments"
ENV MANIFEST_FOLDER="/app/manifests"
ENV MASTER_MANIFEST_PATH="/app/manifests/master.m3u8"

# Command to run the application
CMD ["./hls-moving-window-app-docker", "start", "--port", "8082"]

