package mole

import (
	"archive/tar"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
)

func untar(reader io.Reader, dst string) error {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		switch {

		case err == io.EOF:
			return nil

		case err != nil:
			return err

		case header == nil:
			continue

		}

		target := filepath.Join(dst, header.Name)

		switch header.Typeflag {

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				return err
			}

			err = os.Lchown(target, header.Uid, header.Gid)
			if err != nil {
				return err
			}

		case tar.TypeLink:
			if err := os.Link(path.Join(dst, header.Linkname), target); err != nil {
				return err
			}

		// create directory if doesn't exit
		case tar.TypeDir:
			if err := os.Mkdir(target, header.FileInfo().Mode()); err != nil {
				return err
			}

			err = os.Lchown(target, header.Uid, header.Gid)
			if err != nil {
				return err
			}

			if err := os.Chmod(target, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		// create file
		case tar.TypeReg:

			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, header.FileInfo().Mode().Perm())
			if err != nil {
				return err
			}

			// copy contents to file
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			f.Close()

			err = os.Lchown(target, header.Uid, header.Gid)
			if err != nil {
				return err
			}

			if err := os.Chmod(target, header.FileInfo().Mode().Perm()); err != nil {
				return err
			}

		default:
			log.Fatal(header.Format)
		}

	}
}
