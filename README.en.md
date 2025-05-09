# tools

[中文](README.md) | English

## modup

Upgrade or downgrade dependencies to the latest version compatible with the target Go version.

- `--go`: Specify the target Go version for compatibility checks.
- `--indirect`: Also upgrade indirect dependencies.
- `--proxy`: Use the specified proxy instead of reading from environment variables.
- `-w`, `--write`: Write results to (source) file instead of stdout.

### Installation

```shell
go install github.com/saltfishpr/tools/cmd/modup@latest
```

### Usage

```shell
> modup --help
Upgrade or Downgrade dependencies to the latest version compatible with the target Go version.

Usage:
  modup [-w] [--indirect] [--proxy string] [-go string] <path>

Flags:
      --go string      target Go version to use for compatibility checks
  -h, --help           help for modup
      --indirect       upgrade indirect dependencies
      --proxy string   use the specified proxy instead of reading from the environment
  -w, --write          write result to (source) file instead of stdout
```

## sortimports

Sort import statements into groups: standard library, third-party, and local imports.

- `--mode`: Specify file selection mode (`diff`: only changed files, `staged`: only staged files).
- `-m`, `--module`: Manually specify the prefix for local imports.
- `-w`, `--write`: Write results to (source) file instead of stdout.

### Installation

```shell
go install github.com/saltfishpr/tools/cmd/sortimports@latest
```

### Usage

```shell
> sortimports --help
Sort Go imports into standard library, third-party, and local imports groups.

Usage:
  sortimports [-w] [-m string] [--mode string] <project-path>

Flags:
  -h, --help            help for sortimports
      --mode string     specify file selection mode (diff: changed files, staged: staged files)
  -m, --module string   specify the project module path manually
  -w, --write           write result to (source) file instead of stdout
```
