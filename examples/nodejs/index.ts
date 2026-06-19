import * as pulumi from "@pulumi/pulumi";
import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";
import * as awsx from "@pulumi/awsx";
import * as aws from "@pulumi/aws";

const awsConfig = new pulumi.Config("aws");

const vpc = new awsx.ec2.Vpc("example", {
    cidrBlock: "172.250.0.0/24",
    numberOfAvailabilityZones: 3,
});

const project = new kurrent.Project("sample-project", {
    name: "Improved Chicken Window",
});

const network = new kurrent.Network("sample-network", {
    name: "Chicken Window Net",
    projectId: project.id,
    resourceProvider: "aws",
    region: awsConfig.require("region"),
    cidrBlock: "172.21.0.0/16",
});

const peering = new kurrent.Peering("sample-peering", {
    name: "Sample Peering",
    projectId: project.id,
    networkId: network.id,
    peerResourceProvider: network.resourceProvider,
    peerNetworkRegion: network.region,
    peerAccountId: vpc.vpc.ownerId,
    peerNetworkId: vpc.vpcId,
    routes: [vpc.vpc.cidrBlock],
});

const peer = new aws.ec2.VpcPeeringConnectionAccepter("sample-accept", {
    vpcPeeringConnectionId: peering.providerMetadata["aws_peering_link_id"],
    autoAccept: true,
    tags: {
        Side: "Accepter",
        Source: "Kurrent",
    },
});

const route = new aws.ec2.Route("route", {
    vpcPeeringConnectionId: peer.id,
    routeTableId: vpc.vpc.mainRouteTableId,
    destinationCidrBlock: network.cidrBlock,
});

const cluster = new kurrent.ManagedCluster("wings", {
    // name: "Wings Cluster",
    projectId: project.id,
    networkId: network.id,
    topology: "single-node",
    instanceType: "F1",
    diskSize: 16,
    diskType: "gp3",
    diskIops: 3000,
    diskThroughput: 125,
    serverVersion: "24.10",
});

export let clusterDnsName = cluster.dnsName;
