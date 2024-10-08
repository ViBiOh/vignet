package vignet

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ViBiOh/httputils/v4/pkg/httperror"
	"github.com/ViBiOh/vignet/pkg/model"
)

const defaultScale uint64 = 150

func (s Service) HandlePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	itemType, err := model.ParseItemType(r.URL.Query().Get("type"))
	if err != nil {
		httperror.BadRequest(ctx, w, err)
		s.increaseMetric(r.Context(), "http", "thumbnail", "", "invalid")
		return
	}

	scale := defaultScale
	if rawScale := r.URL.Query().Get("scale"); len(rawScale) > 0 {
		scale, err = strconv.ParseUint(r.URL.Query().Get("scale"), 10, 64)
		if err != nil {
			httperror.BadRequest(ctx, w, fmt.Errorf("parse scale: %w", err))
			s.increaseMetric(r.Context(), "http", "thumbnail", "", "invalid")
			return
		}
	}

	switch itemType {
	case model.TypeImage, model.TypeVideo:
		var inputName string
		inputName, err = s.saveFileLocally(ctx, r.Body, time.Now().String())
		defer cleanLocalFile(ctx, inputName)

		if err == nil {
			outputName := s.getLocalFilename(fmt.Sprintf("output_%s", inputName))
			defer cleanLocalFile(ctx, outputName)

			if err = s.getThumbnailGenerator(itemType)(r.Context(), inputName, outputName, scale); err == nil {
				err = copyLocalFile(ctx, outputName, w)
			}
		}

	default:
		httperror.BadRequest(ctx, w, errors.New("unhandled item type"))
		return
	}

	if err != nil {
		httperror.InternalServerError(ctx, w, err)
		s.increaseMetric(r.Context(), "http", "thumbnail", itemType.String(), "error")
		return
	}

	s.increaseMetric(r.Context(), "http", "thumbnail", itemType.String(), "success")
}
