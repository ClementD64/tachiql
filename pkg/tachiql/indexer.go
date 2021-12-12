package tachiql

import (
	"log"

	"github.com/clementd64/tachiql/pkg/backup"
)

type mangaId struct {
	source int64
	url    string
}

type Indexer struct {
	Backup *backup.Backup

	iMangaTitle map[string]*backup.Manga
	iMangaId    map[mangaId]*backup.Manga
	Thumbnail   map[string]*string

	thumbnail        *Thumbnail
	thumbnailBaseUrl string
}

func NewIndexer(thumbnail *Thumbnail, thumbnailBaseUrl string) *Indexer {
	return &Indexer{
		Thumbnail:        map[string]*string{},
		thumbnail:        thumbnail,
		thumbnailBaseUrl: thumbnailBaseUrl,
	}
}

func (i *Indexer) IndexBackup(b *backup.Backup) {
	iMangaId := map[mangaId]*backup.Manga{}
	iMangaTitle := map[string]*backup.Manga{}

	for _, m := range b.GetManga() {
		if m.Title != nil {
			iMangaTitle[*m.Title] = m
		}
		if m.Source != nil && m.Url != nil {
			iMangaId[mangaId{source: *m.Source, url: *m.Url}] = m
		}
		if m.ThumbnailUrl != nil {
			thumbnail, err := i.thumbnail.GetThumbnail(m)
			if err != nil {
				log.Print(err)
				continue
			}
			thumbnail = i.thumbnailBaseUrl + thumbnail
			i.Thumbnail[*m.ThumbnailUrl] = &thumbnail
		}
	}

	i.Backup = b
	i.iMangaId = iMangaId
	i.iMangaTitle = iMangaTitle
}

func (i *Indexer) GetMangaById(source int64, url string) *backup.Manga {
	if manga, ok := i.iMangaId[mangaId{source: source, url: url}]; ok {
		return manga
	}
	return nil
}

func (i *Indexer) GetMangaByTitle(title string) *backup.Manga {
	if manga, ok := i.iMangaTitle[title]; ok {
		return manga
	}
	return nil
}
