package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"iter"

	"github.com/bodgit/sevenzip"
	"github.com/ulikunitz/xz"
)

var ErrUnknownFormat = fmt.Errorf("unknown format")
var ErrEmptyFile = fmt.Errorf("empty file")

type ReadSeekReadAt interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

type File struct {
	Name string
}

type Format struct {
	Name   string
	Magic  [][]byte
	Offset int
	Decode func(r ReadSeekReadAt, size int64) iter.Seq2[File, error]
	IsFast bool
}

func decodeTar(r io.Reader) iter.Seq2[File, error] {
	tr := tar.NewReader(r)

	return func(yield func(File, error) bool) {
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				return
			}
			if err != nil {
				yield(File{}, err)
				return
			}

			if !yield(File{Name: hdr.Name}, nil) {
				return
			}
		}
	}
}

var formats = []Format{
	{
		Name: "tar",
		Magic: [][]byte{
			{'u', 's', 't', 'a', 'r', 0x00},
			{'u', 's', 't', 'a', 'r', 040, 040, 0x00},
		},
		Offset: 257,
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return decodeTar(r)
		},
	},
	{
		Name: "zip",
		Magic: [][]byte{
			{'P', 'K', 0x03, 0x04},
			{'P', 'K', 0x05, 0x06},
			{'P', 'K', 0x07, 0x08},
		},
		IsFast: true,
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return func(yield func(File, error) bool) {
				zr, err := zip.NewReader(r.(io.ReaderAt), size)
				if err != nil {
					yield(File{}, fmt.Errorf("create zip reader: %w", err))
					return
				}

				for _, f := range zr.File {
					if !yield(File{Name: f.Name}, nil) {
						return
					}
				}
			}
		},
	},
	{
		Name:   "rar",
		Magic:  [][]byte{{'R', 'a', 'r', '!'}},
		IsFast: true,
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return func(yield func(File, error) bool) {
				yield(File{}, fmt.Errorf("rar format not yet supported"))
			}
		},
	},
	{
		Name:  "bzip2",
		Magic: [][]byte{{'B', 'Z', 'h'}},
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return decodeTar(bzip2.NewReader(r))
		},
	},
	{
		Name:  "gzip",
		Magic: [][]byte{{0x1F, 0x8B}},
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return func(yield func(File, error) bool) {
				gr, err := gzip.NewReader(r)
				if err != nil {
					yield(File{}, fmt.Errorf("create gzip reader: %w", err))
					return
				}

				for f, err := range decodeTar(gr) {
					if !yield(f, err) {
						return
					}
				}
			}
		},
	},
	{
		Name:  "7z",
		Magic: [][]byte{{'7', 'z', 0xBC, 0xAF, 0x27, 0x1C}},
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return func(yield func(File, error) bool) {
				szr, err := sevenzip.NewReader(r, size)
				if err != nil {
					yield(File{}, fmt.Errorf("create 7z reader: %w", err))
					return
				}

				for _, f := range szr.File {
					if !yield(File{Name: f.Name}, nil) {
						return
					}
				}
			}
		},
	},
	{
		Name:  "xz",
		Magic: [][]byte{{0xFD, '7', 'z', 'X', 'Z', 0x00}},
		Decode: func(r ReadSeekReadAt, size int64) iter.Seq2[File, error] {
			return func(yield func(File, error) bool) {
				gr, err := xz.NewReader(r)
				if err != nil {
					yield(File{}, fmt.Errorf("create xz reader: %w", err))
					return
				}

				for f, err := range decodeTar(gr) {
					if !yield(f, err) {
						return
					}
				}
			}
		},
	},
}

func detectFormat(f io.Reader) (*Format, error) {
	buf := make([]byte, 512)

	n, err := f.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	if n == 0 {
		return nil, ErrEmptyFile
	}

	buf = buf[:n]

	for _, format := range formats {
		for _, magic := range format.Magic {
			if len(buf) < len(magic) || len(buf) < format.Offset+len(magic) {
				continue
			}

			if bytes.Equal(buf[format.Offset:format.Offset+len(magic)], magic) {
				return &format, nil
			}
		}
	}

	return nil, ErrUnknownFormat
}
