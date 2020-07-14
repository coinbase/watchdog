package config

import (
	"context"
	"os"
	"strings"
	"sync"

	"github.com/coinbase/watchdog/primitives/datadog/types"
	"github.com/coinbase/watchdog/primitives/git"

	"github.com/caarlos0/env"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

// UserConfig defines an interface to load the user configuration.
type UserConfig interface {
	// Reload the user config if the config as updated
	Reload() error

	// UserConfigFilesByComponentID takes a component, id and returns a list of user config files
	// which contain this data.
	UserConfigFilesByComponentID(component types.Component, id string) []*UserConfigFile

	// UserConfigFiles returns a slice of user config files.
	UserConfigFiles() []*UserConfigFile

	// GetUserConfigBasePath returns a base path of user config
	GetUserConfigBasePath() string

	// UserConfigFromFile reads a file from filesystem and returns a UserConfigFile object.
	UserConfigFromFile(path string, pullMaster bool) (*UserConfigFile, error)
}

// NewUserConfigFromGit returns a new instance of a user config from a git repository.
func NewUserConfigFromGit(ctx context.Context) (UserConfig, error) {
	cfg := &fromEnvVar{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse user config parameters from environment variables")
	}

	git, err := git.New(git.WithRSAKey(cfg.GitSSHUser, cfg.GitSSHPassword, []byte(cfg.PrivateKey)), git.WithIgnoreKnownHosts(cfg.IgnoreKnownHosts))
	if err != nil {
		return nil, errors.Wrap(err, "unable to create a new instance of git")
	}

	err = git.Clone(cfg.GitURL, os.Stdout)
	if err != nil {
		return nil, errors.Wrap(err, "unable to clone a repo")
	}

	userCfg := &userGitConfig{
		basePath:     cfg.BaseConfigPath,
		readDirFn:    git.ReadDir,
		readFileFn:   git.ReadFile,
		pullMasterFn: git.PullMaster,

		dashboards:   make(map[string][]*UserConfigFile),
		monitors:     make(map[string][]*UserConfigFile),
		screenboards: make(map[string][]*UserConfigFile),
		downtimes:    make(map[string][]*UserConfigFile),
	}

	return userCfg, userCfg.Reload()
}

// UserConfigFile represents a watchdog config file by a user.
type UserConfigFile struct {
	Meta MetaData

	Dashboards   []string
	Monitors     []string
	Downtimes    []string
	ScreenBoards []string
}

// Components return a mapping of a component to its IDs from a user config file.
func (u UserConfigFile) Components() map[types.Component][]string {
	return map[types.Component][]string{
		types.ComponentDashboard:   u.Dashboards,
		types.ComponentMonitor:     u.Monitors,
		types.ComponentScreenboard: u.ScreenBoards,
		types.ComponentDowntime:    u.Downtimes,
	}
}

// MetaData is a field which holds a user provided metadata.
// Team is a name of a team responsible for a config.
// Project is an name of a project, used in component name, optional.
type MetaData struct {
	Team    string
	Project string
	Slack   string

	FilePath string
}

type fromEnvVar struct {
	// BaseConfigPath is a base path in git repository where users store config files.
	BaseConfigPath string `env:"USER_CONFIG_PATH" envDefault:"/config"`

	// GitURL is a URL to git repo.
	GitURL string `env:"USER_CONFIG_GIT_URL,required"`

	// GitSSHUser is used to configure username to clone a repo over ssh.
	GitSSHUser string `env:"USER_CONFIG_GIT_USER" envDefault:"git"`

	// GitSSHPassword is used to configure a password to clone a repo over ssh.
	GitSSHPassword string `env:"USER_CONFIG_GIT_PASSWORD"`

	// PrivateKey is a key used to clone the repository with configs.
	PrivateKey string `env:"USER_CONFIG_GIT_PRIVATE_KEY,required"`

	// IgnoreKnownHosts is an option to ignore or respect the ssh known hosts when cloning repo over ssh.
	// If set to false, the file from `SSH_KNOWN_HOSTS` env variable will be used.
	// Default to ignore
	IgnoreKnownHosts bool `env:"USER_IGNORE_KNOWN_HOSTS" envDefault:"true"`
}

// userGitConfig is an implementation of a UserConfig interface which has
// will retrieve the user configuration from git repo.
type userGitConfig struct {
	sync.Mutex

	url      string
	basePath string

	dashboards   map[string][]*UserConfigFile
	screenboards map[string][]*UserConfigFile
	monitors     map[string][]*UserConfigFile
	downtimes    map[string][]*UserConfigFile

	userConfigFiles []*UserConfigFile

	readFileFn   func(string) ([]byte, error)
	readDirFn    func(string) ([]os.FileInfo, error)
	pullMasterFn func() error
}

// UserConfigFromFile reads a file from filesystem and returns a UserConfigFile object.
func (u *userGitConfig) UserConfigFromFile(path string, pullMaster bool) (*UserConfigFile, error) {
	if pullMaster {
		if err := u.pullMasterFn(); err != nil {
			return nil, err
		}
	}
	cfg := &UserConfigFile{}
	body, err := u.readFileFn(path)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(body, cfg)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to unmarshal user config %s", path)
	}

	return cfg, nil
}

