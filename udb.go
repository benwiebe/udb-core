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

	/**** Determine Datasources Required by Config ****/
	datasourceMap := make(map[string]types.Datasource[any], len(appConfig.Datasources))
	for _, dsConfig := range appConfig.Datasources {
		plugin := pluginData.ById[dsConfig.Plugin]
		if plugin == nil {
			fmt.Printf("Error: plugin %s not found or failed to load\n", dsConfig.Plugin)
			continue
		}

		pluginType := (*plugin).GetPluginType()
		if pluginType == types.PluginTypeBoards {
			fmt.Printf("Error: plugin %s is a boards-only plugin and cannot provide datasources\n",
				(*plugin).GetName())
			continue
		}

		typedPlugin, ok := (*plugin).(udb_plugin_library.UdbDatasourcePlugin)
		if !ok {
			fmt.Printf("Error: plugin %s does not implement UdbDatasourcePlugin interface\n",
				(*plugin).GetName())
			continue
		}

		ds := typedPlugin.GetDatasourceMap()[dsConfig.DatasourceId]
		if ds == nil {
			fmt.Printf("Error: plugin %s does not contain datasource %s\n", (*plugin).GetName(), dsConfig.DatasourceId)
			continue
		}

		datasourceMap[dsConfig.Id] = ds
	}
	fmt.Println("Datasources loaded")

	/**** Determine Boards Required by Config ****/
	type boardEntry struct {
		board      types.Board[any]
		config     config.BoardConfig
		datasource types.Datasource[any]
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

			board := typedPlugin.GetBoardMap()[boardConfig.BoardId]
			if board == nil {
				fmt.Printf("Error: plugin %s does not contain board %s\n", (*plugin).GetName(), boardConfig.BoardId)
				continue
			}

			boards = append(boards, boardEntry{board: board, config: boardConfig})
		}
	}

	/**** Setup Data Sources ****/
	for i, entry := range boards {
		if entry.config.Datasource != "" {
			// Explicit datasource reference in board config.
			ds, found := datasourceMap[entry.config.Datasource]
			if !found {
				fmt.Printf("Error: datasource %q not found for board %s\n", entry.config.Datasource, entry.config.BoardId)
				continue
			}
			boards[i].datasource = ds
			continue
		}

		// No explicit datasource — auto-determine by matching GetDatasourceType() to GetType().
		requiredType := entry.board.GetDatasourceType()
		if requiredType == "" {
			continue // board needs no datasource
		}
		var matches []types.Datasource[any]
		for _, ds := range datasourceMap {
			if ds.GetType() == requiredType {
				matches = append(matches, ds)
			}
		}
		switch len(matches) {
		case 1:
			boards[i].datasource = matches[0]
			fmt.Printf("Auto-wired datasource %q for board %s\n", requiredType, entry.config.BoardId)
		case 0:
			fmt.Printf("Warning: no datasource of type %q found for board %s; board will have no data\n",
				requiredType, entry.config.BoardId)
		default:
			fmt.Printf("Warning: multiple datasources of type %q available for board %s; specify datasource explicitly\n",
				requiredType, entry.config.BoardId)
		}
	}

	/**** Setup Boards ****/
	initializedBoards := boards[:0]
	for _, entry := range boards {
		if err := entry.board.Init(entry.datasource); err != nil {
			fmt.Printf("Error initializing board %s: %v\n", entry.config.BoardId, err)
			continue
		}
		initializedBoards = append(initializedBoards, entry)
	}
	fmt.Printf("%d board(s) initialized\n", len(initializedBoards))

	/**** Initialize Display ****/
	var displayInstance display.Display = display.InitializeDisplay(appConfig.Display)

	/**** Main Boards Display Loop ****/
	// TODO: this

	/**** Shutdown ****/
	displayInstance.CloseDisplay()
}
