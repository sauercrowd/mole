package mole

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
    "strings"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

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

type ImageManifest struct {
	Config struct{ Digest string }
	Layers []struct{ Digest string }
}

func getManifest(token string, image string, tag string) (*ImageManifest, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://registry-1.docker.io/v2/%s/manifests/%s", image, tag), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Accept", "application/vnd.docker.distribution.manifest.v2+json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	var jsonBody ImageManifest
	err = json.Unmarshal(body, &jsonBody)

	return &jsonBody, err
}

type ImageConfig struct {
	Hostname     string
	Domainname   string
	User         string
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	ExposedPorts map[string]interface{}
	Tty          bool
	OpenStdin    bool
	StdinOnce    bool
	Env          []string
	Cmd          []string
	Image        string
	Volumes      interface{}
	WorkingDir   string
	Entrypoint   []string
	OnBuild      interface{}
	Labels       map[string]interface{}
	StopSignal   string
}

func getConfig(token string, image string, digest string) (*ImageConfig, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://registry-1.docker.io/v2/%s/blobs/%s", image, digest), nil)

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

	if err != nil {
		return nil, err
	}

	var config struct{ Config ImageConfig }
	if err := json.Unmarshal(body, &config); err != nil {
		return nil, err
	}

	return &config.Config, nil

}

func downloadLayer(token string, image string, layer string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/root"
	}

	path := fmt.Sprintf("%s/.mole/layers/%s", homeDir, layer)
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

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	if err := ioutil.WriteFile(fmt.Sprintf("%s/.mole/layers/%s.tar.gz", homeDir, layer), body, 0644); err != nil {
		return "", err
	}

	r := bytes.NewReader(body)

	reader, err := gzip.NewReader(r)
	defer reader.Close()

	if err != nil {
		return "", err
	}

	if err := untar(reader, path); err != nil {
		return "", err
	}

	return path, nil
}

type ImageInfo struct {
	Config   *ImageConfig
	Manifest *ImageManifest
}

func GetConfig(image, tag string) (*ImageInfo, error) {
	token, err := getToken(image)

	if err != nil {
		return nil, err
	}

	manifest, err := getManifest(token, image, tag)

	if err != nil {
		return nil, err
	}

	config, err := getConfig(token, image, manifest.Config.Digest)

	if err != nil {
		return nil, err
	}

	return &ImageInfo{Config: config, Manifest: manifest}, nil
}

func GetImage(dst string, image string, manifest *ImageManifest) error {
	if _, err := os.Stat(dst); os.IsNotExist(err) {
		token, err := getToken(image)
		if err != nil {
			return err
		}

		for _, layer := range manifest.Layers {
			path, err := downloadLayer(token, image, layer.Digest)

			if err != nil {
				return err
			}


			if err := processLayer(path, dst); err != nil {
				return err
			}

		}
	}

	return nil
}

func GetConfigFromDir(dst string) (*ImageConfig, error) {

    target := GetConfigPath(dst)

    body, err := ioutil.ReadFile(target)
    if err != nil {
        return nil, err
    }

    var config ImageConfig
    if err := json.Unmarshal(body, &config); err != nil {
        return nil, err
    }

    return &config, nil
}

func GetConfigPath(dst string) string {
    filteredDst := strings.TrimSuffix(dst, "/")

    splitted := strings.Split(filteredDst, "/")

    splitted[len(splitted)-1] = fmt.Sprintf(".%s.config", splitted[len(splitted)-1] )

    return strings.Join(splitted, "/")
}

func StoreConfigForDir(dst string, config *ImageConfig) error {
    target := GetConfigPath(dst)

    body, err := json.Marshal(config)
    if err != nil {
        return err
    }

    if err := ioutil.WriteFile(target, body, 0644); err != nil {
        return err
    }

    return nil
}
