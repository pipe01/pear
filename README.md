# pear

***Pe**ek into **ar**chives*

Supported formats:
* ar
* tar
* tar in bzip2
* tar in gzip
* tar in xz
* zip
* 7z

## Installation

Download the binary corresponding to your environment from [the latest release](https://github.com/pipe01/pear/releases).

Alternatively, install the [Go](https://go.dev/) toolchain, then run

``` bash
go install github.com/pipe01/pear@latest
```

## Usage

```
-n N
    stop after N files, or -1 for all (default 10)
-s N
    skip the first N files
```

## Examples

```bash
$ pear archive.tar
file1
file2
dir/file3
dir/file4

$ pear -n 3 archive.tar
file1
file2
dir/file3

$ pear -s 1 -n 3 archive.tar
file2
dir/file3
dir/file4
```
