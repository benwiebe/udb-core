package plugins

import (
	"log/slog"
	"plugin"

	"github.com/benwiebe/udb-core/internal/config"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
)

type PluginList = []*udb_plugin_library.UdbPlugin
type PluginMap = map[string]*udb_plugin_library.UdbPlugin
type PluginData struct {
	List PluginList
	ById PluginMap
}

func GetPluginPath(pluginConfig config.PluginConfig) string {
	if pluginConfig.Path != "" {
		return pluginConfig.Path
	}
	return "./plugins/" + pluginConfig.ID + "/" + pluginConfig.ID + ".so"
}

func LoadPlugins(pluginsConfig config.PluginsConfig) PluginData {
	pluginList := make(PluginList, 0, len(pluginsConfig))
	pluginMap := make(PluginMap, len(pluginsConfig))
	for _, pluginConfig := range pluginsConfig {
		plg, err := plugin.Open(GetPluginPath(pluginConfig))
		if err != nil {
			slog.Error("failed to load plugin", "plugin", pluginConfig.ID, "err", err)
			continue
		}

		plgVar, err := plg.Lookup("Plugin")
		if err != nil {
			slog.Error("plugin missing Plugin symbol", "plugin", pluginConfig.ID, "err", err)
			continue
		}

		// Plugins may export their Plugin symbol either as a concrete type that implements
		// UdbPlugin, or as a *UdbPlugin interface value. Handle both conventions.
		var typedPluginVar udb_plugin_library.UdbPlugin
		if pluginPtr, ok := plgVar.(*udb_plugin_library.UdbPlugin); ok {
			typedPluginVar = *pluginPtr
		} else if direct, ok := plgVar.(udb_plugin_library.UdbPlugin); ok {
			typedPluginVar = direct
		} else {
			slog.Error("plugin does not implement UdbPlugin", "plugin", pluginConfig.ID)
			continue
		}
		err = typedPluginVar.Configure(pluginConfig.Config)
		if err != nil {
			slog.Error("failed to configure plugin", "plugin", pluginConfig.ID, "err", err)
			continue
		}
		pluginList = append(pluginList, &typedPluginVar)
		pluginMap[pluginConfig.ID] = &typedPluginVar
	}
	return PluginData{
		List: pluginList,
		ById: pluginMap,
	}
}