// GetUserConfigBasePath returns a base path to user configs trimming leading slash.
func (u *userGitConfig) GetUserConfigBasePath() string {
	return strings.TrimLeft(u.basePath, "/")
}

// UserConfigFiles returns a slice of user config files
func (u *userGitConfig) UserConfigFiles() []*UserConfigFile {
	return u.userConfigFiles
}

func (u *userGitConfig) mapToSlice(m map[string][]*MetaData) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}

	return out
}

// Metadata returns a list of metadata values for a given component and id.
func (u *userGitConfig) UserConfigFilesByComponentID(component types.Component, id string) []*UserConfigFile {
	switch component {
	case types.ComponentDashboard:
		return u.dashboards[id]
	case types.ComponentMonitor:
		return u.monitors[id]
	case types.ComponentScreenboard:
		return u.screenboards[id]
	case types.ComponentDowntime:
		return u.downtimes[id]
	default:
		return nil
	}
}

// Reload the user config in run time.
func (u *userGitConfig) Reload() error {
	u.Lock()
	defer u.Unlock()

	logrus.Infof("Loading a config from git repo %s", u.url)

	u.dashboards = make(map[string][]*UserConfigFile)
	u.monitors = make(map[string][]*UserConfigFile)
	u.screenboards = make(map[string][]*UserConfigFile)
	u.downtimes = make(map[string][]*UserConfigFile)

	u.userConfigFiles = []*UserConfigFile{}

	err := u.pullMasterFn()
	if err != nil {
		return err
	}

	files, err := u.findConfigs(u.basePath)
	if err != nil {
		return err
	}

	return u.readConfigs(files)
}

func (u *userGitConfig) resolveFileNames(files []os.FileInfo) string {
	var tmpNames []string
	for _, file := range files {
		tmpNames = append(tmpNames, file.Name())
	}

	return strings.Join(tmpNames, ", ")
}

func (u *userGitConfig) readConfigs(configs []wrappedFileInfo) error {
	for _, config := range configs {
		userConfigFile, err := u.UserConfigFromFile(config.path, false)
		if err != nil {
			return errors.Wrapf(err, "unable to read user config: %s", config.path)
		}

		userConfigFile.Meta.FilePath = config.path

		u.updateConfig(userConfigFile)
	}

	return nil
}

func (u *userGitConfig) updateConfig(cfgFile *UserConfigFile) {
	u.updateComponent(cfgFile.Dashboards, cfgFile, u.dashboards)
	u.updateComponent(cfgFile.Monitors, cfgFile, u.monitors)
	u.updateComponent(cfgFile.ScreenBoards, cfgFile, u.screenboards)
	u.updateComponent(cfgFile.Downtimes, cfgFile, u.downtimes)

	u.userConfigFiles = append(u.userConfigFiles, cfgFile)
}

func (u *userGitConfig) updateComponent(ids []string, userCfg *UserConfigFile, component map[string][]*UserConfigFile) {
	for _, id := range ids {
		component[id] = append(component[id], userCfg)
	}
}

type wrappedFileInfo struct {
	os.FileInfo
	path string
}

func (u *userGitConfig) findConfigs(path string) ([]wrappedFileInfo, error) {
	items, err := u.readDirFn(path)
	if err != nil {
		return nil, err
	}

	logrus.Infof("found files/dirs in folder %s: %s", path, u.resolveFileNames(items))

	var configFiles []wrappedFileInfo
	for _, item := range items {
		fullPath := strings.Join([]string{path, item.Name()}, "/")
		if item.IsDir() {
			nestedItems, err := u.findConfigs(fullPath)
			if err != nil {
				return nil, err
			}

			configFiles = append(configFiles, nestedItems...)
		}

		if strings.HasSuffix(item.Name(), ".yaml") || strings.HasSuffix(item.Name(), ".yml") {
			logrus.Infof("Found a config file %s", item.Name())
			configFiles = append(configFiles, wrappedFileInfo{item, fullPath})
		}
	}

	return configFiles, nil
}
