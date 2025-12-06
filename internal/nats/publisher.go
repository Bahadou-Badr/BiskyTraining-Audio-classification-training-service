package nats

import (
	"encoding/json"
	"time"

	nats "github.com/nats-io/nats.go"
)

type IngestEvent struct {
	AudioID  int64  `json:"audio_id"`
	S3Path   string `json:"s3_path"`
	Filename string `json:"filename,omitempty"`
}

func PublishIngestEvent(js nats.JetStreamContext, subject string, ev IngestEvent) error {
	payload, err := json.Marshal(ev)
	if err != nil {
		return err
	}

	// publish with Ack (JetStream)
	_, err = js.Publish(subject, payload, nats.MsgId(string(time.Now().UTC().AppendFormat([]byte{}, time.RFC3339Nano))))
	if err != nil {
		return err
	}
	return nil
}
