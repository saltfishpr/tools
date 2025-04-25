# tools

中文 | [English](README.en.md)

## modup

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

升级依赖包到符合当前项目 go 版本的最新版

`--indirect`: 同时升级间接依赖。默认只升级直接依赖

## sortimports

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

对项目下所有 `.go` 文件的 import 排序

- 标准库
- 三方包
- 项目包

`-m <module-path>`: 手动指定项目包路径
`--staged`: 只处理暂存的 `.go` 文件，默认为 `true`，可以设置 `--staged=false` 处理全部文件
