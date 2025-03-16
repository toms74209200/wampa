module wampa/tests/acceptance

go 1.23.4

require (
	github.com/cucumber/godog v0.12.6
	github.com/toms74209200/wampa v0.0.0-20250315135142-3f337becc2b3
)

require (
	github.com/cucumber/gherkin-go/v19 v19.0.3 // indirect
	github.com/cucumber/messages-go/v16 v16.0.1 // indirect
	github.com/gofrs/uuid v4.2.0+incompatible // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-memdb v1.3.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/toms74209200/wampa/pkg => ../../pkg
