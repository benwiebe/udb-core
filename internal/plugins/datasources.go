package plugins

import (
	"context"
	"fmt"

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
			fmt.Printf("Error: plugin %s does not contain datasource %s\n",
				(*plugin).GetName(), dsConfig.DatasourceId)
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
			fmt.Printf("Error starting datasource %s: %v; it will not be wired to any boards\n", id, err)
			continue
		}
		started[id] = ds
	}
	return started
}
