package aslite_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mashiike/aslite"
	"github.com/stretchr/testify/require"
)

func TestStateMachine__Remarshal(t *testing.T) {
	dir, err := os.ReadDir("testdata/asl")
	require.NoError(t, err)
	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		t.Run(file.Name(), func(t *testing.T) {
			bs, err := fs.ReadFile(os.DirFS("."), filepath.Join("testdata/asl", file.Name()))
			require.NoError(t, err)
			var asl aslite.StateMachine
			err = json.Unmarshal(bs, &asl)
			require.NoError(t, err)
			bs2, err := json.Marshal(&asl)
			require.NoError(t, err)
			require.JSONEq(t, string(bs), string(bs2))
		})
	}
}

func TestStateMachine__IsLinearWorkflow(t *testing.T) {
	cases := []struct {
		file string
		want bool
	}{
		{
			file: "testdata/linear-workflow.asl.json",
			want: true,
		},
		{
			file: "testdata/no-linear-workflow-choices.asl.json",
			want: false,
		},
		{
			file: "testdata/linear-workflow-choices.asl.json",
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.file, func(t *testing.T) {
			bs, err := os.ReadFile(tc.file)
			require.NoError(t, err)
			var asl aslite.StateMachine
			err = json.Unmarshal(bs, &asl)
			require.NoError(t, err)
			require.Equal(t, tc.want, asl.IsLinearWorkflow())
		})
	}
}
