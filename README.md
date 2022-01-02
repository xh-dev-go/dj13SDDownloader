# dj13SDDownloader
command line downloader sequence diagram from https://sequence.davidje13.com/

## Download

[Download Link](https://github.com/xh-dev-go/dj13SDDownloader/blob/master/bin/dj13SDDownloader.exe?raw=true)
```shell
curl https://github.com/xh-dev-go/dj13SDDownloader/blob/master/bin/dj13SDDownloader.exe?raw=true --output dj13SDDownloader.exe
```

## Usage

### Check version
```shell
dj13SDDownloader.exe --version
```
### Help 
```shell
dj13SDDownloader.exe -h
```

### From clipboard
Suit for complete modifying the diagram and then copy the script and run the program for local persisting version.
```shell
# dj13SDDownloader.exe --from-clipboard --persist --output-file --naming-pattern {file pattern without extension}
# e.g.
dj13SDDownloader.exe --from-clipboard --persist --output-file --naming-pattern ./diagrams/abc
```
1. Extract the dsl script from clipboard, 
2. Download the svg file as file ./diagrams/abc.svg
3. Store the script to file ./diagrams/abc.sddsl


### From file
Suit for download svg file when no image on hand.
```shell
# dj13SDDownloader.exe --from-file --persist --output-file --naming-pattern {file pattern without extension}
# e.g.
dj13SDDownloader.exe --from-file --persist --output-file --naming-pattern ./diagrams/abc
```
1. Extract the dsl script from ./diagram/abc.sddsl
2. Download the svg file as file ./diagrams/abc.svg
3. Store the script to file ./diagrams/abc.sddsl
