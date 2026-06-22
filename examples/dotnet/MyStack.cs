using Kurrent.Gcp;
using Pulumi;
using Pulumi.KurrentCloud;
using Pulumi.Gcp.Compute;
using Network = Pulumi.KurrentCloud.Network;
using NetworkArgs = Pulumi.KurrentCloud.NetworkArgs;

class MyStack : Stack {
    public MyStack() {
        const string gcpRegion  = "europe-west2";
        const string gcpIpRange = "172.30.0.0/16";
        const string kurrentIpRange = "172.22.110.0/24";

        var cloudResources = new CloudResources(gcpRegion, gcpIpRange);

        var kurrentProject = new Project("kurrent-project", new ProjectArgs {Name = "My Kurrent Cloud Project"});

        var kurrentNetwork = new Network(
            "kurrent-network",
            new NetworkArgs {
                Name             = "Test Network",
                Region           = gcpRegion,
                CidrBlock        = kurrentIpRange,
                ProjectId        = kurrentProject.Id,
                ResourceProvider = "gcp"
            }
        );

        var kurrentPeering = new Peering(
            "kurrent-peering",
            new PeeringArgs {
                Name                 = "Test Peering",
                ProjectId            = kurrentProject.Id,
                NetworkId            = kurrentNetwork.Id,
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
                Name               = "kurrent-peering",
                Network            = cloudResources.Network.Id,
                PeerNetwork        = kurrentPeering.ProviderMetadata.Apply(x => x["gcp_network_id"]),
                ExportCustomRoutes = true,
                ImportCustomRoutes = true
            }
        );

        var cluster = new ManagedCluster(
            "myCluster",
            new ManagedClusterArgs {
                Name            = "Test Cluster",
                ProjectId       = kurrentProject.Id,
                NetworkId       = kurrentNetwork.Id,
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