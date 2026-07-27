# Pulumi provider for Kurrent Cloud

The Kurrent Cloud provider allows you to manage resources in [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud).

## Installation

This package is available in many languages in the standard packaging formats.

### Configure the provider

The following configuration points are available for the `kurrentcloud` provider:

- `kurrentcloud:organizationId` - the organization ID for an existing organization in Kurrent Cloud
- `kurrentcloud:clientId` / `kurrentcloud:clientSecret` - Service Account credentials (recommended for automation); when both are set they take priority over `token`
- `kurrentcloud:token` - a valid refresh token for an Kurrent Cloud account with admin access to the organization

### Install SDK


