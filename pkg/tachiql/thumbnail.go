package tachiql

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/clementd64/tachiql/pkg/backup"
)

type Thumbnail struct {
	Path     string
	Download func(manga *backup.Manga) ([]byte, string, error)
	GetReq   func(manga *backup.Manga) (*http.Request, error)
}

func NewThumbnail(t *Thumbnail) *Thumbnail {
	if t.Download == nil {
		t.Download = func(manga *backup.Manga) ([]byte, string, error) {
			req, err := t.GetReq(manga)
			if err != nil {
				return nil, "", err
			}

			client := &http.Client{}
			res, err := client.Do(req)
			if err != nil {
				return nil, "", err
			}
			defer res.Body.Close()

			if res.StatusCode >= 300 {
				return nil, "", errors.New("failed to fetch image " + manga.GetThumbnailUrl())
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, "", err
			}

			return body, res.Header.Get("Content-Type"), nil
		}
	}

	if t.GetReq == nil {
		t.GetReq = func(manga *backup.Manga) (*http.Request, error) {
			return http.NewRequest("GET", manga.GetThumbnailUrl(), nil)
		}
	}

	return t
}

func (t *Thumbnail) GetThumbnail(manga *backup.Manga) (string, error) {
	Url, err := url.Parse(*manga.ThumbnailUrl)
	if err != nil {
		return "", err
	}

	sha := sha256.Sum256([]byte(Url.String()))
	hash := base64.RawURLEncoding.EncodeToString(sha[:])

	dir, err := os.ReadDir(t.Path)
	if err != nil {
		return "", err
	}
	for _, file := range dir {
		if strings.TrimSuffix(file.Name(), path.Ext(file.Name())) == hash {
			return file.Name(), nil
		}
	}

	thumbnail, mimetype, err := t.Download(manga)
	if err != nil {
		return "", err
	}

	exts, err := mime.ExtensionsByType(mimetype)
	if err != nil {
		return "", err
	}

	ext := ".bin"
	if exts != nil {
		ext = exts[len(exts)-1]
	}

	filename := hash + ext
	err = ioutil.WriteFile(path.Join(t.Path, hash+ext), thumbnail, 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}
