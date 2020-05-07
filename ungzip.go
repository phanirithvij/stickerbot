package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/schollz/progressbar"
)

// UnpackGzipFile https://stackoverflow.com/a/38325264/8608146
func UnpackGzipFile(gzFilePath, dstFilePath string) (int64, error) {
	gzFile, err := os.Open(gzFilePath)
	if err != nil {
		return 0, fmt.Errorf("Failed to open file %s for unpack: %s", gzFilePath, err)
	}
	dstFile, err := os.OpenFile(dstFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return 0, fmt.Errorf("Failed to create destination file %s for unpack: %s", dstFilePath, err)
	}

	ioReader, ioWriter := io.Pipe()

	go func() { // goroutine leak is possible here
		gzReader, _ := gzip.NewReader(gzFile)
		// it is important to close the writer or reading from the other end of the
		// pipe or io.copy() will never finish
		defer func() {
			gzFile.Close()
			gzReader.Close()
			ioWriter.Close()
		}()

		io.Copy(ioWriter, gzReader)
	}()

	written, err := io.Copy(dstFile, ioReader)
	if err != nil {
		return 0, err // goroutine leak is possible here
	}
	ioReader.Close()
	dstFile.Close()

	return written, nil
}

// ExtractStickers ...
func ExtractStickers(src string, dest string) {
	err := os.MkdirAll(dest, os.ModePerm)
	if err != nil {
		panic(err)
	}

	// https://flaviocopes.com/go-list-files/
	var files []string

	err = filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// if a tgs file
		if strings.HasSuffix(path, "tgs") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	count := int64(len(files))
	bar := progressbar.Default(count)
	for _, file := range files {
		bar.Add(1)
		fileBase := filepath.Base(file)
		// fmt.Println("file", file)
		destPath := filepath.Join(dest, fileBase+".json")
		// fmt.Println("dest", destPath)
		_, err := UnpackGzipFile(file, destPath)
		if err != nil {
			panic(err)
		}
		// fmt.Println(d)
	}
}
