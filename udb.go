package main

import (
	"fmt"

	"github.com/benwiebe/udb-core/internal/config"
)

func main() {
	/**** Load App Config ****/
	configLoader := config.NewDefaultConfigLoader()
	if err := configLoader.Load(); err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		return
	}
	_ = configLoader.GetConfig()
	fmt.Println("Config loaded")

}
