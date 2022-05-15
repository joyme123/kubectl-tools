package tools

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joyme123/kubectl-tools/version"
	"k8s.io/utils/path"
)

func init() {
	version := version.Get()
	downloadURL := fmt.Sprintf("https://github.com/joyme123/kubectl-tools/releases/download/%s/tools.tar.gz", version.Version)

	tools := []Tool{
		{
			Name:        "arping",
			DownloadURL: downloadURL,
		},
		{
			Name:        "ping",
			DownloadURL: downloadURL,
		},
		{
			Name:        "ping6",
			DownloadURL: downloadURL,
		},
		{
			Name:        "tracepath",
			DownloadURL: downloadURL,
		},
		{
			Name:        "tracepath6",
			DownloadURL: downloadURL,
		},
		{
			Name:        "traceroute6",
			DownloadURL: downloadURL,
		},
	}

	AddToToolSet(tools)
}

var (
	Set map[string]Tool
)

func AddToToolSet(ts []Tool) {
	if Set == nil {
		Set = make(map[string]Tool)
	}
	for i := range ts {
		Set[ts[i].Name] = ts[i]
	}
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

	items := strings.Split(t.DownloadURL, "/")
	fileName := items[len(items)-1]

	// download file
	resp, err := http.DefaultClient.Get(t.DownloadURL)
	if err != nil {
		return "", err
	}

	downloadFile, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	if resp.StatusCode/100 > 2 {
		return "", fmt.Errorf("download %s error with status code: %d", t.DownloadURL, resp.StatusCode)
	}
	_, err = io.Copy(downloadFile, resp.Body)
	if err != nil {
		return "", err
	}

	if ShouldUnarchieve(fileName) {
		if err := Unarchieve(fileName, dir); err != nil {
			return "", fmt.Errorf("unarchieve failed: %v", err)
		}
	}

	if exist, err := path.Exists(path.CheckFollowSymlink, fpath); err != nil {
		return "", err
	} else if !exist {
		return "", fmt.Errorf("%s can't be found", t.Name)
	}

	return fpath, nil
}
