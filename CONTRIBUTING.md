# Contributing to Pulumi

Do you want to hack on Pulumi?  Awesome!  We are so happy to have you.

Please refer to the [main Pulumi repo](https://github.com/pulumi/pulumi/)'s [CONTRIBUTING.md file](
https://github.com/pulumi/pulumi/blob/master/CONTRIBUTING.md) for details on how to do so.

## Prerequisites

- Go 1.24 or later
- Node.js 18 or later
- Python 3.7 or later
- .NET SDK 6.0 or later
- [`pulumictl`](https://github.com/pulumi/pulumictl)

This repository includes a config to setup a [devcontainer](https://containers.dev/), which will
setup an environment with all necessary requirements installed.

## Committing Generated Code

Code generated for Pulumi SDKs should be checked in as part of the pull request containing a
particular change. To generate code after making changes, run `make build` from the root of this
repository.

## Building the Provider

The provider is generated from the Event Store Cloud Terraform provider. To build the Pulumi provider
and generate SDKs:

1. Ensure you're in the development container
2. Run the following command to build all SDKs:

    ```bash
    make build
    ```

This will:

- Build the provider binary
- Generate SDKs for Node.js, Python, Go, and .NET

## Manual Testing

> [!IMPORTANT]
> This will create and destroy **real** Kurrent Cloud resources

To test your changes before committing them:

1. Build the provider using the [steps above](#building-the-provider)
2. Install the provider locally:

    ```bash
    pulumi plugin install resource eventstorecloud --file bin/pulumi-resource-eventstorecloud
    ```

3. Temporarily set the version in the TypeScript SDKs `package.json`

    ```bash
    sed -i 's/${VERSION}/0.0.1-dev/' sdk/nodejs/package.json
    ```

4. Login to Pulumi locally to prevent being prompted to login to the hosted service

    ```bash
    pulumi login --local
    ```

5. Change directory to the test project

    ```bash
    cd test
    ```

6. Install the local version of the TypeScript SDK

    ```bash
    yarn add @eventstore/pulumi-eventstorecloud@file:../sdk/nodejs
    ```

7. Create a new stack

    ```bash
    PULUMI_CONFIG_PASSPHRASE=very-secure-passphrase pulumi stack init -s dev
    ```

8. Set config values

    ```bash
    pulumi config set eventstorecloud:organizationId <ORG_ID>
    pulumi config set eventstorecloud:token <ACCESS_TOKEN> --secret
    ```

9. Create the stack

    ```bash
    pulumi up -y
    ```

10. Once you're done, clean up resources and revert the change to `package.json`

    ```bash
    PULUMI_CONFIG_PASSPHRASE=very-secure-passphrase pulumi destroy -y &&
    pulumi stack rm dev -y &&
    cd .. && sed -i 's/0.0.1-dev/${VERSION}/' sdk/nodejs/package.json
    ```

## Publishing a New Release

Once all changes have been merged to the `main` branch and you're ready to cut a new release, all
that needs to be done is to create a new tag using semver format, e.g. v0.2.21.

Once the tag has been pushed to the Github repo, an action will be triggered. This will create new a
new Github release with the binary artifacts, as well as publish the release to NPM and Nuget.

The provider docs in the Pulumi Package Registry will be updated automatically sometime later.
