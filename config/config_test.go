// +build !integration

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type validationTestCase struct {
	config Validator
	errMsg string
}

func (v validationTestCase) run(t *testing.T) {
	if v.errMsg == "" {
		assert.NoError(t, v.config.Validate())
	} else {
		err := v.config.Validate()
		if assert.Error(t, err, "expected '%s'", v.errMsg) {
			assert.Contains(t, err.Error(), v.errMsg)
		}
	}
}

func TestConfigValidate(t *testing.T) {
	testCases := []validationTestCase{
		// Top-level config
		{
			OneclogbeatConfig{
				Eventlogs: []EventlogConfig{
					{Name: "App"},
				},
			},
			"", // No Error
		},
		{
			Settings{
				OneclogbeatConfig{
					Eventlogs: []EventlogConfig{
						{Name: "App"},
					},
				},
				map[string]interface{}{"other": "value"},
			},
			"1 error: Invalid top-level key 'other' found. Valid keys are " +
				"filter, logging, oneclogbeat, output",
		},
		{
			OneclogbeatConfig{},
			"1 error: At least one event log must be configured as part of " +
				"event_logs",
		},
	}
	for _, test := range testCases {
		test.run(t)
	}
}
