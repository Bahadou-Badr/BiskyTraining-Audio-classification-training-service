package preprocessing

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type Pipeline struct {
	Repo   Repository
	S3     S3Uploader
	Python string
}

type S3Uploader interface {
	UploadFromReader(ctx context.Context, object string, r *os.File, size int64, ct string) error
}

func (p *Pipeline) ProcessAudio(
	ctx context.Context,
	audioID int64,
	localInput string,
) error {

	tmpDir := filepath.Join(os.TempDir(), fmt.Sprintf("audio_%d", audioID))
	_ = os.MkdirAll(tmpDir, 0755)

	normalized := filepath.Join(tmpDir, "normalized.wav")

	if err := NormalizeAndResample(ctx, localInput, normalized, 16000); err != nil {
		return err
	}

	// Extract features
	featureOut := filepath.Join(tmpDir, "features.npy")

	cmd := exec.CommandContext(
		ctx,
		p.Python,
		"scripts/extract_features.py",
		"--input", normalized,
		"--output", featureOut,
		"--segment", "0",
	)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Upload features
	f, err := os.Open(featureOut)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	s3Path := fmt.Sprintf("features/audio_%d/seg_0.npy", audioID)

	if err := p.S3.UploadFromReader(ctx, s3Path, f, stat.Size(), "application/octet-stream"); err != nil {
		return err
	}

	return p.Repo.InsertFeature(ctx, audioID, 0, s3Path, "mel")
}
