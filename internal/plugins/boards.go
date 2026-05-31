package plugins

import (
	"log/slog"

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
			slog.Error("plugin not found or failed to load", "plugin", boardConfig.Plugin)
			continue
		}

		pluginType := (*plugin).GetPluginType()
		if pluginType == types.PluginTypeDatasource {
			slog.Error("plugin cannot provide boards (datasource-only)", "plugin", (*plugin).GetName())
			continue
		}

		if pluginType == types.PluginTypeBoards || pluginType == types.PluginTypeCombined {
			typedPlugin, ok := (*plugin).(udb_plugin_library.UdbBoardPlugin)
			if !ok {
				slog.Error("plugin does not implement UdbBoardPlugin", "plugin", (*plugin).GetName())
				continue
			}

			b := typedPlugin.GetBoardMap()[boardConfig.BoardId]
			if b == nil {
				slog.Error("plugin does not contain board", "plugin", (*plugin).GetName(), "board", boardConfig.BoardId)
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
				slog.Error("datasource not found for board", "datasource", entry.Config.Datasource, "board", entry.Config.BoardId)
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
			slog.Info("auto-wired datasource", "type", requiredType, "board", entry.Config.BoardId)
		case 0:
			slog.Warn("no datasource found for board; board will have no data", "type", requiredType, "board", entry.Config.BoardId)
		default:
			slog.Warn("multiple datasources match board type; specify datasource explicitly", "type", requiredType, "board", entry.Config.BoardId)
		}
	}
}

// InitBoards initializes each board with the given display dimensions and returns only those that succeeded.
func InitBoards(boards []BoardEntry, dims types.BoardDimensions) []BoardEntry {
	initialized := make([]BoardEntry, 0, len(boards))
	for _, entry := range boards {
		if err := entry.Board.Init(entry.Config.Config, entry.Datasource, dims); err != nil {
			slog.Error("failed to initialize board", "board", entry.Config.BoardId, "err", err)
			continue
		}
		initialized = append(initialized, entry)
	}
	slog.Info("boards initialized", "count", len(initialized))
	return initialized
}
