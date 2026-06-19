module kurrent-sample-go

go 1.24

require (
	github.com/EventStore/pulumi-eventstorecloud/sdk v1.0.0
	github.com/pulumi/pulumi-gcp/sdk/v6 v6.12.0
	github.com/pulumi/pulumi/sdk/v3 v3.24.1
)

// Source-tree example: resolve the SDK from this repository, which contains the
// renamed sdk/go/kurrentcloud package, so the example builds against local changes.
replace github.com/EventStore/pulumi-eventstorecloud/sdk => ../../sdk
