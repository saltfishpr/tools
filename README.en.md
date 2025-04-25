# tools

[中文](README.md) | English

## modup

```shell
go install github.com/saltfishpr/tools/cmd/modup@latest
```

```shell
> modup --help
Upgrade Go module dependencies to the latest compatible version.

Usage:
  modup [-w] [--indirect] [--proxy string] <path>

Flags:
  -h, --help           help for modup
      --indirect       upgrade indirect dependencies
      --proxy string   use the specified proxy instead of reading from the environment
  -w, --write          write result to (source) file instead of stdout
```

Upgrade dependencies to the latest version compatible with the current Go project version.

`--indirect`: Also upgrade indirect dependencies. By default, only direct dependencies are upgraded.

## sortimports

```shell
go install github.com/saltfishpr/tools/cmd/sortimports@latest
```

```shell
> sortimports --help
Sort Go imports into standard library, third-party, and local imports groups.

Usage:
  sortimports [-w] [-m module-path] [--staged] <project-path>

Flags:
  -h, --help            help for sortimports
  -m, --module string   specify the project module path manually
      --staged          only process git staged files (default true)
  -w, --write           write result to (source) file instead of stdout
```

Sort imports in all `.go` files in the project:

- Standard library
- Third-party packages
- Project packages

`-m <module-path>`: Manually specify the project module path.
`--staged`: Only process staged `.go` files. Default is `true`. Set `--staged=false` to process all files.
