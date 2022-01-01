package thumbnail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph"
	"github.com/graphql-go/graphql"
)

type ID struct {
	Source int64
	Url    string
}

func mangaId(manga *backup.Manga) ID {
	return ID{
		*manga.Source,
		*manga.Url,
	}
}

type Config struct {
	Path     string
	Prefix   string
	Download func(*backup.Manga) ([]byte, string, error)
	GetReq   func(*backup.Manga) (*http.Request, error)
	Filename func(manga *backup.Manga) string
}

type Thumbnail struct {
	config      Config
	files       map[ID]string
	rollupFiles map[ID]string
}

func New(config Config) *Thumbnail {
	if config.Download == nil {
		config.Download = func(manga *backup.Manga) ([]byte, string, error) {
			req, err := config.GetReq(manga)
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

	if config.GetReq == nil {
		config.GetReq = func(manga *backup.Manga) (*http.Request, error) {
			return http.NewRequest("GET", manga.GetThumbnailUrl(), nil)
		}
	}

	if config.Filename == nil {
		config.Filename = func(manga *backup.Manga) string {
			sha := sha256.Sum256([]byte(strconv.FormatInt(*manga.Source, 10) + ":" + *manga.Url))
			return base64.RawURLEncoding.EncodeToString(sha[:])
		}
	}

	return &Thumbnail{
		config: config,
		files:  make(map[ID]string),
	}
}

func (t *Thumbnail) DownloadThumbnail(manga *backup.Manga) (string, error) {
	if manga.ThumbnailUrl == nil {
		return "", nil
	}

	filename := t.config.Filename(manga)

	dir, err := os.ReadDir(t.config.Path)
	if err != nil {
		return "", err
	}
	for _, file := range dir {
		if strings.TrimSuffix(file.Name(), path.Ext(file.Name())) == filename {
			return file.Name(), nil
		}
	}

	thumbnail, mimetype, err := t.config.Download(manga)
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

	err = ioutil.WriteFile(path.Join(t.config.Path, filename+ext), thumbnail, 0644)
	if err != nil {
		return "", err
	}

	return filename + ext, nil
}

func (t *Thumbnail) DownloadThumbnails(mangas []*backup.Manga, returnError bool) (map[ID]string, error) {
	files := map[ID]string{}
	for _, manga := range mangas {
		filename, err := t.DownloadThumbnail(manga)
		if err != nil {
			if returnError {
				return nil, err
			}
			log.Print(err)
		} else {
			files[mangaId(manga)] = filename
		}
	}
	return files, nil
}

func (t *Thumbnail) Schema(g *graph.Graph) error {
	g.Types["Manga"].Fields()["thumbnail"] = &graphql.FieldDefinition{
		Type: graphql.String,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			if url, ok := t.files[mangaId(p.Source.(*backup.Manga))]; ok {
				return t.config.Prefix + url, nil
			}
			return nil, nil
		},
	}

	return nil
}

func (t *Thumbnail) Root(_ *graph.Graph, b interface{}) error {
	files, err := t.DownloadThumbnails(b.(*backup.Backup).Mangas, true)
	if err != nil {
		return err
	}
	t.rollupFiles = files
	return nil
}

func (t *Thumbnail) Clean() {
	if t.rollupFiles != nil {
		t.files = t.rollupFiles
	}
	t.rollupFiles = nil
}

func (t *Thumbnail) Worker(ctx context.Context, g *graph.Graph) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(24 * time.Hour):
			t.files, _ = t.DownloadThumbnails(g.Root.(*backup.Backup).Mangas, false)
		}
	}
}
