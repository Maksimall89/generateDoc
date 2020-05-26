package main

type Grafana struct {
	PanelsID    []string `json:"panelsId"`              //Required field, panel id as int
	Hostname    string   `json:"hostname,omitempty"`    //Optional field if assigned at configs, example "http://127.0.0.1:3000"
	Dashboard   string   `json:"dashboard,omitempty"`   //Optional field if assigned at configs, path to dashboard, example "/d/XZsIP9qik/test"
	Width       int      `json:"width,omitempty"`       //Optional field if assigned at configs
	Height      int      `json:"height,omitempty"`      //Optional field if assigned at configs
	Description string   `json:"description,omitempty"` //Optional field if assigned at configs
	DescStyle   string   `json:"descStyle,omitempty"`   //Optional field if assigned at configs
}

type Text struct {
	Content string `json:"content"`         //Required field
	Style   string `json:"style,omitempty"` //Optional field if assigned at configs
}

type ConfigFile struct {
	Times    string
	TimeZone [5]string
	Projects []struct {
		Name    string `json:"name"`
		Key     string `json:"key"`
		Configs struct {
			Grafana  Grafana `json:"grafana"`
			TextJson Text    `json:"text"`
		} `json:"configs"`
		Instructions []struct {
			Grafana  Grafana `json:"grafana"`
			TextJson Text    `json:"text"`
		} `json:"instructions"`
	} `json:"projects"`
}

type ConfigWeb struct {
	TimeTo   string
	TimeFrom string
	Project  string
	TimeZone string
}
