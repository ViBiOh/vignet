package vignet

import (
	"encoding/json"
	"errors"
	"fmt"
)

type ffprobeStreamGroups struct {
	StreamGroups []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Components []struct {
			Subcomponents []struct {
				StreamIndex          int `json:"stream_index"`
				TileHorizontalOffset int `json:"tile_horizontal_offset"`
				TileVerticalOffset   int `json:"tile_vertical_offset"`
			} `json:"subcomponents"`
			NbTiles          int `json:"nb_tiles"`
			CodedWidth       int `json:"coded_width"`
			CodedHeight      int `json:"coded_height"`
			HorizontalOffset int `json:"horizontal_offset"`
			VerticalOffset   int `json:"vertical_offset"`
			Width            int `json:"width"`
			Height           int `json:"height"`
		} `json:"components"`
		Index     int `json:"index"`
		NbStreams int `json:"nb_streams"`
	} `json:"stream_groups"`
}

func getTileFromStreamGroups(payload []byte) (string, error) {
	var content ffprobeStreamGroups

	if err := json.Unmarshal(payload, &content); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}

	if len(content.StreamGroups) == 0 {
		return "", errors.New("no stream group")
	}

	if len(content.StreamGroups[0].Components) == 0 {
		return "", errors.New("no stream group component")
	}

	if len(content.StreamGroups[0].Components[0].Subcomponents) == 0 {
		return "", errors.New("no stream group sub component")
	}

	var horizontal, previousVerticalOffset int

	vertical := 1

	for _, component := range content.StreamGroups[0].Components[0].Subcomponents {
		if previousVerticalOffset != component.TileVerticalOffset {
			previousVerticalOffset = component.TileVerticalOffset
			vertical += 1
		}

		if vertical == 1 {
			horizontal += 1
		}
	}

	return fmt.Sprintf("%dx%d", horizontal, vertical), nil
}
