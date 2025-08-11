# r34-go
a Rule34 CLI downloader written in go

### Instalation

1. Download Latest [Release](https://github.com/moxi-git/r34-go/releases)

2. Extract it

3. run it
   * **on linux**
   ```bash
   ./r34downloader_linux
   ```

   * **on windows**
     
     pwsh
   ```powershell
   ./r34downloader.exe
   ```
   or in cmd
   ```cmd
   .\r34downloader.exe
   ```
### Usage
```
Usage:
  r34downloader [flags]
  r34downloader [command]

Available Commands:
  check       Check if content exists for given tags
  completion  Generate the autocompletion script for the specified shell
  config      Show current configuration
  help        Help about any command

Flags:
  -a, --api               Use API method (faster) instead of HTML parsing (default true)
      --gifs              Download GIFs (default true)
  -h, --help              help for r34downloader
      --images            Download images (default true)
      --no-gifs           Don't download GIFs
      --no-images         Don't download images
      --no-videos         Don't download videos
  -o, --output string     Output directory (default "./downloads")
  -q, --quantity uint16   Number of items to download (default 100)
  -t, --tags string       Tags to search for (required)
      --videos            Download videos (default true)

Use "r34downloader [command] --help" for more information about a command.
```

### Building 
**Windows (powershell)**
```powershell
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -o r34downloader.exe
```

**Building on Windows for linux **
```powershell
$env:CGO_ENABLED="0"; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -o r34downloader_linux
```

**Linux (bash)**
```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o r34downloader_linux
```

### License
This project is under [MIT License](https://github.com/moxi-git/r34-go/blob/main/LICENSE)

### Contributing
Idk sumbit [Pull Requests](https://github.com/moxi-git/r34-go/pulls) or sumbit an Issue

### Issues
Submit issues [Here!](https://github.com/moxi-git/r34-go/issues)

**Inspired by**:

[Rule34.xxx Downloader by DaxEleven](https://github.com/DaxEleven/Rule34.xxx-Downloader) (windows only btw and that's the issue)

PS. Not assosiated with Rule34.xxx in any way i just made a tool to download alot of R34 yk.
