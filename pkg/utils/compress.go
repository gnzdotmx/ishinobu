package utils

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func CompressOutput(srcDir, outputFilename string) error {
	// Create tar.gz file
	file, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer file.Close()

	gw := gzip.NewWriter(file)
	defer gw.Close()

	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Add files to archive
	files, err := filepath.Glob(filepath.Join(srcDir, "*"))
	if err != nil {
		return err
	}

	for _, fileName := range files {
		err = addFileToTarWriter(fileName, tw)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFileToTarWriter(fileName string, tw *tar.Writer) error {
	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(stat, file.Name())
	if err != nil {
		return err
	}
	header.Name = filepath.Base(fileName)
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	return err
}
