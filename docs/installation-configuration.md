---
title: Kurrent Cloud Installation & Configuration
meta_desc: How to set up credentials to use the Kurrent Cloud provider for Pulumi.
layout: package
---

## Installation

The Kurrent Cloud provider is available as a package in all Pulumi languages:

- JavaScript/TypeScript: [`@kurrent-io/pulumi-kurrentcloud`](https://www.npmjs.com/package/@kurrent-io/pulumi-kurrentcloud)
- Go: [`github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud`](https://github.com/kurrent-io/pulumi-eventstorecloud)
- .NET: [`Pulumi.KurrentCloud`](https://www.nuget.org/packages/Pulumi.KurrentCloud)

## Setup

### Download the provider

For projects that use .NET and Go Pulumi SDK you have to install the provider before trying to update the stack.

Use the following command to add the plugin to your environment:

```
pulumi plugin install resource kurrentcloud --server github://api.github.com/kurrent-io
```

### Configure the provider

The Pulumi provider needs credentials to authenticate requests from your computer to Kurrent Cloud. Your credentials are never sent
to pulumi.com. The provider needs to be configured with Kurrent Cloud credentials before it can be used to create resources.

First, you need an access token for your user, which you can obtain from the Kurrent Cloud console.

Then, go to the list of organizations you have access to in the Kurrent Cloud console, choose the organization that you will be provisioning resources for, and find the organization id in the settings.

- `<YOUR_ACCESS_TOKEN>`: your access token
- `<YOUR_ORGANIZATION_ID>`: the Kurrent Cloud organization id

Once the credentials are obtained, there are two ways to communicate your authorization tokens to Pulumi:

1. Set the environment variables `ESC_TOKEN` and `ESC_ORG_ID`:

    ```bash
    $ export ESC_TOKEN=<YOUR_ACCESS_TOKEN>
    $ export ESC_ORG_ID=<YOUR_ORGANIZATION_ID>
    ```

2. Set them using configuration, if you prefer that they be stored alongside your Pulumi stack for easy multi-user access:

    ```bash
    pulumi config set kurrentcloud:token <YOUR_ACCESS_TOKEN> --secret
    pulumi config set kurrentcloud:organizationId <YOUR_ORGANIZATION_ID> --secret
    ```

{{% notes "info" %}}
Required options can be omitted if you configure them using environment variables.
{{% /notes %}}

| Option           | Required/Optional | Description                                                                                       |
| ---------------- | ----------------- | ------------------------------------------------------------------------------------------------- |
| `token`          | Required          | Access token. You can retrieve this from the ‘Access Tokens’ section of the Kurrent Cloud console. |
| `organizationId` | Required          | The organization id. You can find it in the organization settings page of the Kurrent Cloud console. |
| `url`            | Optional          | The URL of the Kurrent Cloud API. This defaults to the public cloud instance of Kurrent Cloud.    |
| `tokenStore`     | Optional          | The location on the local filesystem of the token cache. This is shared with the Kurrent CLI.     |
