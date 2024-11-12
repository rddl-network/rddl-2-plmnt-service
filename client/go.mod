module github.com/rddl-network/rddl-2-plmnt-service/client

go 1.22

toolchain go1.22.8

require github.com/rddl-network/rddl-2-plmnt-service v0.2.0

require gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rddl-network/rddl-2-plmnt-service => ../
