package exec

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/kubeshop/botkube/pkg/api"

	"go.szostok.io/botkube-plugins/internal/exec/template"
)

// Render is an interface that knows how to render a given command output.
type Render interface {
	// RenderMessage receives command output and a template and produce a final message.
	RenderMessage(cmd, output string, msgCtx *template.Interactive) (api.Message, error)
}

// Renderer provides functionality to render command output in requested format.
type Renderer struct {
	mux      sync.RWMutex
	renderer map[string]Render
}

// NewRenderer returns a new Renderer instance.
func NewRenderer() *Renderer {
	return &Renderer{
		renderer: map[string]Render{},
	}
}

func (r *Renderer) Register(name string, render Render) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	_, found := r.renderer[name]
	if found {
		return fmt.Errorf("conflicts: %q was already registered", name)
	}
	r.renderer[name] = render
	return nil
}

// Get return renderer for a given output
func (r *Renderer) Get(output string) (Render, error) {
	r.mux.RLock()
	defer r.mux.RUnlock()

	printer, found := r.renderer[output]
	if !found {
		return nil, fmt.Errorf("formatter %q is not available, allowed formatters %q", output, r.availablePrinters())
	}
	return printer, nil
}

func (r *Renderer) availablePrinters() string {
	out := make([]string, 0, len(r.renderer))
	for key := range r.renderer {
		out = append(out, key)
	}

	sort.Strings(out)
	return strings.Join(out, " | ")
}
