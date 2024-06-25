module github.com/rddl-network/rddl-2-plmnt-service/client

go 1.21.5

require github.com/rddl-network/rddl-2-plmnt-service v0.2.0

require gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/testify v1.9.0
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/rddl-network/rddl-2-plmnt-service => ../
