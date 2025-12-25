package nats

import (
	"encoding/json"

	nats "github.com/nats-io/nats.go"
)

type IngestEvent struct {
	AudioID  int64  `json:"audio_id"`
	S3Path   string `json:"s3_path"`
	Filename string `json:"filename,omitempty"`
}

func PublishIngestEvent(js nats.JetStreamContext, subject string, ev IngestEvent) error {
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = js.Publish(subject, data)
	return err
}
