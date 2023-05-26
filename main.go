package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	image := "library/nginx"

	targetDir := "./nginx"

	token, err := getToken(image)

	if err != nil {
		log.Fatal(err)
	}

	layers, err := getManifest(token, image, "latest")

	if err != nil {
		log.Fatal(err)
	}

	for i := range layers {
        layer := layers[len(layers)-1-i]
		path, err := downloadLayer(token, image, layer)

		if err != nil {
			log.Fatal(err)
		}

		if err := processLayer(path, targetDir); err != nil {
			log.Fatal(err)
		}

		fmt.Println(layer)
	}
}

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
			} else {
				if err := copyFile(dst, file.path, file.info); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func copyFile(dst, src string, info os.FileInfo) error {
	fileInfo, err := os.Lstat(src)
	if err != nil {
		return err
	}


	if fileInfo.Mode()&os.ModeType == os.ModeSymlink {
		link, err := os.Readlink(src)
		if err != nil {
			return err
		}
		return os.Symlink(link, dst)
	}

	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, input, info.Mode())
}

func getToken(image string) (string, error) {
	resp, err := http.Get(fmt.Sprintf("https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull", image))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var token struct{ Token string }
	err = json.Unmarshal(body, &token)

	return token.Token, err
}

func getManifest(token string, image string, tag string) ([]string, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", image, tag), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

    log.Println("body", string(body))

	var jsonBody struct{ FsLayers []struct{ BlobSum string } }
	err = json.Unmarshal(body, &jsonBody)

	if err != nil {
		return nil, err
	}

	var layers []string

	for _, layer := range jsonBody.FsLayers {
		layers = append(layers, layer.BlobSum)
	}
	return layers, nil

}

func downloadLayer(token string, image string, layer string) (string, error) {
	path := fmt.Sprintf("/home/jonas/.ducker/layers/%s", layer)
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		return path, nil

	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return "", err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", image, layer), nil)

	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	reader, err := gzip.NewReader(resp.Body)
	defer reader.Close()

	if err != nil {
		return "", err
	}

	if err := untar(reader, path); err != nil {
		return "", err
	}

	return path, nil
}

func untar(reader io.Reader, dst string) error {
	tr := tar.NewReader(reader)

	for {
		header, err := tr.Next()
		switch {
		// no more files
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := filepath.Join(dst, header.Name)

		// if _, err := os.Stat(dst); err != nil {
		// 	if err := os.MkdirAll(dst, 0755); err != nil {
		// 		return err
		// 	}
		// }

		// base := path.Dir(target) + "/"

		switch header.Typeflag {

		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, target); err != nil {
				log.Fatal(err)
			}

		case tar.TypeLink:
			if err := os.Link(path.Join(dst, header.Linkname), target); err != nil {
				log.Fatal(err)
			}

		// create directory if doesn't exit
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.Mkdir(target, header.FileInfo().Mode()); err != nil {
					return err
				}
			}
		// create file
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()

			// copy contents to file
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

		default:
			log.Fatal(header.Format)
		}

	}
}
