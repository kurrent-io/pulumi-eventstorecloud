using esc.Gcp;
using Pulumi;
using Pulumi.KurrentCloud;
using Pulumi.Gcp.Compute;
using Network = Pulumi.KurrentCloud.Network;
using NetworkArgs = Pulumi.KurrentCloud.NetworkArgs;

class MyStack : Stack {
    public MyStack() {
        const string gcpRegion  = "europe-west2";
        const string gcpIpRange = "172.30.0.0/16";
        const string escIpRange = "172.22.110.0/24";

        var cloudResources = new CloudResources(gcpRegion, gcpIpRange);

        var escProject = new Project("esc-project", new ProjectArgs {Name = "My ES Cloud Project"});

        var escNetwork = new Network(
            "esc-network",
            new NetworkArgs {
                Name             = "Test Network",
                Region           = gcpRegion,
                CidrBlock        = escIpRange,
                ProjectId        = escProject.Id,
                ResourceProvider = "gcp"
            }
        );

        var escPeering = new Peering(
            "esc-peering",
            new PeeringArgs {
                Name                 = "Test Peering",
                ProjectId            = escProject.Id,
                NetworkId            = escNetwork.Id,
                PeerAccountId        = cloudResources.Network.Project,
                PeerNetworkId        = cloudResources.Network.Name,
                PeerNetworkRegion    = cloudResources.Subnet.Region,
                PeerResourceProvider = "gcp",
                Routes               = new[] {gcpIpRange}
            }
        );

        var gcpPeering = new NetworkPeering(
            "gcp-peering",
            new NetworkPeeringArgs {
                Name               = "esc-peering",
                Network            = cloudResources.Network.Id,
                PeerNetwork        = escPeering.ProviderMetadata.Apply(x => x["gcp_network_id"]),
                ExportCustomRoutes = true,
                ImportCustomRoutes = true
            }
        );

        var cluster = new ManagedCluster(
            "myCluster",
            new ManagedClusterArgs {
                Name            = "Test Cluster",
                ProjectId       = escProject.Id,
                NetworkId       = escNetwork.Id,
                Topology        = "single-node",
                InstanceType    = "F1",
                DiskSize        = 10,
                DiskType        = "ssd",
                ServerVersion   = "24.10",
                ProjectionLevel = "user"
            }
        );

        ClusterId        = cluster.Id;
        ConnectionString = cluster.Id.Apply(id => $"esdb+discover://{id}.mesdb.eventstore.cloud:2113");
    }

    [Output] public Output<string> ClusterId { get; set; }

    [Output] public Output<string> ConnectionString { get; set; }
}