# rddl-2-plmnt-service
This service receives `POST requests` on `http(s)://localhost:8080/mint/:txhash`  and checks for the corresponding Liquid transaction and if the amount has already been minted on Planetmint. The request body must contain the Planetmint beneficiary address.

**Curl example:**
```
curl -X POST -H "Content-Type: application/json" -d '{"beneficiary": "plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5"}' localhost:8080/mint/eb738e1db87406c03922246370ed53b6b873f81ac37fd76e86e31e121018a8e3
```

## Execution
The service can be executed via the following go command without having it previously built:
```
go run cmd/rddl-2-plmnt-service main.go
```

## Configuration
The service needs to be configured via the ```./app.env``` file or environment variables. The defaults are
```
PLANETMINT_GO=planetmint-god
PLANETMINT_ADDRESS=plmnt15xuq0yfxtd70l7jzr5hg722sxzcqqdcr8ptpl5
PLANETMINT_KEYRING= # optional
RPC_HOST=testnet-explorer.rddl.io:18884
RPC_USER=
RPC_PASS=
SERVICE_PORT=8080
SERVICE_BIND=localhost
```
A sample ```./app.env``` file is at ```./app.env.template```

**Important:** The `PLANETMINT_ADDRESS` needs to be the `MintAddress` configured on Planetmint in order to pass the `AnteHandler` check.