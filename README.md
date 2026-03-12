# devstats

A Go library for collecting and storing developer statistics and metrics.

⚠️⚠️⚠️
This is very much a work in progress, so expect breaking changes

## Features

- 💾 SQLite storage for persistent data
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

This will create the following files in the current folder:
- `devstats.db` - Contains raw keystroke and file change data
- `devstats_anon.db` - Contains anonymized statistics for sharing 


