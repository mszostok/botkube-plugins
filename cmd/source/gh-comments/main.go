package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gregjones/httpcache"
	"github.com/hashicorp/go-plugin"

	"github.com/kubeshop/botkube/pkg/api"
	"github.com/kubeshop/botkube/pkg/api/source"
	"github.com/kubeshop/botkube/pkg/pluginx"

	"go.szostok.io/botkube-plugins/internal/ptr"
)

const (
	pluginName  = "gh-comments-trigger"
	description = "Watches for new GitHub comments, developed to showcase running executors from GitHub comments."
)

var (
	// version is set via ldflags by GoReleaser.
	version = "dev"

	//go:embed config_schema.json
	configJSONSchema string

	defaultConfig = Config{
		OnRepository: WatchRepository{
			RecheckInterval:       5 * time.Second,
			CommentRequiredPrefix: ptr.FromType("#run"),
		},
	}
)

type (
	// Config holds source configuration.
	Config struct {
		GitHub struct {
			Auth struct {
				// The GitHub access token.
				// Instruction for creating a token can be found here: https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token.
				AccessToken string `yaml:"accessToken"`
			} `yaml:"auth"`
		} `yaml:"github"`
		// For simplification, we only support one repository. It could be changed to a slice of repos.
		OnRepository WatchRepository `yaml:"onRepository"`
	}

	// WatchRepository holds configuration for a repository that should be watched.
	WatchRepository struct {
		RecheckInterval       time.Duration `yaml:"recheckInterval"`
		CommentRequiredPrefix *string       `yaml:"commentRequiredPrefix,omitempty"`
		Name                  string        `yaml:"name"`
	}
)

func (c Config) Validate() error {
	var issues error
	if c.OnRepository.Name == "" {
		issues = errors.Join(issues, fmt.Errorf("Repository name is required"))
	}

	split := strings.Split(c.OnRepository.Name, "/")
	if len(split) != 2 {
		issues = errors.Join(issues, fmt.Errorf(`Wrong repository name. Expected pattern "owner/repository", got %q`, c.OnRepository.Name))
	}

	return issues
}

// GHComments implements Botkube source plugin.
type GHComments struct {
	source.HandleExternalRequestUnimplemented
}

// Metadata returns details about GitHub comments watcher plugin.
func (GHComments) Metadata(_ context.Context) (api.MetadataOutput, error) {
	return api.MetadataOutput{
		Version:     version,
		Description: description,
		JSONSchema: api.JSONSchema{
			Value: configJSONSchema,
		},
	}, nil
}

// Stream watches for new GitHub comments and sends an event when a comment with a given prefix is posted.
//
//nolint:gocritic // hugeParam: in is heavy (120 bytes); consider passing it by pointer
func (GHComments) Stream(ctx context.Context, in source.StreamInput) (source.StreamOutput, error) {
	var cfg Config
	err := pluginx.MergeSourceConfigsWithDefaults(defaultConfig, in.Configs, &cfg)
	if err != nil {
		return source.StreamOutput{}, fmt.Errorf("while merging input configuration: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return source.StreamOutput{}, fmt.Errorf("while validating configuration: %w", err)
	}

	out := source.StreamOutput{
		Event: make(chan source.Event),
	}

	go listenEvents(ctx, cfg, out.Event)

	return out, nil
}

// Payload is the payload of the emitted event.
type Payload struct {
	Command string `json:"command"`
}

// IssueListCommentsOptions specifies the optional parameters for querying GitHub comments API.
type IssueListCommentsOptions struct {
	// Since filters comments by time.
	Since *time.Time `url:"since,omitempty"`
}

func listenEvents(ctx context.Context, cfg Config, sink chan source.Event) {
	timer := time.NewTimer(time.Second) // start immediately
	defer timer.Stop()

	lastCheck := time.Now()
	prefix := ptr.ToValue(cfg.OnRepository.CommentRequiredPrefix)

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: httpcache.NewMemoryCacheTransport(),
	}

	for {
		select {
		case <-ctx.Done():
		case <-timer.C:

			out, err := ListComments(httpClient, cfg, lastCheck)
			if err != nil {
				log.Print(err)
			}

			for _, item := range out {
				comment := ptr.ToValue(item.Body)
				comment = strings.TrimSpace(comment)
				if comment == "" || !strings.HasPrefix(comment, prefix) {
					continue
				}
				comment = strings.TrimSpace(strings.TrimPrefix(comment, prefix))

				sink <- source.Event{
					RawObject: Payload{
						Command: comment,
					},
				}
			}

			lastCheck = time.Now()
			timer.Reset(cfg.OnRepository.RecheckInterval)
		}
	}
}

// IssueComment represents a comment left on an issue.
type IssueComment struct {
	Body *string `json:"body,omitempty"`
}

// ListComments lists all GitHub comments.
func ListComments(cli *http.Client, cfg Config, since time.Time) ([]IssueComment, error) {
	u := fmt.Sprintf("https://api.github.com/repos/%s/issues/comments", cfg.OnRepository.Name)

	q := url.Values{}
	q.Set("since", since.Format(time.RFC3339))

	u = fmt.Sprintf("%s?%s", u, q.Encode())
	req, err := http.NewRequest(http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("if-modified-since", since.Format(http.TimeFormat))
	if cfg.GitHub.Auth.AccessToken != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", cfg.GitHub.Auth.AccessToken))
	}

	var out []IssueComment
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &out)
	if err != nil && err != io.EOF {
		log.Println("Raw body: ", string(raw))
		return nil, err
	}
	return out, nil
}

func main() {
	source.Serve(map[string]plugin.Plugin{
		pluginName: &source.Plugin{
			Source: &GHComments{},
		},
	})
}
