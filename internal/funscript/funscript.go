package funscript

import (
	"context"
	"errors"
	"fmt"
	"github.com/h2non/bimg"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

var NotFoundErr = errors.New("not found")

func fetchFile(ctx context.Context, client *http.Client, fileUrl string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, NotFoundErr
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
func GetHeatmapCover(ctx context.Context, coverUrl string, heatmapUrl string) ([]byte, error) {
	client := http.Client{}

	cover, err := fetchFile(ctx, &client, coverUrl)
	if err != nil {
		return nil, fmt.Errorf("fetch cover: %w", err)
	}
	log.Ctx(ctx).Debug().Int("cover size", len(cover)).Send()
	heatmap, err := fetchFile(ctx, &client, heatmapUrl)
	if err != nil {
		return nil, fmt.Errorf("fetch heatmap: %w", err)
	}
	log.Ctx(ctx).Debug().Int("heatmap size", len(heatmap)).Send()
	heatmapCover, err := overlay(ctx, cover, heatmap)
	if err != nil {
		return nil, fmt.Errorf("overlay heatmap on cover: %w", err)
	}
	log.Ctx(ctx).Trace().Msg("heatmap overlayed on cover")
	return heatmapCover, nil
}

func overlay(ctx context.Context, cover []byte, heatmap []byte) ([]byte, error) {
	log.Ctx(ctx).Debug().Interface("cover type", bimg.DetermineImageType(cover)).Send()
	log.Ctx(ctx).Debug().Interface("heatmap type", bimg.DetermineImageType(heatmap)).Send()

	coverImage := bimg.NewImage(cover)
	coverSize, err := coverImage.Size()
	if err != nil {
		return nil, err
	}
	heatmapImage := bimg.NewImage(heatmap)
	heatmapSize, err := heatmapImage.Size()
	if err != nil {
		return nil, err
	}
	resized, err := heatmapImage.ForceResize(coverSize.Width, heatmapSize.Height)
	if err != nil {
		return nil, err
	}
	overlayed, err := coverImage.WatermarkImage(bimg.WatermarkImage{
		Left:    0,
		Top:     coverSize.Height - heatmapSize.Height,
		Buf:     resized,
		Opacity: 0,
	})
	if err != nil {
		return nil, err
	}
	return overlayed, nil
}
