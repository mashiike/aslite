package aslite_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mashiike/aslite"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflowGenerate(t *testing.T) {
	g := goldie.New(t,
		goldie.WithFixtureDir("testdata/generated_asl"),
		goldie.WithNameSuffix(".golden.json"),
	)
	dir, err := os.ReadDir("testdata/workflow")
	require.NoError(t, err)
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}
		base := filepath.Base(file.Name())
		t.Run(base, func(t *testing.T) {
			bs, err := fs.ReadFile(os.DirFS("."), filepath.Join("testdata/workflow", file.Name()))
			require.NoError(t, err)
			var w aslite.Workflow
			err = yaml.Unmarshal(bs, &w)
			require.NoError(t, err)
			require.NoError(t, w.Validate())
			asl, err := w.NewStateMachine()
			require.NoError(t, err)
			g.AssertJson(t, base, &asl)
		})
	}
}
