package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	maxCount := flag.Int("n", 10, "stop after `N` files, or -1 for all")
	skipCount := flag.Int("s", 0, "skip the first `N` files")
	showVersion := flag.Bool("version", false, "show version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <archive>\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *showVersion {
		fmt.Println("pear version 1.0.0")
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	if err := processFile(flag.Arg(0), *skipCount, *maxCount); err != nil {
		fmt.Fprint(os.Stderr, "error: ")
		fmt.Fprintln(os.Stderr, err)

		switch err {
		case ErrUnknownFormat:
			os.Exit(2)
		case ErrEmptyFile:
			os.Exit(3)
		}
		os.Exit(1)
	}
}

func processFile(path string, skipCount, maxCount int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	format, err := detectFormat(f)
	if err != nil {
		return err
	}

	size, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	f.Seek(0, io.SeekStart)

	count := 0

	for file, err := range format.Decode(f, size) {
		if err != nil {
			return err
		}

		if count < skipCount {
			count++
			continue
		}

		if maxCount < 0 || count < maxCount+skipCount {
			fmt.Println(file.Name)
			count++
		} else {
			break
		}
	}

	return nil
}
