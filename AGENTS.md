# Freshtime Agent Guide

Freshtime is a Go CLI project using Cobra. This file is for engineering and
operations guidance. For product/design context, read `CLAUDE.md`.

## Commands

```sh
go build -o freshtime ./cmd/freshtime
go test ./...
go install ./cmd/freshtime
```

## Project Structure

- `cmd/freshtime/`: main entry point.
- `internal/`: internal packages for API, commands, config, and formatting.
