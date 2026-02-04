package vignet

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	absto "github.com/ViBiOh/absto/pkg/model"
	"github.com/ViBiOh/httputils/v4/pkg/hash"
	"github.com/ViBiOh/httputils/v4/pkg/telemetry"
	"github.com/ViBiOh/vignet/pkg/model"
)

const thumbnailDuration = 5

func (s Service) storageThumbnail(ctx context.Context, itemType model.ItemType, input, output string, scale uint64) (err error) {
	if err = s.storage.Mkdir(ctx, path.Dir(output), absto.DirectoryPerm); err != nil {
		return fmt.Errorf("create directory for output: %w", err)
	}

	var inputName string
	var finalizeInput func()

	inputName, finalizeInput, err = s.getInputName(ctx, input)
	if err != nil {
		err = fmt.Errorf("get input name: %w", err)
	} else {
		outputName, finalizeOutput := s.getOutputName(ctx, output)
		err = errors.Join(s.getThumbnailGenerator(itemType)(ctx, inputName, outputName, scale), finalizeOutput())
		finalizeInput()
	}

	return err
}

func (s Service) imageThumbnail(ctx context.Context, inputName, outputName string, scale uint64) error {
	var err error

	ctx, end := telemetry.StartSpan(ctx, s.tracer, "ffmpeg_thumbnail")
	defer end(&err)

	var formatOption string

	if path.Ext(inputName) == ".heic" {
		tile, rotation, err := s.getHeicDetails(ctx, inputName)
		if err != nil {
			return fmt.Errorf("get heic tiling: %w", err)
		}

		formatOption += "tile=" + tile + ","

		inputName, err = s.generateHeicMap(ctx, inputName)
		if err != nil {
			return fmt.Errorf("render heic map: %w", err)
		}

		if rotation != 0 {
			if rotation < 0 {
				rotation *= -1
			}

			formatOption += fmt.Sprintf("rotate=%d*PI/180,", rotation)
		}

		defer func() {
			if err := s.rmDir(path.Dir(inputName)); err != nil {
				slog.Error("Unable to clean heic temp files", slog.Any("error", err))
			}
		}()
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-hwaccel", "auto", "-i", inputName, "-map_metadata", "-1", "-vf", fmt.Sprintf("%scrop='min(iw,ih)':'min(iw,ih)',scale=%d:%d", formatOption, scale, scale), "-vcodec", "libwebp", "-lossless", "0", "-compression_level", "6", "-q:v", qualityForScale(scale), "-an", "-preset", "picture", "-y", "-f", "webp", "-frames:v", "1", outputName)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	if err = cmd.Run(); err != nil {
		cleanLocalFile(ctx, outputName)
		return fmt.Errorf("ffmpeg image: %s: %w", buffer.String(), err)
	}

	return nil
}

func (s Service) videoThumbnail(ctx context.Context, inputName, outputName string, scale uint64) error {
	var err error

	ctx, end := telemetry.StartSpan(ctx, s.tracer, "ffmpeg_video_thumbnail")
	defer end(&err)

	ffmpegOpts := []string{"-hwaccel", "auto"}
	var customOpts []string

	if _, duration, err := s.getVideoDetailsFromLocal(ctx, inputName); err != nil {
		slog.LogAttrs(ctx, slog.LevelError, "get container duration", slog.String("input", inputName), slog.Any("error", err))
		ffmpegOpts = append(ffmpegOpts, "-ss", "1.000")
	} else {
		startPoint := duration / 2
		if duration > thumbnailDuration {
			startPoint -= thumbnailDuration / 2
		}

		ffmpegOpts = append(ffmpegOpts, "-ss", fmt.Sprintf("%.3f", startPoint))
	}

	format := fmt.Sprintf("crop='min(iw,ih)':'min(iw,ih)',scale=%d:%d", scale, scale)
	ffmpegOpts = append(ffmpegOpts, "-t", strconv.Itoa(thumbnailDuration))
	customOpts = []string{"-r", "8", "-loop", "0"}

	ffmpegOpts = append(ffmpegOpts, "-i", inputName, "-map_metadata", "-1", "-vf", format, "-vcodec", "libwebp", "-lossless", "0", "-compression_level", "6", "-q:v", qualityForScale(scale), "-an", "-preset", "picture", "-y", "-f", "webp")
	ffmpegOpts = append(ffmpegOpts, customOpts...)
	ffmpegOpts = append(ffmpegOpts, outputName)
	cmd := exec.Command("ffmpeg", ffmpegOpts...)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	if err = cmd.Run(); err != nil {
		cleanLocalFile(ctx, outputName)
		return fmt.Errorf("ffmpeg video: %s: %w", buffer.String(), err)
	}

	return nil
}

func (s Service) getVideoDetailsFromLocal(ctx context.Context, name string) (int64, float64, error) {
	reader, err := os.OpenFile(name, os.O_RDONLY, absto.RegularFilePerm)
	if err != nil {
		return 0, 0, fmt.Errorf("open file: %w", err)
	}
	defer closeWithLog(ctx, reader, "getVideoBitrate", name)

	return s.getVideoDetails(ctx, name)
}

func (s Service) getHeicDetails(ctx context.Context, inputName string) (tile string, rotation int, err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "ffprobe")
	defer end(&err)

	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-print_format", "json", "-show_stream_groups", "-show_entries", "stream_side_data=rotation", inputName)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	if err = cmd.Run(); err != nil {
		return "", 0, fmt.Errorf("ffprobe error `%s`: %s", err, buffer.String())
	}

	return getTileFromStreamGroups(buffer.Bytes())
}

func (s Service) generateHeicMap(ctx context.Context, inputName string) (output string, err error) {
	ctx, end := telemetry.StartSpan(ctx, s.tracer, "ffprobe")
	defer end(&err)

	output = filepath.Join(s.tmpFolder, hash.String(inputName))

	if err := os.MkdirAll(output, 0o700); err != nil {
		return output, fmt.Errorf("create dir `%s`: %w", output, err)
	}

	output = filepath.Join(output, "part_%d.jpeg")

	cmd := exec.CommandContext(ctx, "ffmpeg", "-v", "error", "-hwaccel", "auto", "-i", inputName, "-map", "0", output)

	buffer := bufferPool.Get().(*bytes.Buffer)
	defer bufferPool.Put(buffer)

	buffer.Reset()
	cmd.Stdout = buffer
	cmd.Stderr = buffer

	if err = cmd.Run(); err != nil {
		return output, fmt.Errorf("ffmpeg map: %s: %w", buffer.String(), err)
	}

	return output, nil
}

func qualityForScale(scale uint64) string {
	if scale == SmallSize {
		return "66"
	}
	return "80"
}
