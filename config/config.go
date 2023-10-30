package config

import "sync"

const DefaultConfigTemplate = `
PLANETMINT_GO={{ .Planetmint }}
PLANETMINT_ADDRESS={{ .PlanetmintAddress }}
RPC_HOST={{ .RPCHost }}
RPC_USER={{ .RPCUser }}
RPC_PASS={{ .RPCPass }}
PM_RPC_HOST={{ .PlanetmintRPCHost }}
SERVICE_BIND={{ .ServiceBind }}
SERVICE_PORT={{ .ServicePort }}
REISSUANCE_ASSET={{ .ReissuanceAsset }}
`

type Config struct {
	Planetmint        string `mapstructure:"planetmint"`
	PlanetmintAddress string `mapstructure:"planetmint-address"`
	RPCHost           string `mapstructure:"rpc-host"`
	RPCUser           string `mapstructure:"rpc-user"`
	RPCPass           string `mapstructure:"rpc-pass"`
	PlanetmintRPCHost string `mapstructure:"planetmint-rpc-host"`
	ServicePort       int    `mapstructure:"service-port"`
	ServiceHost       string `mapstructure:"service-host"`
	ReissuanceAsset   string `mapstructure:"reissuance-asset"`
}

// global singleton
var (
	config     *Config
	initConfig sync.Once
)

// DefaultConfig returns RDDL-2-PLMNT default config
func DefaultConfig() *Config {
	return &Config{
		Planetmint:        "planetmint-god",
		PlanetmintAddress: "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5",
		RPCHost:           "planetmint-go-testnet-3.rddl.io:18884",
		RPCUser:           "user",
		RPCPass:           "password",
		PlanetmintRPCHost: "127.0.0.1:9090",
		ServicePort:       8080,
		ServiceHost:       "localhost",
		ReissuanceAsset:   "7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9",
	}
}

func GetConfig() *Config {
	initConfig.Do(func() {
		config = DefaultConfig()
	})
	return config
}
