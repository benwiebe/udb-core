package main

// plugin_imports.go — add blank imports here to include plugins in the binary.
//
// Each import triggers the plugin's init() function, which calls
// udb_plugin_library.Register() to make it available at runtime.
//
// To build with specific plugins:
//
//	make build PLUGINS="github.com/benwiebe/udb-plugin-nhl@v1.0.0"
//
// udb-builder manages this file automatically when building a custom binary.
// For local development, add blank imports manually and ensure the modules
// are available in your go.work workspace or go.mod.
