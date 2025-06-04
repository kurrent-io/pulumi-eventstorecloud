import * as pulumi from "@pulumi/pulumi";
import * as esc from "@eventstore/pulumi-eventstorecloud";
import * as random from "@pulumi/random";

const name = new random.RandomPet("project-name", {});

const project = new esc.Project("project", {
    name: pulumi.interpolate`test-project-${name.id}`,
});

const network = new esc.Network("network", {
    name: pulumi.interpolate`network-${name.id}`,
    projectId: project.id,
    resourceProvider: "aws",
    region: "us-west-2",
    cidrBlock: "172.21.0.0/16",
});

const cluster = new esc.ManagedCluster("server", {
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

export let clusterDnsName = cluster.dnsName;
