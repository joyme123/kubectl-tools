package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"k8s.io/utils/path"
)

func init() {
	ping := Tool{
		Name:        "ping",
		DownloadURL: "https://github.com/joyme123/kubectl-tools/releases/download/v0.1.0/tools.tar.gz",
	}

	AddToToolSet(ping)
}

var (
	Set map[string]Tool
)

func AddToToolSet(t Tool) {
	if Set == nil {
		Set = make(map[string]Tool)
	}
	Set[t.Name] = t
}

type Tool struct {
	Name        string `json:"name" yaml:"name"`
	DownloadURL string `json:"downloadURL" yaml:"downloadURL"`
}

// GetLocalPath ...
func GetLocalPath(t Tool) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	dir := home + ".kubetools/"
	fpath := dir + t.Name
	if exist, err := path.Exists(path.CheckFollowSymlink, fpath); err != nil {
		return "", err
	} else if exist {
		return fpath, nil
	}

	if err := os.MkdirAll(dir, os.ModeDir); err != nil {
		return "", err
	}

	// download file
	resp, err := http.DefaultClient.Get(t.DownloadURL)
	if err != nil {
		return "", err
	}

	file, err := os.Create(fpath)
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 > 2 {
		return "", fmt.Errorf("download %s error with status code: %d", t.DownloadURL, resp.StatusCode)
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return "", err
	}
	return fpath, nil
}
