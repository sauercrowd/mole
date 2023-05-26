package mole

import (
	"io/ioutil"
    "errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

type fileInfo struct {
	path string
	info os.FileInfo
}

func processLayer(srcPath, dstPath string) error {
	files := make([]fileInfo, 0)
	if err := filepath.Walk(srcPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		files = append(files, fileInfo{path, info})
		return nil

	}); err != nil {
		return err
	}

	// deal with whiteouts
	for _, file := range files {

		relPath, err := filepath.Rel(srcPath, file.path)
		if err != nil {
			return err
		}

		baseName := path.Base(relPath)

		if strings.HasSuffix(relPath, ".wh..wh..opq") {
			relPath = strings.TrimSuffix(relPath, ".wh..wh..opq")
			dst := filepath.Join(dstPath, relPath)

			if err := os.RemoveAll(dst); err != nil {
				return err
			}
		} else if strings.HasPrefix(baseName, ".wh.") {
			relPath = path.Join(path.Dir(relPath), strings.TrimPrefix(baseName, ".wh."))
			dst := filepath.Join(dstPath, relPath)

			if err := os.Remove(dst); err != nil {
				return err
			}
		}
	}

	for _, file := range files {

		relPath, err := filepath.Rel(srcPath, file.path)
		if err != nil {
			return err
		}

		dst := filepath.Join(dstPath, relPath)

		baseName := path.Base(relPath)

		if strings.HasSuffix(relPath, ".wh..wh..opq") {
			continue
		} else if strings.HasPrefix(baseName, ".wh.") {
			continue
		} else {

			if file.info.IsDir() {
				if err := os.MkdirAll(dst, file.info.Mode()); err != nil {
					return err
				}

				sysStat, ok := file.info.Sys().(*syscall.Stat_t)
				if !ok {
					return errors.New("Failed to get raw syscall.Stat_t data")
				}

				if err := os.Lchown(dst, int(sysStat.Uid), int(sysStat.Gid)); err != nil {
					return err
				}
			if err := os.Chmod(dst, file.info.Mode().Perm()); err != nil {
				return err
			}

			} else {
				if err := copyFile(dst, file.path, file.info); err != nil {
					return err
				}
				sysStat, ok := file.info.Sys().(*syscall.Stat_t)
				if !ok {
					return errors.New("Failed to get raw syscall.Stat_t data")
				}

				if err := os.Lchown(dst, int(sysStat.Uid), int(sysStat.Gid)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func copyFile(dst, src string, info os.FileInfo) error {
	if _, err := os.Stat(dst); err == nil {
		if err := os.Remove(dst); err != nil {
			return err
		}
	}

	fileInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}

	if fileInfo.Mode()&os.ModeType == os.ModeSymlink {
		link, err := os.Readlink(src)
		if err != nil {
			panic(err)
		}
		if err := os.Symlink(link, dst); err != nil {
			panic(err)
		}
		return nil
	}

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile(dst, input, info.Mode()); err != nil {
		return err
	}

	return os.Chmod(dst, fileInfo.Mode())
}


