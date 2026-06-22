using Pulumi.Gcp.Compute;

namespace Kurrent.Gcp {
    public class CloudResources {
        public CloudResources(string region, string ipRange) {
            Network = new Network(
                "gcp-network",
                new NetworkArgs {
                    Description           = "Test VPC",
                    Name                  = "test-vpc",
                    AutoCreateSubnetworks = false
                }
            );

            Subnet = new Subnetwork(
                "gcp-subnet",
                new SubnetworkArgs {
                    Name        = "test-subnet",
                    Network     = Network.Name,
                    Region      = region,
                    IpCidrRange = ipRange
                }
            );
        }

        public Network    Network { get; }
        public Subnetwork Subnet  { get; }
    }
}