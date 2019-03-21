package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/coinbase/watchdog/primitives/datadog/types"

	"github.com/pkg/errors"
)

// NewConfig returns a new instance of Config object.
func NewConfig(ctx context.Context, sysCfg SystemConfig, userCfg UserConfig) (*Config, error) {
	var err error

	if sysCfg == nil {
		sysCfg, err = NewSystemConfig()
		if err != nil {
			return nil, errors.Wrap(err, "unable to create a new system config")
		}
	}

	if userCfg == nil {
		userCfg, err = NewUserConfigFromGit(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "unable to create a new user config")
		}
	}

	return &Config{
		SystemConfig: sysCfg,
		UserConfig:   userCfg,
	}, nil
}

// Config represents a config structure for the application.
type Config struct {
	SystemConfig
	UserConfig
}

// ComponentPath returns a path to a component json representation.
func (c *Config) ComponentPath(component types.Component, team, project string, id int) string {
	destDir := strings.Join([]string{c.SystemConfig.GetDatadogDataPath(), team}, "/")
	filename := fmt.Sprintf("%s/%s-%d.json", destDir, component, id)

	if project != "" {
		filename = fmt.Sprintf("%s/%s/%s-%d.json", destDir, project, component, id)
	}

	return filename
}
