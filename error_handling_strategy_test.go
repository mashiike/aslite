package aslite_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mashiike/aslite"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

func TestErrorHandlingStrategy__ApplyToStateMachine(t *testing.T) {
	g := goldie.New(t,
		goldie.WithFixtureDir("testdata/modify_asl"),
		goldie.WithNameSuffix(".golden.json"),
	)
	dir, err := os.ReadDir("testdata/asl")
	require.NoError(t, err)
	ehs := aslite.ErrorHandlingStrategy{
		Retry: []*aslite.Retrier{
			{
				ErrorEquals: []string{
					"States.ALL",
				},
				IntervalSeconds: 1,
				MaxAttempts:     3,
				BackoffRate:     2,
			},
		},
		Catch: []*aslite.ErrorHandlingStrategyCatcher{
			{
				Name: "DefaultCatcher",
				ErrorEquals: []string{
					"States.ALL",
				},
				ResultPath:  "$.error",
				SNSTopicARN: "arn:aws:sns:us-east-1:123456789012:my-topic",
			},
		},
	}

	for _, file := range dir {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}
		base := filepath.Base(file.Name())
		t.Run(base, func(t *testing.T) {
			bs, err := fs.ReadFile(os.DirFS("."), filepath.Join("testdata/asl", file.Name()))
			require.NoError(t, err)
			var asl aslite.StateMachine
			err = json.Unmarshal(bs, &asl)
			require.NoError(t, err)
			err = ehs.ApplyToStateMachine(&asl)
			require.NoError(t, err)
			g.AssertJson(t, base, &asl)
		})
	}
}
