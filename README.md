# sing-ruleset

A CLI tool designed to download and convert various blocklists and rulesets into the [sing-box](https://github.com/SagerNet/sing-box) rule-set format (`.srs`).

## Features

- **Adguard Support**: Converts Adguard-compatible blocklists to sing-box rule sets.
- **IP List Support**: Processes raw IP text lists into sing-box rule sets.
- **Concurrency**: Parallel processing of rule sources for fast updates.
- **Configurable**: Fully data-driven configuration via `config.json`.
- **CLI Interface**: Easy-to-use command line interface with Cobra.

## Installation

You can build the project from source using Go:

```bash
git clone https://github.com/chaosoffire/sing-ruleset.git
cd sing-ruleset
go build -o sing-ruleset .
```

## Usage

The application uses the `run` command to start the generation process.

```bash
./sing-ruleset run [flags]
```

### Flags

| Flag        | Shorthand | Description                           | Default           |
| ----------- | --------- | ------------------------------------- | ----------------- |
| `--workdir` | `-d`      | Base working directory                | `.` (Current Dir) |
| `--output`  | `-o`      | Output directory for `.srs` files     | `output`          |
| `--config`  | `-c`      | Path to configuration file            | `config.json`     |
| `--workers` | `-w`      | Number of concurrent download workers | System CPU Count  |

### Examples

Run with default settings (looks for config.json in current dir):

```bash
./sing-ruleset run
```

Run with custom config and output directory:

```bash
./sing-ruleset run -c rules.json -o ./dist
```

## Configuration

The application relies on a JSON configuration file (default `config.json`). It organizes sources by category.

### Structure

```json
{
  "version": "1",
  "sources": {
    "adguard": [
      {
        "name": "EasyList",
        "url": "https://easylist.to/easylist/easylist.txt"
      }
    ],
    "iplists": [
      {
        "name": "CINS_army",
        "url": "https://cinsarmy.com/list/ci-badguys.txt",
        "type": "iplist"
      }
    ]
  }
}
```

### Source Types

1. **adguard**:
   - Standard Adguard/EasyList filter format.
   - Converted directly to SRS rules.

2. **iplists**:
   - Primitive text files with IPs (CIDR or single IP).
   - Must specify `"type": "iplist"` in the config object.
   - Converted to sing-box IP-cidr rule sets.

## Disclaimer

This project is primarily for personal use.
