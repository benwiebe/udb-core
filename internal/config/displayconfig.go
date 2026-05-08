package config

type DisplayConfig struct {
	Type        string `json:"type,omitempty"`
	Height      int    `json:"height"`
	Width       int    `json:"width"`
	Brightness  int    `json:"brightness,omitempty"`
	GpioMapping string `json:"gpio_mapping,omitempty"`
	Scale       int    `json:"scale,omitempty"`
}
