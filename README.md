# YouTube to MP3 Converter :notes:
YouTube :film_strip: to MP3 :musical_note: converter terminal app. Application is dependant on [FFMPEG](https://ffmpeg.org/) software for MP4 to MP3 conversion. Firstly, make sure you download it and add it to `PATH`. 

_**Disclaimer**: it contains a lot of bugs, is user unfriendly and feels like flying a spaceship. This application is only for academic purposes, please don't sue me._

## Usage
Acceptable source arguments are YouTube playlists that are **Public** or **Unlisted**, and text files with links, each on a separate line. You can limit amount of MP3s downloaded using `-n_links=42` and `-skip=42` flag.
```bash
> yt2mp3 -n_links={number} -skip={number} {source} {output_folder} 
```
![yt2mp3](https://user-images.githubusercontent.com/36798549/209480711-a7930ec4-2984-45b2-b158-6dc448d7dee1.gif)
