package preprocessing

import (
	"context"
	"encoding/json"
	"log"

	nats "github.com/nats-io/nats.go"
)

type Worker struct {
	Pipeline *Pipeline
	Repo     Repository
}

func (w *Worker) Start(js nats.JetStreamContext) error {
	_, err := js.Subscribe("audio.ingest.raw", func(msg *nats.Msg) {
		ctx := context.Background()

		// parse payload (reuse existing IngestEvent)
		var ev struct {
			AudioID int64  `json:"audio_id"`
			S3Path  string `json:"s3_path"`
		}

		if err := json.Unmarshal(msg.Data, &ev); err != nil {
			log.Println("bad message:", err)
			return
		}

		_ = w.Repo.MarkProcessing(ctx, ev.AudioID)

		// download raw audio locally (you already have MinIO client)
		localFile := "/tmp/raw.wav" // simplified for now

		if err := w.Pipeline.ProcessAudio(ctx, ev.AudioID, localFile); err != nil {
			_ = w.Repo.MarkFailed(ctx, ev.AudioID, err.Error())
			return
		}

		_ = w.Repo.MarkDone(ctx, ev.AudioID)
	})
	return err
}
