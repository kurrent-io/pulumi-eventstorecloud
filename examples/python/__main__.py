import pulumi
import pulumi_kurrentcloud as kurrent

project = kurrent.Project("sample-project", name="My Kurrent Cloud Project")

network = kurrent.Network(
    "sample-network",
    name="sample-network",
    project_id=project.id,
    resource_provider="aws",
    region="us-west-2",
    cidr_block="172.21.0.0/16",
)

cluster = kurrent.ManagedCluster(
    "sample-cluster",
    name="sample-cluster",
    project_id=project.id,
    network_id=network.id,
    topology="single-node",
    instance_type="F1",
    disk_size=16,
    disk_type="gp3",
    disk_iops=3000,
    disk_throughput=125,
    server_version="24.10",
)

# Read-only replica set attached to the managed cluster (new in the kurrentcloud provider).
replica = kurrent.ManagedClusterReplicaset(
    "sample-replica",
    project_id=project.id,
    cluster_id=cluster.id,
    replica_count=1,
)

pulumi.export("cluster_dns_name", cluster.dns_name)
