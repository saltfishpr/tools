# tools

中文 | [English](README.en.md)

## modup

将依赖项升级或降级到与目标 Go 版本兼容的最新版本。

- `--go`：指定用于兼容性检查的目标 Go 版本。
- `--indirect`：同时升级间接依赖项。
- `--proxy`：使用指定的代理，而不是从环境变量中读取。
- `-w`, `--write`：将结果写入（源）文件，而不是输出到标准输出。

### 安装

```shell
go install github.com/saltfishpr/tools/cmd/modup@latest
```

### 使用

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

将导入语句按照标准库、第三方库和本地库进行分组排序。

- `--mode`：指定文件选择模式（`diff`：仅处理已更改的文件，`staged`：仅处理已暂存的文件）。
- `-m`, `--module`：手动指定本地库的前缀。
- `-w`, `--write`：将结果写入（源）文件，而不是输出到标准输出。

### 安装

```shell
go install github.com/saltfishpr/tools/cmd/sortimports@latest
```

### 使用

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
