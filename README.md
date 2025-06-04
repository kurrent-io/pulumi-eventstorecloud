# Pulumi provider for Kurrent Cloud (formerly Event Store Cloud)

The `eventstorecloud` provider allows you to manage resources in [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud).

## Installing

This package is available in many languages in the standard packaging formats.

### Get the plugin

For projects that use .NET and Go Pulumi SDK you have to install the provider before trying to update the stack.

Use the following command to add the plugin to your environment:

```bash
pulumi plugin install resource eventstorecloud --server github://api.github.com/kurrent-io
```

### Configuration

The following configuration options are required for the `eventstorecloud` provider:

- `eventstorecloud:organizationId` - the organization ID for an existing organization in Event Store Cloud
- `eventstorecloud:token` - a valid refresh token for an Event Store Cloud account with admin access to the organization

Alternatively, these values can be set via the `ESC_ORG_ID` and `ESC_TOKEN` environment variables.

### Node.js (Java/TypeScript)

To use from JavaScript or TypeScript in Node.js, install using either `npm`:

```bash
npm install @eventstore/pulumi-eventstorecloud
```

or `yarn`:

```bash
yarn add @eventstore/pulumi-eventstorecloud
```

### Go

To use from Go, use `go get` to grab the latest version of the library

```bash
go get github.com/EventStore/pulumi-eventstorecloud/sdk/go/eventstorecloud
```

### .NET

To use from .NET, install using `dotnet add package`:

```bash
dotnet add package Pulumi.EventStoreCloud
```

### Python

\[WIP]

## Reference

For detailed reference documentation, please visit [the Pulumi registry](https://www.pulumi.com/registry/packages/eventstorecloud/).
