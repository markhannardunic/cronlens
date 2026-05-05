# cronlens

Lightweight cron job monitor that surfaces failures and runtimes via a local dashboard.

## Installation

```bash
go install github.com/yourusername/cronlens@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/cronlens.git && cd cronlens && go build ./...
```

## Usage

Wrap your cron command with `cronlens exec` to start tracking it:

```bash
cronlens exec --name "db-backup" -- /usr/local/bin/backup.sh
```

Then launch the local dashboard to view job history, failure rates, and runtimes:

```bash
cronlens dashboard
```

The dashboard is served at `http://localhost:7777` by default. You can change the port with `--port`:

```bash
cronlens dashboard --port 8080
```

### Example crontab entry

```cron
0 2 * * * cronlens exec --name "nightly-backup" -- /usr/local/bin/backup.sh
*/5 * * * * cronlens exec --name "health-check" -- /usr/local/bin/healthcheck.sh
```

## Configuration

cronlens stores job data in `~/.cronlens/data.db` by default. Override with the `CRONLENS_DATA_DIR` environment variable.

```bash
export CRONLENS_DATA_DIR=/var/lib/cronlens
```

## Requirements

- Go 1.21+
- Linux / macOS

## License

MIT © 2024 [Your Name](https://github.com/yourusername)