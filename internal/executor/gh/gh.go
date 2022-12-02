package gh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-shellwords"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"go.szostok.io/botkube-plugins/internal/gh"
)

// Config holds executor configuration.
type Config struct {
	Token string
	Repo  struct {
		Owner string
		Name  string
	}
}

// Executor implements Botkube executor plugin.
type Executor struct {
	client *kubernetes.Clientset
}

func NewExecutor(client *kubernetes.Clientset) *Executor {
	return &Executor{client: client}
}

// Execute returns a given command as response.
func (e *Executor) Execute(ctx context.Context, cfg Config, command string) (string, error) {
	podName, podNamespace, reason, err := e.resolveCommandArgs(ctx, command)
	if err != nil {
		return "", fmt.Errorf("while resolving command arguments: %w", err)
	}

	ghClient, err := gh.NewClient(cfg.Token)
	if err != nil {
		return "", fmt.Errorf("while creating gh client: %w", err)
	}

	repo := gh.NewRepo(cfg.Repo.Owner, cfg.Repo.Name)

	logs, err := e.getPodLogs(ctx, podNamespace, podName)
	if err != nil {
		return "", fmt.Errorf("while fetching logs : %w", err)
	}

	describe, err := e.getPodDescription(ctx, podNamespace, podName)
	if err != nil {
		return "", fmt.Errorf("while fetching describe : %w", err)
	}

	bodyData, err := e.getBody(podNamespace, podName, reason, logs, describe)
	if err != nil {
		return "", fmt.Errorf("while getting body data: %w", err)
	}
	body, err := RenderIssueBody(bodyData)
	if err != nil {
		return "", fmt.Errorf("while rendering body: %w", err)
	}
	issue, err := ghClient.GetIssues(repo)
	if err != nil {
		return "", fmt.Errorf("while searching for existing issue: %w", err)
	}

	if issue.Number != 0 {
		fmt.Println(issue.Number)
		// TODO: add comment instead of creating a new issue
	}

	in := gh.CreateIssueInput{
		Title: fmt.Sprintf("Issue for %s/%s in %s", bodyData.PodNamespace, bodyData.PodName, bodyData.Cluster.Name),
		Body:  body,
	}
	newIssue, err := ghClient.CreateIssue(repo, in)
	if err != nil {
		return "", fmt.Errorf("while creating issue: %w", err)
	}

	return fmt.Sprintf("New issue created. See: %s", newIssue.URL), nil
}

func (e *Executor) getPodLogs(ctx context.Context, namespace, name string) (string, error) {
	req := e.client.CoreV1().Pods(namespace).GetLogs(name, &v1.PodLogOptions{})
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("while opening stream: %w", err)
	}
	defer podLogs.Close()

	var buf bytes.Buffer
	_, err = io.Copy(&buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("while copy information from podLogs to buf: %w", err)
	}
	return buf.String(), nil
}

func (e *Executor) getBody(namespace, name, reason, logs, describe string) (*Body, error) {
	ver, err := e.client.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("while getting k8s version: %w", err)
	}

	return &Body{
		Cluster: Cluster{
			Name:    "labs",
			Version: ver.String(),
		},
		PodName:      name,
		PodNamespace: namespace,
		Error:        reason,
		PodLogs:      logs,
		PodDescribe:  describe,
	}, nil
}

func MergeConfigs(configs [][]byte) (Config, error) {
	// In our case we don't have complex merge strategy,
	// the last one that was specified wins :)
	finalCfg := Config{}
	for _, rawCfg := range configs {
		var cfg Config
		err := yaml.Unmarshal(rawCfg, &cfg)
		if err != nil {
			return Config{}, err
		}

		if cfg.Repo.Name != "" {
			finalCfg.Repo.Name = cfg.Repo.Name
		}
		if cfg.Repo.Owner != "" {
			finalCfg.Repo.Owner = cfg.Repo.Owner
		}
		if cfg.Token != "" {
			finalCfg.Token = cfg.Token
		}
	}

	return finalCfg, nil
}

func (e *Executor) resolveCommandArgs(ctx context.Context, command string) (string, string, string, error) {
	var opts CreateIssueOptions
	cmd := NewRoot(&opts)
	args, err := shellwords.Parse(command)
	if err != nil {
		return "", "", "", fmt.Errorf("while parsing command: %w", err)
	}
	args = args[1:]
	cmd.SetArgs(args)

	err = cmd.ExecuteContext(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("while resolving execution options: %w", err)
	}

	kind, resourceName, found := strings.Cut(opts.For, "/")
	if !found {
		return "", "", "", fmt.Errorf("wrong syntax, expected --for={kind}/{resourceName}")
	}

	switch strings.ToLower(kind) {
	case "pod":
		if opts.Namespace == "" {
			opts.Namespace = "default"
		}
		return resourceName, opts.Namespace, opts.Reason, nil
	default:
		return "", "", "", fmt.Errorf("unsupported k8s kind %s", kind)
	}
}

func (e *Executor) getPodDescription(ctx context.Context, namespace, name string) (string, error) {
	pod, err := e.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("while getting pod definition: %w", err)
	}
	rawPod, err := yaml.Marshal(pod)
	if err != nil {
		return "", fmt.Errorf("while marshaling Pod: %w", err)
	}
	return string(rawPod), nil
}
