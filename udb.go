package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/benwiebe/udb-core/internal/config"
	"github.com/benwiebe/udb-core/internal/display"
	"github.com/benwiebe/udb-core/internal/plugins"
	"github.com/benwiebe/udb-core/internal/scheduler"
	"github.com/benwiebe/udb-plugin-library/types"
)

func main() {
	slog.Info("starting Universal Display Board")

	configLoader := config.NewDefaultConfigLoader()
	if err := configLoader.Load(); err != nil {
		slog.Error("failed to load config", "err", err)
		return
	}
	appConfig := configLoader.GetConfig()
	slog.Info("config loaded")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		slog.Info("shutting down")
		cancel()
	}()

	pluginData := plugins.LoadPlugins(appConfig.Plugins)
	slog.Info("plugins loaded")

	datasourceMap := plugins.LoadDatasources(pluginData, appConfig.Datasources)
	datasourceMap = plugins.StartDatasources(ctx, datasourceMap)
	slog.Info("datasources started")

	dims := types.BoardDimensions{
		Width:  appConfig.Display.Width,
		Height: appConfig.Display.Height,
	}

	boards := plugins.LoadBoards(pluginData, appConfig.Boards)
	plugins.WireDatasources(boards, datasourceMap)
	initializedBoards := plugins.InitBoards(boards, dims)

	displayInstance, err := display.NewDisplay(appConfig.Display)
	if err != nil {
		slog.Error("failed to initialize display", "err", err)
		return
	}
	defer displayInstance.CloseDisplay()

	if len(initializedBoards) == 0 {
		slog.Warn("no boards to display")
		return
	}

	slog.Info("display loop started")
	scheduler.Run(ctx, displayInstance, initializedBoards)
}
