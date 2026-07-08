# yt-gif

`yt-gif` is a simple CLI tool for creating GIFs from YouTube videos.

## Requirements

- [yt-dlp](https://github.com/yt-dlp/yt-dlp) in `PATH`
- [ffmpeg](https://ffmpeg.org/) in `PATH`

## Installation

If you have Go installed on your machine:
```text
go install github.com/Kori-ts/yt-gif@latest
```

Windows with WinGet:
```text
winget install Kori.YTGIF
```

Linux/macOS with Homebrew:
```text
brew install Kori-ts/yt-gif/yt-gif
```

Check dependencies:
```text
yt-gif check
```

## Usage

```text
yt-gif <url> <start-end>... [options]
```

Example:
```text
yt-gif https://www.youtube.com/watch?v=dQw4w9WgXcQ 0:00-0:07.364
```

Options:
- `--out <dir>`: Output directory. Defaults to the current directory.
- `--force`: Overwrite existing `clip-{n}.gif` files.
- `--jobs <n>`: Parallel jobs. Defaults to `2`.
- `--fps <n|source>`: GIF framerate. Must be a numeric value in the range `1-60`. Defaults to `15`.
- `--width <px|source>`: Output width. Use `source` to leave width unconstrained. Defaults to `640`.
- `--height <px|source>`: Output height. Use `source` to leave height unconstrained.
- `--timing`: Show download, conversion, and total elapsed times.
- `--accurate-cut`: Force exact section cuts in `yt-dlp`. Slower.
- `--verbose`: Show `yt-dlp` and `ffmpeg` output with job prefixes.
- `--help`: Display usage information.

## YouTube Precise Timestamp Userscript

To get more precise YouTube timestamps, use the [YouTube Precise Timestamps](https://github.com/Kori-ts/yt-precise-timestamps) user script.
