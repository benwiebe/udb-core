package main

import (
	"fmt"

	"github.com/benwiebe/udb-core/internal/config"
	"github.com/benwiebe/udb-core/internal/display"
	"github.com/benwiebe/udb-core/internal/plugins"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
	"github.com/benwiebe/udb-plugin-library/types"
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
	pluginData := plugins.LoadPlugins(appConfig.Plugins)
	fmt.Println("Plugins loaded")

	/**** Determine Boards and Data Sources Required by Config ****/
	type boardEntry struct {
		board  types.Board[any]
		config config.BoardConfig
	}
	boards := make([]boardEntry, 0, len(appConfig.Boards))
	for _, boardConfig := range appConfig.Boards {
		plugin := pluginData.ById[boardConfig.Plugin]
		if plugin == nil {
			fmt.Printf("Error: plugin %s not found or failed to load\n", boardConfig.Plugin)
			continue
		}
		pluginType := (*plugin).GetPluginType()
		if pluginType == types.PluginTypeDatasource {
			fmt.Printf("Error: plugin %s is a datasource-only plugin and cannot provide boards\n",
				(*plugin).GetName())
			continue
		}
		if pluginType == types.PluginTypeBoards || pluginType == types.PluginTypeCombined {
			typedPlugin, ok := (*plugin).(udb_plugin_library.UdbBoardPlugin)
			if !ok {
				fmt.Printf("Error: plugin %s does not implement UdbBoardPlugin interface\n",
					(*plugin).GetName())
				continue
			}
			board := typedPlugin.GetBoardMap()[boardConfig.ID]
			if board == nil {
				fmt.Printf("Error: plugin %s does not contain board %s\n", (*plugin).GetName(), boardConfig.ID)
				continue
			}
			boards = append(boards, boardEntry{board: board, config: boardConfig})
		}
	}

	/**** Setup Data Sources ****/
	// TODO: set up datasources

	/**** Setup Boards ****/
	// TODO: this

	/**** Initialize Display ****/
	displayInstance := display.InitializeDisplay(appConfig.Display)

	/**** Main Boards Display Loop ****/
	// TODO: this

	/**** Shutdown ****/
	displayInstance.CloseDisplay()

}
