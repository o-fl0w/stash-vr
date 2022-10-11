package heatmap

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"golang.org/x/image/draw"
	"golang.org/x/sync/errgroup"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
)

var errImageNotFound = errors.New("image not found")

func fetchImage(ctx context.Context, fileUrl string) (image.Image, error) {
	log.Ctx(ctx).Trace().Str("url", fileUrl).Msg("Fetching image")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, errImageNotFound
	}

	img, format, err := image.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Ctx(ctx).Trace().Str("format", format).Msg("Fetched image")
	return img, nil
}

func buildHeatmapCover(ctx context.Context, coverUrl string, heatmapUrl string) (image.Image, error) {
	chCover := make(chan draw.Image, 1)
	chHeatmap := make(chan image.Image, 1)

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		cover, err := fetchImage(gCtx, coverUrl)
		if err != nil {
			return fmt.Errorf("fetch cover: %w", err)
		}
		chCover <- cover.(draw.Image)
		close(chCover)
		return nil
	})

	g.Go(func() error {
		heatmap, err := fetchImage(gCtx, heatmapUrl)
		if err != nil {
			return fmt.Errorf("fetch heatmap: %w", err)
		}
		chHeatmap <- heatmap
		close(chHeatmap)
		return nil
	})

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	cover := <-chCover
	heatmap := <-chHeatmap

	heatmapCover := overlay(cover, heatmap)
	return heatmapCover, nil
}

func overlay(dest draw.Image, heatmap image.Image) image.Image {
	destSize := dest.Bounds().Size()
	draw.NearestNeighbor.Scale(dest, image.Rect(0, destSize.Y, destSize.X, destSize.Y-heatmap.Bounds().Size().Y), heatmap, heatmap.Bounds(), draw.Src, nil)
	return dest
}
