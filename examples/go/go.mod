module esc-sample-go

go 1.14

require (
	github.com/EventStore/pulumi-eventstorecloud/sdk v0.2.3
	github.com/pulumi/pulumi-gcp/sdk/v6 v6.12.0
	github.com/pulumi/pulumi/sdk/v3 v3.24.1
)

// Source-tree example: resolve the SDK from this repository, which contains the
// renamed sdk/go/kurrentcloud package (the pinned version above predates it).
replace github.com/EventStore/pulumi-eventstorecloud/sdk => ../../sdk
