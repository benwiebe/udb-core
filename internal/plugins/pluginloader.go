package plugins

import (
	"log/slog"

	"github.com/benwiebe/udb-core/internal/config"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
	"github.com/benwiebe/udb-plugin-library/types"
)

type PluginList = []*udb_plugin_library.UdbPlugin
type PluginMap = map[string]*udb_plugin_library.UdbPlugin
type PluginData struct {
	List PluginList
	ById PluginMap
}

// LoadPlugins wires registered plugins to their config blocks. All plugins that
// called Register() from their init() are available; the plugins config block
// provides credentials and settings for those that need it. Plugins with no
// config entry receive a nil PluginConfig, which is safe for plugins that need
// no configuration.
func LoadPlugins(pluginsConfig config.PluginsConfig) PluginData {
	configByID := make(map[string]config.PluginConfig, len(pluginsConfig))
	for _, pc := range pluginsConfig {
		configByID[pc.ID] = pc
	}

	registered := udb_plugin_library.Registered()
	pluginList := make(PluginList, 0, len(registered))
	pluginMap := make(PluginMap, len(registered))

	for _, plg := range registered {
		plg := plg
		id := plg.GetId()

		var cfg types.PluginConfig // nil RawMessage — safe for plugins needing no config
		if pc, ok := configByID[id]; ok {
			cfg = pc.Config
		}

		if err := plg.Configure(cfg); err != nil {
			slog.Error("failed to configure plugin", "plugin", id, "err", err)
			continue
		}

		pluginList = append(pluginList, &plg)
		pluginMap[id] = &plg
		slog.Info("plugin registered", "plugin", id)
	}

	return PluginData{List: pluginList, ById: pluginMap}
}
