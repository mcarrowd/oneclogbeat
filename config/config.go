package config

import (
	"fmt"
	"sort"
	"strings"

	"github.com/joeshaw/multierror"
)

const (
	DefaultRegistryFile = ".oneclogbeat.yml"
)

type Validator interface {
	Validate() error
}

type Settings struct {
	Oneclogbeat OneclogbeatConfig      `config:"oneclogbeat"`
	Raw         map[string]interface{} `config:",inline"`
}

func (s Settings) Validate() error {
	validKeys := []string{"filter", "logging", "output", "oneclogbeat"}
	sort.Strings(validKeys)
	// Check for invalid top-level keys.
	var errs multierror.Errors
	for k := range s.Raw {
		k = strings.ToLower(k)
		i := sort.SearchStrings(validKeys, k)
		if i >= len(validKeys) || validKeys[i] != k {
			errs = append(errs, fmt.Errorf("Invalid top-level key '%s' "+
				"found. Valid keys are %s", k, strings.Join(validKeys, ", ")))
		}
	}
	err := s.Oneclogbeat.Validate()
	if err != nil {
		errs = append(errs, err)
	}
	return errs.Err()
}

type OneclogbeatConfig struct {
	Eventlogs    []EventlogConfig `config:"event_logs"`
	RegistryFile string           `config:"registry_file"`
}

type EventlogConfig struct {
	Name string `config:"name"`
	Path string `config:"path"`
}

func (obc OneclogbeatConfig) Validate() error {
	var errs multierror.Errors
	if len(obc.Eventlogs) == 0 {
		errs = append(errs, fmt.Errorf("At least one event log must be configured as part of event_logs"))
	}
	return errs.Err()
}
