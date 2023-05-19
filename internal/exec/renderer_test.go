package exec_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	x "go.szostok.io/botkube-plugins/internal/exec"
	"go.szostok.io/botkube-plugins/internal/exec/output"
)

func TestRendererGet(t *testing.T) {

	renderer := x.NewRenderer()
	err := renderer.Register("parser:table:.*", output.NewTableCommandParser(nil))
	require.NoError(t, err)
	err = renderer.Register("wrapper", output.NewCommandWrapper(nil))
	require.NoError(t, err)

	get, err := renderer.Get("parser:table:char:|")
	require.NoError(t, err)
	assert.NotNil(t, get)
}
