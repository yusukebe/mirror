# mirror

`mirror` is command line tool for mirroring a web page.

## Caution

**Do not abuse.**

## Feature

- Using Headless Chrome.
- Fetch HTML although rendered by JavaScript.
- Download all assets that emit when page is loaded.
- Decode Gzip and Brotli encoding content.

## Installation

```plain
go install github.com/yusukebe/mirror/cmd/mirror@latest
```

## Usage

```plain
mirror is a command line tool for mirroring web page

Usage:
  mirror [url] [flags]

Flags:
  -A, --agent string        User-Agent name (default "mirror/v0.0.1")
  -h, --help                help for mirror
  -o, --output-dir string   Output Directory (default "output")
```

## Author

Yusuke Wada <https://github.com/yusukebe>

## License

MIT
