package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
)

// ZipFiles compresses one or many files into a single zip archive file.
// Param 1: filename is the output zip file's name.
// Param 2: files is a list of files to add to the zip.
func ZipFiles(filename string, folder string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	currDir, err := os.Getwd()

	// log.Println(currDir, folder)

	// Add files to zip
	if info, _ := os.Stat(folder); info.IsDir() {
		localFiles := []string{}
		err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
			// if a file
			if fileinfo, err := os.Stat(path); fileinfo.Mode().IsRegular() && err == nil {
				localFiles = append(localFiles, path)
			}
			return nil
		})

		// Important
		// or it will become highly nested
		os.Chdir(filepath.Join("../", folder))
		for _, loc := range localFiles {
			if err = addFileToZip(zipWriter, filepath.Join(folder, filepath.Base(loc))); err != nil {
				return err
			}
		}
		os.Chdir(currDir)
		return nil
	}
	return errors.New("not a directory")
}

// addFileToZip Adds a file to the zip
func addFileToZip(zipWriter *zip.Writer, filename string) error {

	// log.Println(filename)

	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = filename

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}
