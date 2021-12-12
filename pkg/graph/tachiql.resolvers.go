package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"github.com/clementd64/tachiql/pkg/backup"
	"github.com/clementd64/tachiql/pkg/graph/generated"
)

func (r *chapterResolver) ChapterNumber(ctx context.Context, obj *backup.Chapter) (*float64, error) {
	return Float(obj.ChapterNumber), nil
}

func (r *mangaResolver) Thumbnail(ctx context.Context, obj *backup.Manga) (*string, error) {
	return r.Indexer.Thumbnail[*obj.ThumbnailUrl], nil
}

func (r *mangaResolver) TotalChapters(ctx context.Context, obj *backup.Manga) (int32, error) {
	return int32(len(obj.Chapters)), nil
}

func (r *mangaResolver) ReadChapters(ctx context.Context, obj *backup.Manga) (int32, error) {
	var read int32 = 0
	for _, chapter := range obj.Chapters {
		if chapter.GetRead() {
			read++
		}
	}
	return read, nil
}

func (r *queryResolver) Mangas(ctx context.Context, status *int32, source *int64) ([]*backup.Manga, error) {
	if status == nil && source == nil {
		return r.Indexer.Backup.Manga, nil
	}

	mangas := []*backup.Manga{}
	for _, manga := range r.Indexer.Backup.Manga {
		if (status == nil || *manga.Status == *status) && (source == nil || *manga.Source == *source) {
			mangas = append(mangas, manga)
		}
	}
	return mangas, nil
}

func (r *queryResolver) Manga(ctx context.Context, title *string, source *int64, url *string) (*backup.Manga, error) {
	if title != nil && source != nil && url != nil {
		return nil, errors.New("title and source + url are mutually exclusive")
	}

	if source != nil && url != nil {
		return r.Indexer.GetMangaById(*source, *url), nil
	}

	if title != nil {
		return r.Indexer.GetMangaByTitle(*title), nil
	}

	return nil, errors.New("invalid query, title or source and url is required")
}

func (r *queryResolver) Categories(ctx context.Context) ([]*backup.Category, error) {
	return r.Indexer.Backup.Categories, nil
}

func (r *queryResolver) Sources(ctx context.Context) ([]*backup.Source, error) {
	return r.Indexer.Backup.Sources, nil
}

func (r *trackingResolver) LastChapterRead(ctx context.Context, obj *backup.Tracking) (*float64, error) {
	return Float(obj.LastChapterRead), nil
}

func (r *trackingResolver) Score(ctx context.Context, obj *backup.Tracking) (*float64, error) {
	return Float(obj.Score), nil
}

// Chapter returns generated.ChapterResolver implementation.
func (r *Resolver) Chapter() generated.ChapterResolver { return &chapterResolver{r} }

// Manga returns generated.MangaResolver implementation.
func (r *Resolver) Manga() generated.MangaResolver { return &mangaResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Tracking returns generated.TrackingResolver implementation.
func (r *Resolver) Tracking() generated.TrackingResolver { return &trackingResolver{r} }

type chapterResolver struct{ *Resolver }
type mangaResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type trackingResolver struct{ *Resolver }
