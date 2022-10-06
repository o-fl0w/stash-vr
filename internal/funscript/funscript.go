package funscript

import (
	"github.com/h2non/bimg"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
)

func fetchFile(client *http.Client, fileUrl string) ([]byte, error) {
	resp, err := client.Get(fileUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	log.Info().Str("url", fileUrl).Int("size", len(buf)).Send()
	return buf, nil
}
func GetOverlay(coverUrl string, heatmapUrl string) ([]byte, error) {
	//http: //10.11.0.10:9667/scene/1737/interactive_heatmap

	client := http.Client{}

	cover, err := fetchFile(&client, coverUrl)
	if err != nil {
		return nil, err
	}
	heatmap, err := fetchFile(&client, heatmapUrl)
	if err != nil {
		return nil, err
	}
	heatmapCover, err := GetHeatmapCover(cover, heatmap)
	if err != nil {
		return nil, err
	}
	return heatmapCover, nil
}

func GetHeatmapCover(cover []byte, heatmap []byte) ([]byte, error) {
	log.Info().Interface("type", bimg.DetermineImageType(cover)).Send()
	log.Info().Interface("type", bimg.DetermineImageType(heatmap)).Send()
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
