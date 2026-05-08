package main

import (
	"context"
	"fmt"
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
	fmt.Println("Starting Universal Display Board...")

	configLoader := config.NewDefaultConfigLoader()
	if err := configLoader.Load(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	appConfig := configLoader.GetConfig()
	fmt.Println("Config loaded")

	pluginData := plugins.LoadPlugins(appConfig.Plugins)
	fmt.Println("Plugins loaded")

	datasourceMap := plugins.LoadDatasources(pluginData, appConfig.Datasources)
	fmt.Println("Datasources loaded")

	boards := plugins.LoadBoards(pluginData, appConfig.Boards)
	plugins.WireDatasources(boards, datasourceMap)
	initializedBoards := plugins.InitBoards(boards)

	displayInstance, err := display.NewDisplay(appConfig.Display)
	if err != nil {
		fmt.Printf("Failed to initialize display: %v\n", err)
		return
	}
	defer displayInstance.CloseDisplay()

	dims := types.BoardDimensions{
		Width:  appConfig.Display.Width,
		Height: appConfig.Display.Height,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	if len(initializedBoards) == 0 {
		fmt.Println("Warning: no boards to display")
		return
	}

	fmt.Println("Display loop started, press Ctrl+C to stop")
	scheduler.Run(ctx, displayInstance, initializedBoards, dims)
}
