package main

import (
	"context"
	"fmt"
	"os"

	"github.com/coinbase/watchdog/config"
	"github.com/coinbase/watchdog/controller"
	"github.com/coinbase/watchdog/primitives/datadog/client"
	"github.com/coinbase/watchdog/server"

	"github.com/sirupsen/logrus"
)

func main() {
	version, err := NewVersion()
	if err != nil {
		logrus.Errorf("version was not set. Please use Makefile to set build SHA and time.")
	}
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version.String())
		return
	}

	cfg, err := config.NewConfig(context.Background(), nil, nil)
	if err != nil {
		logrus.Fatalf("unable to initialize a config: %s", err)
	}

	// enable the json logging if value is set
	if cfg.GetLoggingJSON() {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}

	// redirect all logging output to stdout
	logrus.SetOutput(os.Stdout)

	// set the requested logging level
	loggingLevel := cfg.GetLoggingLevel()
	if loggingLevel != "" {
		level, err := logrus.ParseLevel(loggingLevel)
		if err != nil {
			logrus.Fatalf("unable to parse logging level %s: %s", loggingLevel, err)
		}

		logrus.SetLevel(level)
	}

	// setup default fields to be remove from dashboard/monitor/screen board response.
	clientOptions := []client.Option{
		client.WithRemoveDashboardFields([]string{"dash.modified"}),
		client.WithRemoveMonitorFields([]string{"modified", "overall_state", "overall_state_modified"}, []string{"state"}),
		client.WithRemoveScreenBoardFields([]string{"modified"}),
	}

	// construct the controller options
	options := []controller.Option{
		controller.WithDatadog(cfg.GetDatadogAPIKey(), cfg.GetDatadogAPPKey(), clientOptions...),
		controller.WithGithub(cfg.GetGithubProjectOwner(), cfg.GetGithubRepo(), cfg.GithubAPIURL(),
			cfg.GetGithubIntegrationID(), cfg.GetGithubAppInstallationID(), cfg.GithubAppPrivateKeyBytes()),
		controller.WithSSHGit(cfg.GitURL(), cfg.GitUser(), cfg.GitEmail(), cfg.GithubAppPrivateKeyBytes(), cfg.GetIgnoreKnownHosts()),
	}

	// use polling scheduler based on config
	// TODO: refactor this part
	switch cfg.GetDatadogPollingScheduler() {
	case "simple":
		options = append(options, controller.WithSimplePollster(cfg.GetDatadogPollingInterval(), cfg))
	default:
		logrus.Fatalf("invalid datadog polling scheduler %s", cfg.GetDatadogPollingScheduler())
	}

	// initialize a new watchdog controller
	c, err := controller.New(cfg, options...)
	if err != nil {
		logrus.Fatalf("unable to initialize controller %s", err)
	}

	// Start the polling scheduler in the background
	go c.PollDatadog(context.Background())

	// setup http router
	routerOpts := []server.Option{
		server.WithController(c),
		server.WithGithubWebhook(cfg.GetGithubWebhookSecret()),
	}

	if version != nil {
		routerOpts = append(routerOpts, server.WithVersion(version))
	}

	router, err := server.New(cfg, routerOpts...)
	if err != nil {
		logrus.Fatalf("unable to create a new HTTP server %s", err)
	}

	err = router.Start()
	if err != nil {
		logrus.Fatalf("unable to start an HTTP server %s", err)
	}
}
