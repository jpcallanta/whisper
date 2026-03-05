# whisper

CLI to migrate, verify, and dump AWS Secrets Manager secrets between regions.

## Requirements

- Go 1.25+
- AWS credentials (environment variables or default credential chain)

## Install

```bash
go install
```

Or build a binary:

```bash
go build -o whisper .
```

### Makefile

From the project root you can use:

| Target | Command | Description |
|--------|---------|-------------|
| build | `make` or `make build` | Build the `whisper` binary. |
| test | `make test` | Run tests. |
| run | `make run` or `make run ARGS="migrate --help"` | Build and run; pass subcommand/args via `ARGS`. |
| install | `sudo make install` | Install the binary to `/usr/local/bin`. |
| clean | `make clean` | Remove the built binary and test artifacts. |

## Usage

### Migrate secrets

Copy secrets from a source region to a destination region. Omit `--secret-id` to migrate all secrets.

```bash
whisper migrate --source-region <region> --dest-region <region> [--secret-id <name-or-arn>]
```

### Verify secrets

Compare secret names in source and destination; optionally add secrets missing in the target.

```bash
whisper verify --source-region <region> --dest-region <region> [--add-missing]
```

### Dump secrets

Export all secrets in a region to a CSV file (columns: Name, ARN, SecretString, SecretBinary). Omit `--output` to write to stdout.

```bash
whisper dump --region <region> [--output secrets.csv]
```

## Testing

```bash
go test ./...
```

Or: `make test`.
