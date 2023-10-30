# rddl-2-plmnt-service
This service receives `POST requests` on `http(s)://localhost:8080/mint/:txhash`  and checks for the corresponding Liquid transaction and if the amount has already been minted on Planetmint. The request body must contain the Planetmint beneficiary address.

**Curl example:**
```
curl -X POST -H "Content-Type: application/json" -d '{"beneficiary": "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5"}' localhost:8080/mint/eb738e1db87406c03922246370ed53b6b873f81ac37fd76e86e31e121018a8e3
```

## Execution
The service can be executed via the following go command without having it previously built:
```
go run cmd/rddl-2-plmnt-service/main.go
```

## Configuration
The service needs to be configured via the ```./app.toml``` file or environment variables. The defaults are
```
planetmint = "planetmint-god"
planetmint-address = "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5"
rpc-host = "planetmint-go-testnet-3.rddl.io:18884"
rpc-user = "user"
rpc-pass = "password"
planetmint-rpc-host = "127.0.0.1:9090"
service-port = 8080
service-host = "localhost"
reissuance-asset = "7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9"
```
The defaults can be found at ```./config/config.go```.

**Important:** The `planetmint-address` needs to be the `MintAddress` configured on Planetmint in order to pass the `AnteHandler` check.