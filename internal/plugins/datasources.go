package plugins

import (
	"context"
	"log/slog"

	"github.com/benwiebe/udb-core/internal/config"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
	"github.com/benwiebe/udb-plugin-library/types"
)

// LoadDatasources resolves datasource instances from plugin data and config.
func LoadDatasources(pluginData PluginData, dsConfigs config.DatasourcesConfig) map[string]types.Datasource[any] {
	datasourceMap := make(map[string]types.Datasource[any], len(dsConfigs))
	for _, dsConfig := range dsConfigs {
		plugin := pluginData.ById[dsConfig.Plugin]
		if plugin == nil {
			slog.Error("plugin not found or failed to load", "plugin", dsConfig.Plugin)
			continue
		}

		pluginType := (*plugin).GetPluginType()
		if pluginType == types.PluginTypeBoards {
			slog.Error("plugin cannot provide datasources (boards-only)", "plugin", (*plugin).GetName())
			continue
		}

		typedPlugin, ok := (*plugin).(udb_plugin_library.UdbDatasourcePlugin)
		if !ok {
			slog.Error("plugin does not implement UdbDatasourcePlugin", "plugin", (*plugin).GetName())
			continue
		}

		ds := typedPlugin.GetDatasourceMap()[dsConfig.DatasourceId]
		if ds == nil {
			slog.Error("plugin does not contain datasource", "plugin", (*plugin).GetName(), "datasource", dsConfig.DatasourceId)
			continue
		}

		datasourceMap[dsConfig.Id] = ds
	}
	return datasourceMap
}

// StartDatasources calls Start on each datasource. Datasources that fail to start are removed
// from the returned map so they are not wired to boards.
func StartDatasources(ctx context.Context, datasourceMap map[string]types.Datasource[any]) map[string]types.Datasource[any] {
	started := make(map[string]types.Datasource[any], len(datasourceMap))
	for id, ds := range datasourceMap {
		if err := ds.Start(ctx); err != nil {
			slog.Error("failed to start datasource; it will not be wired to any boards", "datasource", id, "err", err)
			continue
		}
		started[id] = ds
	}
	return started
}
