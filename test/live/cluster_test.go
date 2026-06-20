package live

import (
	"os"
	"testing"

	"github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestManagedClusterProjectionUpdate validates the in-place projectionLevel update
// (the headline kurrentcloud v2.1.0 behavior: changing projectionLevel updates the
// cluster instead of replacing it). It uses a cost-contained single-node cluster,
// since this behavior does not require a multi-node topology.
//
// Set KURRENT_TEST_SKIP_CLUSTER=1 to skip the billable cluster tests.
func TestManagedClusterProjectionUpdate(t *testing.T) {
	requireCreds(t)
	if os.Getenv("KURRENT_TEST_SKIP_CLUSTER") != "" {
		t.Skip("KURRENT_TEST_SKIP_CLUSTER set; skipping the billable cluster tests")
	}
	cp := clusterParamsFromEnv()
	projectionLevel := "off"
	// Captured once: the closure re-runs on each up(), so regenerating names inside
	// it would change Name inputs on the second up() and defeat the no-replace check.
	projName := resName("tt-proj-project")
	netName := resName("tt-proj-network")
	clusName := resName("tt-proj-cluster")

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(projName),
		})
		if err != nil {
			return err
		}
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(netName),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String(cp.ResourceProvider),
			Region:           pulumi.String(cp.Region),
			CidrBlock:        pulumi.String(cp.CidrBlock),
		})
		if err != nil {
			return err
		}
		cluster, err := kurrentcloud.NewManagedCluster(ctx, "cluster", &kurrentcloud.ManagedClusterArgs{
			Name:            pulumi.String(clusName),
			ProjectId:       p.ID().ToStringOutput(),
			NetworkId:       net.ID().ToStringOutput(),
			Topology:        pulumi.String(cp.Topology),
			InstanceType:    pulumi.String(cp.InstanceType),
			DiskSize:        pulumi.Int(cp.DiskSize),
			DiskType:        pulumi.String(cp.DiskType),
			DiskIops:        pulumi.Int(cp.DiskIops),
			DiskThroughput:  pulumi.Int(cp.DiskThroughput),
			ServerVersion:   pulumi.String(cp.ServerVersion),
			ProjectionLevel: pulumi.String(projectionLevel),
		})
		if err != nil {
			return err
		}
		ctx.Export("clusterId", cluster.ID())
		ctx.Export("clusterDnsName", cluster.DnsName)
		ctx.Export("clusterProjectionLevel", cluster.ProjectionLevel.Elem())
		return nil
	})

	// Create with projectionLevel "off".
	res := up(t, ctx, stack)
	clusterID := assertNonEmpty(t, res, "clusterId")
	assertNonEmpty(t, res, "clusterDnsName")
	if got := outString(t, res, "clusterProjectionLevel"); got != "off" {
		t.Errorf("projectionLevel = %q, want %q", got, "off")
	}

	// In-place update off -> system: nothing should be replaced.
	projectionLevel = "system"
	res2 := up(t, ctx, stack)
	assertNoReplacements(t, res2)
	if got := outString(t, res2, "clusterProjectionLevel"); got != "system" {
		t.Errorf("after update, projectionLevel = %q, want %q", got, "system")
	}
	if id2 := assertNonEmpty(t, res2, "clusterId"); id2 != clusterID {
		t.Errorf("cluster id changed across projectionLevel update (%q -> %q): unexpected replacement", clusterID, id2)
	}
}

// TestManagedClusterReplicaset validates the new ManagedClusterReplicaset resource
// (and ScheduledBackup) end-to-end. Read-only replica sets require a
// three-node-multi-zone cluster (per the upstream provider example); a single-node
// cluster rejects replica creation with "operation cannot be performed in the
// current state". This is the most expensive test (a multi-node cluster + replica).
func TestManagedClusterReplicaset(t *testing.T) {
	requireCreds(t)
	if os.Getenv("KURRENT_TEST_SKIP_CLUSTER") != "" {
		t.Skip("KURRENT_TEST_SKIP_CLUSTER set; skipping the billable cluster tests")
	}
	cp := clusterParamsFromEnv()
	projName := resName("tt-rep-project")
	netName := resName("tt-rep-network")
	clusName := resName("tt-rep-cluster")

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(projName),
		})
		if err != nil {
			return err
		}
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(netName),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String(cp.ResourceProvider),
			Region:           pulumi.String(cp.Region),
			CidrBlock:        pulumi.String(cp.CidrBlock),
		})
		if err != nil {
			return err
		}
		cluster, err := kurrentcloud.NewManagedCluster(ctx, "cluster", &kurrentcloud.ManagedClusterArgs{
			Name:      pulumi.String(clusName),
			ProjectId: p.ID().ToStringOutput(),
			NetworkId: net.ID().ToStringOutput(),
			// Read-only replicas require a multi-node topology (per the upstream
			// kurrentcloud_managed_cluster_replicaset example).
			Topology:        pulumi.String("three-node-multi-zone"),
			InstanceType:    pulumi.String(cp.InstanceType),
			DiskSize:        pulumi.Int(24),
			DiskType:        pulumi.String(cp.DiskType),
			DiskIops:        pulumi.Int(cp.DiskIops),
			DiskThroughput:  pulumi.Int(cp.DiskThroughput),
			ServerVersion:   pulumi.String(cp.ServerVersion),
			ProjectionLevel: pulumi.String("off"),
		})
		if err != nil {
			return err
		}
		replica, err := kurrentcloud.NewManagedClusterReplicaset(ctx, "replica", &kurrentcloud.ManagedClusterReplicasetArgs{
			ProjectId:    p.ID().ToStringOutput(),
			ClusterId:    cluster.ID().ToStringOutput(),
			ReplicaCount: pulumi.Int(1),
		})
		if err != nil {
			return err
		}
		backup, err := kurrentcloud.NewScheduledBackup(ctx, "backup", &kurrentcloud.ScheduledBackupArgs{
			ProjectId:         p.ID().ToStringOutput(),
			SourceClusterId:   cluster.ID().ToStringOutput(),
			Schedule:          pulumi.String("0 0 * * *"),
			MaxBackupCount:    pulumi.Int(3),
			Description:       pulumi.String("kurrentcloud live test backup"),
			BackupDescription: pulumi.String("automated test backup"),
			// Serialize the two cluster mutations (replica add, backup create) rather
			// than letting Pulumi run them concurrently against the same cluster.
		}, pulumi.DependsOn([]pulumi.Resource{replica}))
		if err != nil {
			return err
		}
		ctx.Export("clusterId", cluster.ID())
		ctx.Export("clusterDnsName", cluster.DnsName)
		ctx.Export("replicaId", replica.ID())
		ctx.Export("replicaStatus", replica.Status)
		ctx.Export("backupId", backup.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "clusterId")
	assertNonEmpty(t, res, "clusterDnsName")
	assertNonEmpty(t, res, "replicaId")
	assertNonEmpty(t, res, "backupId")
	if got := outString(t, res, "replicaStatus"); got != "available" {
		t.Errorf("replicaStatus = %q, want %q", got, "available")
	}
}
