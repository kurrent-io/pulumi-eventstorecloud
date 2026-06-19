# Pulumi provider for Kurrent Cloud (formerly Event Store Cloud)

The `kurrentcloud` provider lets you manage [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud)
resources with Pulumi in TypeScript/JavaScript, Python, Go, and .NET. It is generated from the
[Kurrent Cloud Terraform provider](https://github.com/kurrent-io/terraform-provider-kurrentcloud)
via the Pulumi Terraform Bridge, so its resources track the Terraform provider.

> **Renamed from `eventstorecloud`.** Existing stacks keep working: every resource declares a Pulumi
> alias to its historical `eventstorecloud:index:*` type, so running `pulumi up` after upgrading
> refreshes onto the new `kurrentcloud:*` tokens **without replacing** any cloud resources.
> See [MIGRATION.md](./MIGRATION.md).

## Installing

### Get the plugin

For projects that use the .NET and Go Pulumi SDKs you must install the provider plugin before
updating the stack:

```bash
pulumi plugin install resource kurrentcloud --server github://api.github.com/kurrent-io
```

### Node.js (TypeScript/JavaScript)

```bash
npm install @kurrent-io/pulumi-kurrentcloud
# or
yarn add @kurrent-io/pulumi-kurrentcloud
```

### Python

```bash
pip install pulumi_kurrentcloud
```

### Go

```bash
go get github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud
```

### .NET

```bash
dotnet add package Pulumi.KurrentCloud
```

## Configuration

The following configuration options are required for the `kurrentcloud` provider:

- `kurrentcloud:organizationId` - the organization ID for an existing organization in Kurrent Cloud
- `kurrentcloud:token` - a valid refresh token for a Kurrent Cloud account with admin access to the organization

Alternatively, these values can be set via the `ESC_ORG_ID` and `ESC_TOKEN` environment variables.

## Resources

This provider mirrors the [Kurrent Cloud Terraform provider](https://github.com/kurrent-io/terraform-provider-kurrentcloud):

- `Project`, `Network`, `Peering`, `Acl`
- `ManagedCluster` and `ManagedClusterReplicaset` (read-only replica sets attached to a managed cluster)
- `ScheduledBackup`
- `Integration`, `AWSCloudWatchLogsIntegration`, `AWSCloudWatchMetricsIntegration`
- data sources: `getProject`, `getNetwork`

### Example (TypeScript): a managed cluster with a read-only replica set

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";

const project = new kurrent.Project("project", { name: "my-project" });

const network = new kurrent.Network("network", {
    name: "my-network",
    projectId: project.id,
    resourceProvider: "aws",
    region: "us-west-2",
    cidrBlock: "172.21.0.0/16",
});

const cluster = new kurrent.ManagedCluster("cluster", {
    name: "my-cluster",
    projectId: project.id,
    networkId: network.id,
    topology: "single-node",
    instanceType: "F1",
    diskSize: 24,
    diskType: "gp3",
    diskIops: 3000,
    diskThroughput: 125,
    serverVersion: "24.10",
    projectionLevel: "off",
});

// New in this provider: a read-only replica set attached to the managed cluster.
const replica = new kurrent.ManagedClusterReplicaset("replica", {
    projectId: project.id,
    clusterId: cluster.id,
    replicaCount: 1,
});

export const clusterDnsName = cluster.dnsName;
```

## Testing

End-to-end tests exercise every resource and data source against a real Kurrent Cloud
organization (via the Pulumi Automation API), alongside an offline guard that verifies the
`eventstorecloud`→`kurrentcloud` migration aliases. See [TESTING.md](./TESTING.md) for the
harness, the required environment variables (and where their values come from), and how to run
it locally and in CI.

## Reference

For detailed reference documentation, please visit [the Pulumi registry](https://www.pulumi.com/registry/packages/kurrentcloud/).
