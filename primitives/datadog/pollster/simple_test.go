package pollster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/primitives/datadog/types"
)

type fakeUserConfig struct {
}

func (c fakeUserConfig) UserConfigFilesByComponentID(component types.Component, id string) []*config.UserConfigFile {
	return []*config.UserConfigFile{
		&config.UserConfigFile{
			Meta: config.MetaData{
				FilePath: "foo/bar",
			},
		},
		&config.UserConfigFile{
			Meta: config.MetaData{
				FilePath: "foo/bar2",
			},
		},
	}
}

func (c fakeUserConfig) GetUserConfigBasePath() string {
	return ""
}

func (c fakeUserConfig) Reload() error {
	return nil
}

func (c fakeUserConfig) UserConfigFiles() []*config.UserConfigFile {
	return nil
}

func (c fakeUserConfig) UserConfigFromFile(path string, a bool) (*config.UserConfigFile, error) {
	return nil, nil
}

func TestNewSimplePollster(t *testing.T) {
	modified := time.Now().Add(time.Second).Format(time.RFC3339Nano)

	dashboard := fmt.Sprintf(`{"dashboards":[{"id":"1","modified":"%s"}]}`, modified)
	monitors := fmt.Sprintf(`[{"id":2,"modified":"%s"}]`, modified)
	screenboards := fmt.Sprintf(`{"screenboards":[{"id":3,"modified":"%s"}]}`, modified)

	p := &simplePoller{
		interval: time.Millisecond * 100,
		ca: &componentAccessors{
			getDashboards: func() (client.DashboardsResponse, error) {
				return []byte(dashboard), nil
			},
			getMonitors: func() (client.MonitorsResponse, error) {
				return []byte(monitors), nil
			},
			getScreenBoards: func() (client.ScreenBoardsResponse, error) {
				return []byte(screenboards), nil
			},
		},

		cfg: &config.Config{
			UserConfig: &fakeUserConfig{},
		},
	}

	ch := p.Do(context.Background())
	var cfgFiles []*config.UserConfigFile
	for {
		select {
		case value := <-ch:
			cfgFiles = append(cfgFiles, value.UserConfigFile)
			if len(cfgFiles) == 2 {
				if cfgFiles[0].Meta.FilePath != "foo/bar" {
					t.Fatalf("expect config file boo/bar. Got %s", cfgFiles[0].Meta.FilePath)
				}

				if cfgFiles[1].Meta.FilePath != "foo/bar2" {
					t.Fatalf("expect config file boo/bar. Got %s", cfgFiles[1].Meta.FilePath)
				}
				return
			}

		case <-time.After(time.Second * 2):
			t.Fatal("time out waiting for channel")
		}
	}

}
