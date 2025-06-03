package heresphere

import "stash-vr/internal/util"

var resolutions = []int{240, 360, 480, 540, 720, 1080, 1440, 2160, 4320}

func nearestResolution(n int) (int, string) {
	nearest := resolutions[0]
	minDiff := util.Abs(n - nearest)

	for _, r := range resolutions[1:] {
		if d := util.Abs(n - r); d < minDiff || (d == minDiff && r < nearest) {
			minDiff = d
			nearest = r
		}
	}

	var tier string
	switch {
	case nearest <= 540:
		tier = "Low"
	case nearest <= 1080:
		tier = "Medium"
	default:
		tier = "High"
	}

	return nearest, tier
}
