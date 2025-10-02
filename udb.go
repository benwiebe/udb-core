package main

import (
	"fmt"

	"github.com/benwiebe/udb-core/internal/config"
	"github.com/benwiebe/udb-core/internal/display"
	"github.com/benwiebe/udb-core/internal/plugins"
)

func main() {
	fmt.Println("Starting Universal Display Board...")
	/**** Load App Config ****/
	configLoader := config.NewDefaultConfigLoader()
	if err := configLoader.Load(); err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}
	appConfig := configLoader.GetConfig()
	fmt.Println("Config loaded")

	/**** Load Plugins ****/
	_ = plugins.LoadPlugins(appConfig.Plugins)
	fmt.Println("Plugins loaded")

	/**** Determine Boards and Data Sources Required by Config ****/
	// TODO: this

	/**** Setup Data Sources ****/
	// TODO: this

	/**** Setup Boards ****/
	// TODO: this

	/**** Initialize Display ****/
	displayInstance := display.InitializeDisplayWithConfig(appConfig.Display.ConvertToHardwareConfig())

	/**** Main Boards Display Loop ****/
	// TODO: this

	/**** Shutdown ****/
	displayInstance.CloseDisplay()

}
