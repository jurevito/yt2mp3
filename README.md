# YouTube to MP3 Converter
YouTube to MP3 converter TUI app.

## Dependancies
Project is currently dependant on [FFMPEG](https://ffmpeg.org/) for **mp4** to **mp3** conversion. First make sure you download it and add it to `PATH`.

```bash
go build        # build binary
go run .        # build and run source code
go test ./...   # run tests
```

### Usage
```bash
> yt2mp3 -n_links=<first_n> <source> <output_folder> 

go run . -n_links=2 "https://www.youtube.com/playlist?list=PL6YgdMS9Bn4FLSnpv368M3s3_cysoeBkT" "C:/Users/jurev/Documents/Project/yt2mp3/output/"
```