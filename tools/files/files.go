package files

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

func CopyFile(sourceFile string, targetFile string) error {
	source, err := os.Open(sourceFile)
	if err != nil {
		return err
	}
	defer source.Close()

	target, err := os.Create(targetFile)
	if err != nil {
		return err
	}
	defer target.Close()

	_, err = io.Copy(target, source)

	return err
}

func MoveFolder(sourceFolder string, destinationFolder string) (err error) {
	err = filepath.Walk(sourceFolder, func(file string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", file, err)
			return err
		}

		relativePath, err := filepath.Rel(sourceFolder, file)
		targetPath := path.Join(destinationFolder, relativePath)

		if !info.IsDir() {
			CopyFile(file, targetPath)
			os.Remove(file)
		} else {
			os.Mkdir(targetPath, 777)
		}

		return nil
	})
	return err
}

func MakeAbsoluteUrl(relativeUrl string, requestUrl *url.URL) (string, error) {
	uri, err := url.Parse(strings.TrimSpace(relativeUrl))
	if err != nil {
		return relativeUrl, err
	}
	return requestUrl.ResolveReference(uri).String(), nil
}

func DownloadFile(downloadUrl string, destinationFolder string) (localFile string, err error) {
	resp, err := http.Get(downloadUrl)
	if err != nil {
		return "", err
	}

	urlPath := resp.Request.URL.Path
	i := strings.LastIndex(urlPath, "/")+1
	filename := urlPath[i:]
	contentDisposition := resp.Header.Get("Content-Disposition")
	if strings.Contains(contentDisposition, "filename=") {
		filename = contentDisposition[strings.LastIndex(contentDisposition, "=")+1:]
		if strings.HasPrefix(filename, "\"") {
			filename = filename[1:]
		}
		if strings.HasSuffix(filename, "\"") {
			filename = filename[:len(filename)-1]
		}
	}
	localFile = path.Join(destinationFolder, filename)

	expectedSize, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	info, err := os.Stat(localFile)
	if err == nil && info.Size() == expectedSize {
		return localFile, nil
	}

	file, err := os.Create(localFile)
	if err == nil {
		defer file.Close()
		io.Copy(file, resp.Body)
		defer resp.Body.Close()
		file.Sync()
	}

	return localFile, err
}