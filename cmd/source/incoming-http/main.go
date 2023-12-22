package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/go-plugin"

	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/source"
)

const (
	pluginName  = "incoming-http"
	description = "Emits events sends via HTTP POST requests."
)

var (
	// version is set via ldflags by GoReleaser.
	version = "dev"

	//go:embed config_schema.json
	configJSONSchema string
	//go:embed webhook_schema.json
	incomingWebhookJSONSchema string
)

// IncomingHTTP implements Botkube source plugin.
type IncomingHTTP struct{}

// Metadata returns details about incoming HTTP plugin.
func (IncomingHTTP) Metadata(_ context.Context) (api.MetadataOutput, error) {
	return api.MetadataOutput{
		Version:     version,
		Description: description,
		JSONSchema: api.JSONSchema{
			Value: configJSONSchema,
		},
		ExternalRequest: api.ExternalRequestMetadata{
			Payload: api.ExternalRequestPayload{
				JSONSchema: api.JSONSchema{
					Value: incomingWebhookJSONSchema,
				},
			},
		},
	}, nil
}

// Stream is not implemented as we only watch for external requests from Botkube incoming webhook.
func (IncomingHTTP) Stream(context.Context, source.StreamInput) (source.StreamOutput, error) {
	return source.StreamOutput{}, nil
}

// Payload is the payload of the emitted event.
type Payload struct {
	Command string `json:"command"`
}

// HandleExternalRequest handles incoming Payload and returns an event based on it.
//
//nolint:gocritic // hugeParam: in is heavy (120 bytes); consider passing it by pointer
func (IncomingHTTP) HandleExternalRequest(_ context.Context, in source.ExternalRequestInput) (source.ExternalRequestOutput, error) {
	var p Payload
	err := json.Unmarshal(in.Payload, &p)
	if err != nil {
		return source.ExternalRequestOutput{}, fmt.Errorf("while unmarshaling Payload: %w", err)
	}

	p.Command = strings.TrimSpace(p.Command)
	if p.Command == "" {
		return source.ExternalRequestOutput{}, fmt.Errorf("command cannot be empty")
	}

	return source.ExternalRequestOutput{
		Event: source.Event{
			RawObject: p,
		},
	}, nil
}

func main() {
	source.Serve(map[string]plugin.Plugin{
		pluginName: &source.Plugin{
			Source: &IncomingHTTP{},
		},
	})
}
