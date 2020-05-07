// https://github.com/devilsray/golang-viper-config-example/blob/master/config/server.go

package config

// Configuration ...
type Configuration struct {
	API APIConfiguration
}

// APIConfiguration contains the token
type APIConfiguration struct {
	Token string
}
