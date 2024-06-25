package config

import (
	"bytes"
	"log"
	"os"
	"text/template"

	"github.com/spf13/viper"
)

func LoadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("toml")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		cfg := GetConfig()
		cfg.PlanetmintAddress = v.GetString("planetmint-address")
		cfg.PlanetmintChainID = v.GetString("planetmint-chain-id")
		cfg.RPCHost = v.GetString("rpc-host")
		cfg.RPCUser = v.GetString("rpc-user")
		cfg.RPCPass = v.GetString("rpc-pass")
		cfg.PlanetmintRPCHost = v.GetString("planetmint-rpc-host")
		cfg.ServicePort = v.GetInt("service-port")
		cfg.ServiceBind = v.GetString("service-bind")
		cfg.AcceptedAsset = v.GetString("accepted-asset")
		cfg.Wallet = v.GetString("wallet")
		cfg.Confirmations = v.GetInt64("confirmations")
		cfg.LogLevel = v.GetString("log-level")
		return
	}
	log.Println("no config file found.")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	err = configTemplate.Execute(&buffer, GetConfig())
	if err != nil {
		return
	}

	err = v.ReadConfig(&buffer)
	if err != nil {
		return
	}
	err = v.SafeWriteConfig()
	if err != nil {
		return
	}

	log.Println("default config file created. please adapt it and restart the application. exiting...")
	os.Exit(0)
	return
}
