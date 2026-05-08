package plugins

import (
	"fmt"

	"github.com/benwiebe/udb-core/internal/config"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
	"github.com/benwiebe/udb-plugin-library/types"
)

// BoardEntry holds a resolved board with its config and wired datasource.
type BoardEntry struct {
	Board      types.Board[any]
	Config     config.BoardConfig
	Datasource types.Datasource[any]
}

// LoadBoards resolves board instances from plugin data and config.
func LoadBoards(pluginData PluginData, boardConfigs config.BoardsConfig) []BoardEntry {
	boards := make([]BoardEntry, 0, len(boardConfigs))
	for _, boardConfig := range boardConfigs {
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

			b := typedPlugin.GetBoardMap()[boardConfig.BoardId]
			if b == nil {
				fmt.Printf("Error: plugin %s does not contain board %s\n",
					(*plugin).GetName(), boardConfig.BoardId)
				continue
			}

			boards = append(boards, BoardEntry{Board: b, Config: boardConfig})
		}
	}
	return boards
}

// WireDatasources assigns datasources to boards, either by explicit reference or by type matching.
func WireDatasources(boards []BoardEntry, datasourceMap map[string]types.Datasource[any]) {
	for i, entry := range boards {
		if entry.Config.Datasource != "" {
			ds, found := datasourceMap[entry.Config.Datasource]
			if !found {
				fmt.Printf("Error: datasource %q not found for board %s\n",
					entry.Config.Datasource, entry.Config.BoardId)
				continue
			}
			boards[i].Datasource = ds
			continue
		}

		// No explicit datasource — auto-determine by matching GetDatasourceType() to GetType().
		requiredType := entry.Board.GetDatasourceType()
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
			boards[i].Datasource = matches[0]
			fmt.Printf("Auto-wired datasource %q for board %s\n", requiredType, entry.Config.BoardId)
		case 0:
			fmt.Printf("Warning: no datasource of type %q found for board %s; board will have no data\n",
				requiredType, entry.Config.BoardId)
		default:
			fmt.Printf("Warning: multiple datasources of type %q available for board %s; specify datasource explicitly\n",
				requiredType, entry.Config.BoardId)
		}
	}
}

// InitBoards initializes each board with the given display dimensions and returns only those that succeeded.
func InitBoards(boards []BoardEntry, dims types.BoardDimensions) []BoardEntry {
	initialized := make([]BoardEntry, 0, len(boards))
	for _, entry := range boards {
		if err := entry.Board.Init(entry.Config.Config, entry.Datasource, dims); err != nil {
			fmt.Printf("Error initializing board %s: %v\n", entry.Config.BoardId, err)
			continue
		}
		initialized = append(initialized, entry)
	}
	fmt.Printf("%d board(s) initialized\n", len(initialized))
	return initialized
}
