import * as pulumi from "@pulumi/pulumi";
import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";
import * as random from "@pulumi/random";

const name = new random.RandomPet("project-name", {});

const project = new kurrent.Project("project", {
    name: pulumi.interpolate`test-project-${name.id}`,
});

const network = new kurrent.Network("network", {
    name: pulumi.interpolate`network-${name.id}`,
    projectId: project.id,
    resourceProvider: "aws",
    region: "us-west-2",
    cidrBlock: "172.21.0.0/16",
});

const cluster = new kurrent.ManagedCluster("server", {
    name: pulumi.interpolate`cluster-${name.id}`,
    projectId: project.id,
    networkId: network.id,
    topology: "single-node",
    instanceType: "F1",
    diskSize: 10,
    diskType: "gp3",
    diskIops: 3000,
    diskThroughput: 125,
    serverVersion: "24.10",
});

// Read-only replica set attached to the managed cluster (added with kurrentcloud v2 parity).
const replica = new kurrent.ManagedClusterReplicaset("replica", {
    projectId: project.id,
    clusterId: cluster.id,
    replicaCount: 1,
});

export let clusterDnsName = cluster.dnsName;
