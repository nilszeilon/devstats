# devstats

A Go library for collecting and storing developer statistics and metrics.

⚠️⚠️⚠️
This is very much a work in progress, so expect breaking changes

## Features

- 💾 Flexible data storage (json or sqlite)
- ⌨️  Keypress tracking (background macOS support)
- 📊 Language tracking (keeps track of file changes)
- 🔒 Automatic anonymization of your data (don't send ALL your keystrokes to some server)

## How to run


```bash
go mod tidy
```

I run the repository as a background process

```bash
go run cmd/cli/main.go & 
```

but you can just as well run it as long as the window is open

```bash
go run cmd/cli/main.go 
```

This will save the files keypresses.json & filchanges.json in the current folder. 


