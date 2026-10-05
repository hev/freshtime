# freshtime

A CLI tool for managing OAuth tokens and interacting with time-tracking APIs.

## Workflow

A typical consulting engagement, start to finish:

```bash
freshtime setup                                  # one-time OAuth with FreshBooks
freshtime clients create --org "Acme Inc"        # new client (skip if existing)
freshtime project create --name "Acme SOW" \
  --client acme --rate 250 --budget 100h --init  # project from the SOW; --init writes .freshtime.json
freshtime log -m "kickoff call" -d 1h            # log time (uses .freshtime.json defaults)
freshtime log -m "design review" -d 2h --date 2026-06-08   # backdate an entry
freshtime start -m "pairing session"             # or run a timer...
freshtime stop                                   # ...and log it on stop
freshtime list                                   # recent entries with IDs
freshtime edit 12345 -d 1h30m                    # fix an entry
freshtime delete 12345                           # or remove it
freshtime weekly                                 # hours by client this week
freshtime invoice <client-id>                    # invoice unbilled time
freshtime invoice open                           # invoices with money outstanding
freshtime invoice paid 0000044 --dry-run         # record a payment (full balance, today, ACH by default)
```

Time with no client (R&D, bizdev, admin) goes on an internal project:

```bash
freshtime project create --internal --name "R&D"
freshtime log -m "retrieval benchmark" -d 2h --project <id> --no-billable
```

`weekly` and `list` show it as "Internal".

## Build

```bash
go build -o freshtime ./cmd/freshtime
```

## Install

```bash
go install ./cmd/freshtime
```

## Run

```bash
freshtime --help
```

## Test

```bash
go test ./...
```
