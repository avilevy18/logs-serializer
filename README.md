# Logs Serializer

A simple gRPC server that receives logs and metrics and serializes them to
either a file or standard output.

## Usage

### Flags

* `--stdout`: (optional) Print logs and metrics to standard out. If not
  provided, output will be written to files.
* `--logs`: (optional) Run only the logs server.
* `--metrics`: (optional) Run only the metrics server.
* `--all`: (optional) Run both servers (default).
* `--path`: (optional) The path to write the files to. Defaults to
  `/tmp/log-serializer`.

### Examples

**Run both servers and write to files in the default directory:**

```bash
go run .
```

**Run only the logs server and print to standard output:**

```bash
go run . --logs --stdout
```

**Run only the metrics server and write to files in a custom directory:**

```bash
go run . --metrics --path /path/to/logs
```
