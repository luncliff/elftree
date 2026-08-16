# ELF tree

`elftree` is a command line tool that shows the shared library dependencies of
an ELF binary as a tree.

The tool only reads files, so it also runs on Windows and macOS when you need to
inspect Linux binaries from a different host.

## Installation

With a Go 1.24 (or later) toolchain:

```console
$ go install github.com/luncliff/elftree@latest
```

The binary is placed in `$(go env GOBIN)`, or `$(go env GOPATH)/bin` when
`GOBIN` is not set. Make sure that directory is in your `PATH`.

To build from a checkout:

```console
$ go build -o elftree .
```

Pre-built binaries for Linux, macOS, and Windows (amd64 and arm64) are attached
to each [release](https://github.com/luncliff/elftree/releases), together with a
`SHA256SUMS` file:

```console
$ curl -LO https://github.com/luncliff/elftree/releases/latest/download/SHA256SUMS
```

## Usage

```console
$ elftree --help
elftree - show library dependencies of an ELF binary

Usage:
  elftree [options] <file>

Options:
  -d int
    	alias for --depth
  -depth int
    	limit the dependency depth (0 means no limit)
  -dynamic
    	show the .dynamic section entries
  -f string
    	alias for --format (default "tree")
  -format string
    	output format: tree, flat or json (default "tree")
  -p	alias for --path
  -path
    	show the resolved path of each library
  -sections
    	show the section headers
  -segments
    	show the program headers
  -symbols
    	show the symbol tables
  -v	alias for --verbose
  -verbose
    	show a summary of the ELF file
  -version
    	show the version and exit
```

Both `-flag` and `--flag` spellings are accepted.

### Examples

Print the dependency tree:

```console
$ elftree /bin/ls
ls
   libselinux.so.1
      libpcre2-8.so.0
      libc.so.6
         ld-linux-x86-64.so.2
      ld-linux-x86-64.so.2
   libc.so.6
```

Show where each library was resolved from, limited to the direct dependencies:

```console
$ elftree --path --depth 2 /bin/ls
ls  => /usr/bin/ls
   libselinux.so.1  => /usr/lib/x86_64-linux-gnu/libselinux.so.1
   libc.so.6  => /usr/lib/x86_64-linux-gnu/libc.so.6
```

List every unique dependency, one per line:

```console
$ elftree --format flat /bin/ls
ls
libselinux.so.1
libpcre2-8.so.0
libc.so.6
ld-linux-x86-64.so.2
```

Machine readable output for scripts:

```console
$ elftree --format json /bin/ls | jq -r '.children[].name'
libselinux.so.1
libc.so.6
```

Inspect the ELF file itself:

```console
$ elftree --verbose --segments --sections --dynamic --symbols /bin/ls
```

Libraries are resolved in the order described in `ld.so(8)`: `DT_RPATH`,
`LD_LIBRARY_PATH`, `DT_RUNPATH`, `/etc/ld.so.conf`, and the default library
directories.

### Exit codes

| Code | Meaning                                             |
| ---- | --------------------------------------------------- |
| 0    | Success                                             |
| 1    | The file could not be read or a dependency is missing |
| 2    | Invalid command line usage                          |

## Development

```console
$ gofmt -l .
$ go vet ./...
$ go test ./...
```

The CLI lives in `main.go`; the ELF parsing, dependency resolution, and
rendering code lives in `internal/elftree`.

## License

MIT. See [LICENSE](LICENSE).
