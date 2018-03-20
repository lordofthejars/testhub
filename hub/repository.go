package hub

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func CreateBuildLayout(home string, project string, build string) (string, error) {
	fullPath := filepath.Join(home, project, build)
	err := os.MkdirAll(fullPath, 0755)

	Debug("Directory %s created to store test results", fullPath)

	return fullPath, err
}

func UncompressContent(destination string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	defer gzr.Close()
	if err != nil {
		return err
	}

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			Debug("Test Results uncompressed at %s", destination)
			return nil

		case err != nil:
			return err

		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(destination, header.Name)

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
}
