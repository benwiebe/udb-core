package plugins

import (
	"fmt"
	"plugin"

	"github.com/benwiebe/udb-core/internal/config"
	udb_plugin_library "github.com/benwiebe/udb-plugin-library"
)

type PluginList = []*udb_plugin_library.UdbPlugin

func GetPluginPath(pluginConfig config.PluginsConfig) string {
	if pluginConfig.Path != "" {
		return pluginConfig.Path
	}
	return "./plugins/" + pluginConfig.ID + "/" + pluginConfig.ID + ".so"
}

func LoadPlugins(pluginsConfig []config.PluginsConfig) PluginList {
	pluginList := make(PluginList, 0, len(pluginsConfig))
	for _, pluginConfig := range pluginsConfig {
		// We use plugin.Open to load the plugin by path
		plg, err := plugin.Open(GetPluginPath(pluginConfig))
		if err != nil {
			fmt.Printf("Error loading plugin %s: %v\n", pluginConfig.ID, err)
			continue
		}

		plgVar, err := plg.Lookup("Plugin")
		if err != nil {
			fmt.Printf("Error finding 'Plugin' symbol in plugin %s: %v\n", pluginConfig.ID, err)
			continue
		}

		typedPluginVar, ok := plgVar.(udb_plugin_library.UdbPlugin)
		if !ok {
			fmt.Printf("Error: plugin %s does not implement UdbPlugin interface\n", pluginConfig.ID)
			continue
		}
		err = typedPluginVar.Configure(pluginConfig.Config)
		if err != nil {
			fmt.Printf("Error configuring plugin %s: %v\n", pluginConfig.ID, err)
			continue
		}
		pluginList = append(pluginList, &typedPluginVar)
	}
	return pluginList
}
