# anonymizer

Anonymizes AWS Configuration Item JSON files by redacting sensitive fields such as Account IDs, ARNs, resource names, tags, and more.

## Features
- Redacts AWS Account IDs, ARNs, resource IDs, resource names, and tags
- Supports single JSON objects or arrays of objects
- Handles nested fields (e.g., configuration, relationships)
- CLI options for input/output files, pretty-printing, and dry-run

## Usage

```bash
go run main.go --input <input.json> [--output <output.json>] [--dry-run] [--pretty=false]
```

### Flags
- `--input` (required): Path to input AWS Configuration Item JSON file
- `--output`: Optional path to save anonymized JSON (default: stdout)
- `--dry-run`: Print anonymized JSON to stdout without saving
- `--pretty`: Pretty-print JSON output (default: true)

## Example

```bash
go run main.go --input config.json --output anonymized.json
```

## How It Works
Sensitive fields such as `AccountId`, `ResourceId`, `ResourceName`, `ARN`, and `Tags` are replaced with generic values. ARNs are parsed and the Account ID and Resource portion are redacted. Nested fields and arrays are recursively anonymized.
