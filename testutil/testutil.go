package testutil

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func AssertMatchJSON(t *testing.T, expected map[string]any, actual string) {
	t.Helper()

	expectedJSON, err := json.MarshalIndent(expected, "", "\t")
	assert.NoError(t, err)

	var actualJSON = make(map[string]string)
	err = json.Unmarshal([]byte(actual), &actualJSON)
	assert.NoError(t, err)
	actualJSONBytes, err := json.MarshalIndent(actualJSON, "", "\t")
	assert.NoError(t, err)

	assert.Equal(t, string(expectedJSON), string(actualJSONBytes))
}
