# rddl-2-plmnt-service
This service receives `POST requests` on `http(s)://localhost:8080/mint` with a `Conversion` object and checks for the corresponding Liquid transaction, verifies the provided signature, address descriptor and if the amount has already been minted on Planetmint. The request body must look like the following:
```json
{
    "conversion": {
        "beneficiary": "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx",
        "liquid-tx-hash": "b356413f906468a3220f403c350d01a5880dbd1417f3ff294a4a2ff62faf0839",
        "descriptor": "wpkh([6a00c946/0'/0'/501']02e24c96e967524fb2ad3b3e3c29c275e05934b12f420b7871443143d05ffe11c8)#8ktzldqn"
    },
    "signature": "ICucxAHOsf1kanl9UAjxMXemLmnP0deHWwyqdav68e8XCknJeaNBPFl9t7h52Ny1/XNgiQFu8XzrGLM8qahSy38="
}
```

Where the `beneficiary` is the receiving address on Planetmint, `liquid-tx-hash` is the liquid tx that transferred RDDL token to the `rddl2plmnt` wallet, `descriptor` is the function that derives the sending address of the liquid tx and the signature is signed with the private key that sent the liquid tx.


**Curl example:**
```
curl -X POST -H "Content-Type: application/json" -d '{"conversion": {"beneficiary": "plmnt1w5dww335zhh98pzv783hqre355ck3u4w4hjxcx","liquid-tx-hash": "b356413f906468a3220f403c350d01a5880dbd1417f3ff294a4a2ff62faf0839","descriptor": "wpkh([6a00c946/0'/0'/501']02e24c96e967524fb2ad3b3e3c29c275e05934b12f420b7871443143d05ffe11c8)#8ktzldqn"},"signature": "ICucxAHOsf1kanl9UAjxMXemLmnP0deHWwyqdav68e8XCknJeaNBPFl9t7h52Ny1/XNgiQFu8XzrGLM8qahSy38="}' localhost:8080/mint
```

## Execution
The service can be executed via the following go command without having it previously built:
```
go run cmd/rddl-2-plmnt-service/main.go
```

## Configuration
The service needs to be configured via the ```./app.toml``` file or environment variables. The defaults are
```
planetmint-address = "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5"
planetmint-chain-id = "planetmint-testnet-1"
rpc-host = "planetmint-go-testnet-3.rddl.io:18884"
rpc-user = "user"
rpc-pass = "password"
planetmint-rpc-host = "127.0.0.1:9090"
service-port = 8080
service-host = "localhost"
accepted-asset = "7add40beb27df701e02ee85089c5bc0021bc813823fedb5f1dcb5debda7f3da9"
wallet = "rddl2plmnt"
```
The defaults can be found at ```./config/config.go```.

**Important:** The `planetmint-address` needs to be the `MintAddress` configured on Planetmint in order to pass the `AnteHandler` check.