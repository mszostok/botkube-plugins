package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/hashicorp/go-plugin"
	"github.com/kubeshop/botkube/pkg/api/executor"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"go.szostok.io/botkube-plugins/internal/executor/gh"
)

const pluginName = "gh"

// Executor implements Botkube executor plugin.
type Executor struct {
	ghProcessor *gh.Executor
}

// Execute process the gh command
func (e *Executor) Execute(ctx context.Context, req *executor.ExecuteRequest) (*executor.ExecuteResponse, error) {
	cfg, err := gh.MergeConfigs(req.Configs)
	if err != nil {
		return nil, fmt.Errorf("while merging configs: %w", err)
	}

	data, err := e.ghProcessor.Execute(ctx, cfg, req.Command)
	if err != nil {
		return nil, fmt.Errorf("while executing gh command: %w", err)
	}

	return &executor.ExecuteResponse{Data: data}, nil
}

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	exitOnError(err)
	k8sClient, err := kubernetes.NewForConfig(config)
	exitOnError(err)

	e := gh.NewExecutor(k8sClient)

	executor.Serve(map[string]plugin.Plugin{
		pluginName: &executor.Plugin{
			Executor: &Executor{
				ghProcessor: e,
			},
		},
	})
}

func exitOnError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
