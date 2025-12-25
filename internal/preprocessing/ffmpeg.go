package preprocessing

import (
	"context"
	"fmt"
	"os/exec"
)

func NormalizeAndResample(
	ctx context.Context,
	inputPath string,
	outputPath string,
	sampleRate int,
) error {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", inputPath,
		"-ac", "1",
		"-ar", fmt.Sprint(sampleRate),
		"-af", "loudnorm",
		outputPath,
	)
	return cmd.Run()
}
