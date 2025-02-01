# Video2mp3 is an open-source project that converts video files (e.g., .mp4, .avi, .mov) to MP3 audio format, with the option to extract and embed a cover image from the video.

##Features:

Video to MP3 Conversion: Extracts audio from video files and saves them as MP3.
Cover Image: Automatically extracts the first frame of the video as a cover image for the MP3 file.
Batch Processing: Supports batch conversion of multiple video files in a directory.
Command-Line Interface: Provides simple batch scripts or Python scripts for easy usage.



# FFmpeg Installation Guide

Due to FFmpeg's licensing requirements, this software does not include FFmpeg directly. Please follow these steps to download and install FFmpeg:

## 1. Directory Structure
├── ffmpeg/
│ └── bin/
│ ├── ffmpeg.exe # Windows
│ └── ffprobe.exe # Windows
│ ├── ffmpeg # Linux/Mac
│ └── ffprobe # Linux/Mac




## 2. Download Links

### Windows Users
1. Visit: [FFmpeg Windows Builds](https://github.com/BtbN/FFmpeg-Builds/releases)
2. Download: `ffmpeg-master-latest-win64-gpl.zip`
3. Extract and copy `ffmpeg.exe` and `ffprobe.exe` from the `bin` folder to `ffmpeg/bin/`

### MacOS Users
1. Visit: [FFmpeg macOS Builds](https://evermeet.cx/ffmpeg/)
2. Download `ffmpeg` and `ffprobe`
3. Copy to `ffmpeg/bin/` and add execute permissions

### Linux Users
1. Visit: [FFmpeg Linux Builds](https://johnvansickle.com/ffmpeg/)
2. Download: `ffmpeg-release-amd64-static.tar.xz`
3. Extract and copy `ffmpeg` and `ffprobe` to `ffmpeg/bin/`

## 3. Verify Installation

Run the following command to verify the installation:

Windows
ffmpeg\bin\ffmpeg.exe -version
MacOS/Linux
./ffmpeg/bin/ffmpeg -version