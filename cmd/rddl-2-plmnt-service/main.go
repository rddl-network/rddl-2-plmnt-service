package main

import (
	"bytes"
	"log"
	"os"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/planetmint/planetmint-go/app"
	"github.com/planetmint/planetmint-go/lib"
	"github.com/rddl-network/rddl-2-plmnt-service/config"
	"github.com/rddl-network/rddl-2-plmnt-service/service"
)

var (
	planetmintAddress string
	rpcHost           string
	rpcUser           string
	rpcPass           string
	pmRPCHost         string
	acceptedAsset     string
	wallet            string
)

func loadConfig(path string) (v *viper.Viper, err error) {
	v = viper.New()
	v.AddConfigPath(path)
	v.SetConfigName("app")
	v.SetConfigType("toml")

	v.AutomaticEnv()

	err = v.ReadInConfig()
	if err == nil {
		return
	}
	log.Println("no config file found.")

	tmpl := template.New("appConfigFileTemplate")
	configTemplate, err := tmpl.Parse(config.DefaultConfigTemplate)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	err = configTemplate.Execute(&buffer, config.GetConfig())
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

var libConfig *lib.Config

func main() {
	config, err := loadConfig("./")
	if err != nil {
		log.Fatalf("fatal error loading config file: %s", err)
	}

	planetmintAddress = config.GetString("planetmint-address")
	if err != nil || planetmintAddress == "" {
		panic("Could not read configuration")
	}

	rpcHost = config.GetString("rpc-host")
	rpcUser = config.GetString("rpc-user")
	rpcPass = config.GetString("rpc-pass")
	pmRPCHost = config.GetString("planetmint-rpc-host")
	if rpcHost == "" || rpcUser == "" || rpcPass == "" || pmRPCHost == "" {
		panic("Could not read configuration")
	}

	encodingConfig := app.MakeEncodingConfig()

	libConfig = lib.GetConfig()
	libConfig.SetEncodingConfig(encodingConfig)

	planetmintChainID := config.GetString("planetmint-chain-id")
	if planetmintChainID == "" {
		log.Fatalf("chain id must not be empty")
	}
	libConfig.SetChainID(planetmintChainID)

	acceptedAsset = config.GetString("accepted-asset")
	wallet = config.GetString("wallet")
	if acceptedAsset == "" || wallet == "" {
		panic("Could not read configuration")
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	pmClient := service.NewPlanetmintClient()
	eClient := service.NewElementsClient()
	service := service.NewR2PService(router, pmClient, eClient)

	service.Run(config)
}
