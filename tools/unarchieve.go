package tools

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
)

// ShouldUnarchieve determine file should unarchieve
func ShouldUnarchieve(fileName string) bool {
	items := strings.Split(fileName, ".")
	if len(items) < 2 {
		return false
	}
	n := len(items)
	if items[n-2] == "tar" && items[n-1] == "gz" {
		return true
	} else if items[n-1] == "tar" {
		return true
	}
	return false
}

// Unarchieve unarchieves archieved file. support .tar and .tar.gz only now
func Unarchieve(fileName string, dstPath string) error {
	items := strings.Split(fileName, ".")
	if len(items) < 2 {
		return fmt.Errorf("unsupport file format")
	}
	n := len(items)

	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	if items[n-2] == "tar" && items[n-1] == "gz" {
		return unarchieveTarGz(f, dstPath)
	} else if items[n-1] == "tar" {
		return unarchieveTar(f, dstPath)
	}

	return fmt.Errorf("unsupport file format")
}

func unarchieveTarGz(stream io.Reader, dstPath string) error {
	uncompressedStream, err := gzip.NewReader(stream)
	if err != nil {
		return fmt.Errorf("ExtractTarGz: NewReader failed")
	}

	return unarchieveTar(uncompressedStream, dstPath)
}

func unarchieveTar(stream io.Reader, dstPath string) error {
	tarReader := tar.NewReader(stream)
	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return fmt.Errorf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(dstPath+header.Name, 0755); err != nil {
				return fmt.Errorf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			outFile, err := os.Create(dstPath + header.Name)
			if err != nil {
				return fmt.Errorf("ExtractTarGz: Create() failed: %s", err.Error())
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				return fmt.Errorf("ExtractTarGz: Copy() failed: %s", err.Error())
			}
			outFile.Close()
		default:
			return fmt.Errorf("ExtractTarGz: uknown type: %v in %s", header.Typeflag, header.Name)
		}
	}
	return nil
}
