package config

import "sync"

const DefaultConfigTemplate = `
planetmint-address="{{ .PlanetmintAddress }}"
planetmint-chain-id="{{ .PlanetmintChainID }}"
rpc-host="{{ .RPCHost }}"
rpc-user="{{ .RPCUser }}"
rpc-pass="{{ .RPCPass }}"
planetmint-rpc-host="{{ .PlanetmintRPCHost }}"
service-bind="{{ .ServiceBind }}"
service-port={{ .ServicePort }}
accepted-asset="{{ .AcceptedAsset }}"
wallet="{{ .Wallet }}"
`

type Config struct {
	PlanetmintAddress string `mapstructure:"planetmint-address"`
	PlanetmintChainID string `mapstructure:"planetmint-chain-id"`
	RPCHost           string `mapstructure:"rpc-host"`
	RPCUser           string `mapstructure:"rpc-user"`
	RPCPass           string `mapstructure:"rpc-pass"`
	PlanetmintRPCHost string `mapstructure:"planetmint-rpc-host"`
	ServicePort       int    `mapstructure:"service-port"`
	ServiceBind       string `mapstructure:"service-bind"`
	AcceptedAsset     string `mapstructure:"accepted-asset"`
	Wallet            string `mapstructure:"wallet"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns RDDL-2-PLMNT default config
func DefaultConfig() *Config {
	return &Config{
		PlanetmintAddress: "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5",
		PlanetmintChainID: "planetmint-testnet-1",
		RPCHost:           "planetmint-go-testnet-3.rddl.io:18884",
		RPCUser:           "user",
		RPCPass:           "password",
		PlanetmintRPCHost: "127.0.0.1:9090",
		ServicePort:       8080,
		ServiceBind:       "localhost",
		AcceptedAsset:     "7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9",
		Wallet:            "rddl2plmnt",
	}
}

func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
