package main

import (
	kurrent "github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi-gcp/sdk/v6/go/gcp/compute"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		gcpConfig := config.New(ctx, "gcp")
		gcpRegion := gcpConfig.Require("region")
		gcpIpRange := "172.30.0.0/16"

		network, err := compute.NewNetwork(ctx, "gcp-network", &compute.NetworkArgs{
			AutoCreateSubnetworks: pulumi.Bool(false),
			Description:           pulumi.String("Test VPC"),
			Name:                  pulumi.String("test-vpc"),
		})
		if err != nil {
			return err
		}

		_, err = compute.NewSubnetwork(ctx, "gpc-subnet", &compute.SubnetworkArgs{
			Description: pulumi.String("test-subnet"),
			IpCidrRange: pulumi.String(gcpIpRange),
			Name:        pulumi.String("test-subnet"),
			Network:     network.Name,
		})
		if err != nil {
			return err
		}

		project, err := kurrent.NewProject(ctx, "sample-project", &kurrent.ProjectArgs{Name: pulumi.String("sample-project")})
		if err != nil {
			return err
		}

		escNetwork, err := kurrent.NewNetwork(ctx, "kurrent-network", &kurrent.NetworkArgs{
			CidrBlock:        pulumi.String("172.22.110.0/24"),
			Name:             pulumi.String("sample-network"),
			ProjectId:        project.ID(),
			Region:           pulumi.String(gcpRegion),
			ResourceProvider: pulumi.String("gcp"),
		})
		if err != nil {
			return err
		}

		escPeering, err := kurrent.NewPeering(ctx, "kurrent-peering", &kurrent.PeeringArgs{
			Name:                 pulumi.String("sample-peering"),
			NetworkId:            escNetwork.ID(),
			PeerAccountId:        network.Project,
			PeerNetworkId:        network.Name,
			PeerNetworkRegion:    pulumi.String(gcpRegion),
			PeerResourceProvider: pulumi.String("gcp"),
			ProjectId:            project.ID(),
			Routes:               pulumi.StringArray{pulumi.String(gcpIpRange)},
		})
		if err != nil {
			return err
		}

		escGcpPeeringId := escPeering.ProviderMetadata.MapIndex(pulumi.String("gcp_network_id"))
		_, err = compute.NewNetworkPeering(ctx, "gcp-peering", &compute.NetworkPeeringArgs{
			ExportCustomRoutes: pulumi.Bool(true),
			ImportCustomRoutes: pulumi.Bool(true),
			Name:               pulumi.String("kurrent-peering"),
			Network:            network.ID(),
			PeerNetwork:        escGcpPeeringId,
		})
		if err != nil {
			return err
		}

		cluster, err := kurrent.NewManagedCluster(ctx, "sample-cluster", &kurrent.ManagedClusterArgs{
			DiskSize:        pulumi.Int(10),
			DiskType:        pulumi.String("ssd"),
			InstanceType:    pulumi.String("F1"),
			Name:            pulumi.String("sample-cluster"),
			NetworkId:       escNetwork.ID(),
			ProjectId:       project.ID(),
			ProjectionLevel: pulumi.String("user"),
			ServerVersion:   pulumi.String("24.10"),
			Topology:        pulumi.String("single-node"),
		})

		ctx.Export("clusterId", cluster.ID())
		return nil
	})
}
