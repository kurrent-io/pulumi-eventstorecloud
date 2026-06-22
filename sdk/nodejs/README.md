# Pulumi provider for Kurrent Cloud

The Kurrent Cloud provider allows you to manage resources in [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud).

## Installation

This package is available in many languages in the standard packaging formats.

### Configure the provider

The following configuration points are available for the `kurrentcloud` provider:

- `kurrentcloud:organizationId` - the organization ID for an existing organization in Kurrent Cloud
- `kurrentcloud:token` - a valid refresh token for an Kurrent Cloud account with admin access to the organization

### Install SDK


Install the NodeJS SDK using either `npm`:

```bash
$ npm install @kurrent-io/pulumi-kurrentcloud
```

or `yarn`:

```bash
$ yarn add @kurrent-io/pulumi-kurrentcloud
```
