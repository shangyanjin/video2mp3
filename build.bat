@echo off
mkdir assets 2>nul

REM Check if ffmpeg exists in the correct location
if not exist "ffmpeg\bin\ffmpeg.exe" (
    echo Error: ffmpeg.exe not found in ffmpeg\bin directory
    exit /b 1
)
if not exist "ffmpeg\bin\ffprobe.exe" (
    echo Error: ffprobe.exe not found in ffmpeg\bin directory
    exit /b 1
)

REM Copy FFmpeg files to assets
copy /Y "ffmpeg\bin\ffmpeg.exe" "assets\ffmpeg.exe"
copy /Y "ffmpeg\bin\ffprobe.exe" "assets\ffprobe.exe"

REM Build the program
go build -ldflags="-s -w" -o video2mp3.exe

REM Check if build was successful
if errorlevel 1 (
    echo Build failed
    exit /b 1
)

echo Build completed successfully 