# Pulumi provider for Kurrent Cloud

The Kurrent Cloud provider allows you to manage resources in [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud).

## Installation

This package is available in many languages in the standard packaging formats.

### Configure the provider

The following configuration points are available for the `kurrentcloud` provider:

- `kurrentcloud:organizationId` - the organization ID for an existing organization in Kurrent Cloud
- `kurrentcloud:token` - a valid refresh token for an Kurrent Cloud account with admin access to the organization

### Install SDK


Add the NuGet package `Pulumi.KurrentCloud` to your Pulumi project, which uses the .NET Pulumi SDK.
### Get the plugin

For projects that use .NET and Go Pulumi SDK you have to install the provider before trying to update the stack.

Use the following command to add the plugin to your environment:

```
pulumi plugin install resource kurrentcloud [version] \
  --server https://github.com/kurrent-io/pulumi-eventstorecloud/releases/download/[version]
```

Example:

```
pulumi plugin install resource kurrentcloud v1.0.0 \
  --server https://github.com/kurrent-io/pulumi-eventstorecloud/releases/download/v1.0.0
```
