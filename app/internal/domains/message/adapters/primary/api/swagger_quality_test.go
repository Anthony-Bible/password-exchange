package api

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwaggerQuality(t *testing.T) {
	// This test checks if the swagger.json has detailed descriptions for model fields.
	data, err := ioutil.ReadFile("../../../../../../docs/swagger.json")
	require.NoError(t, err)

	var swagger struct {
		Definitions map[string]struct {
			Properties map[string]struct {
				Description string `json:"description"`
			} `json:"properties"`
		} `json:"definitions"`
	}

	err = json.Unmarshal(data, &swagger)
	require.NoError(t, err)

	healthCheck, ok := swagger.Definitions["models.HealthCheckResponse"]
	require.True(t, ok, "models.HealthCheckResponse definition should exist")

	servicesProp, ok := healthCheck.Properties["services"]
	require.True(t, ok, "services property should exist")

	assert.NotEmpty(t, servicesProp.Description, "services property should have a description annotation")
}
