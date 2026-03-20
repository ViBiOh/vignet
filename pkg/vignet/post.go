package vignet

import (
	"errors"
	"fmt"
	"log/slog"
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

	name := r.URL.Query().Get("name")
	if len(name) == 0 {
		name = time.Now().String()
	}

	scale := defaultScale
	if rawScale := r.URL.Query().Get("scale"); len(rawScale) > 0 {
		if scale, err = strconv.ParseUint(r.URL.Query().Get("scale"), 10, 64); err != nil {
			httperror.BadRequest(ctx, w, fmt.Errorf("parse scale: %w", err))
			s.increaseMetric(r.Context(), "http", "thumbnail", "", "invalid")
			return
		}
	}

	switch itemType {
	case model.TypeImage, model.TypeVideo:
		var inputName string
		inputName, err = s.saveFileLocally(ctx, r.Body, name)
		defer cleanLocalFile(ctx, inputName)

		if err == nil {
			outputName := s.getLocalFilename(fmt.Sprintf("output_%s", inputName))
			defer cleanLocalFile(ctx, outputName)

			if err = s.getThumbnailGenerator(itemType)(r.Context(), name, inputName, outputName, scale); err == nil {
				if copyErr := copyLocalFile(ctx, outputName, w); copyErr != nil {
					slog.ErrorContext(ctx, "unable to copy file to HTTP response", slog.Any("error", copyErr))
					s.increaseMetric(r.Context(), "http", "thumbnail", itemType.String(), "error")
					return
				}
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
