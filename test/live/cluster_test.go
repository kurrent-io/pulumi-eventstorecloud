package live

import (
	"os"
	"testing"

	"github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestManagedClusterLifecycle is the expensive, end-to-end cluster test. To keep
// cost down it provisions a single cost-contained managed cluster and exercises
// every cluster-dependent resource against it: ManagedCluster, the new
// ManagedClusterReplicaset, and ScheduledBackup. It then performs an in-place
// projectionLevel update and asserts no resources are replaced (the headline
// kurrentcloud v2.1.0 behavior).
//
// Set KURRENT_TEST_SKIP_CLUSTER=1 to skip just this test during fast iterations.
func TestManagedClusterLifecycle(t *testing.T) {
	requireCreds(t)
	if os.Getenv("KURRENT_TEST_SKIP_CLUSTER") != "" {
		t.Skip("KURRENT_TEST_SKIP_CLUSTER set; skipping the billable cluster test")
	}
	cp := clusterParamsFromEnv()
	projectionLevel := "off"
	// Capture randomized names once, outside the program closure. The closure
	// runs on every up(), so calling resName() inside it would regenerate the
	// Name inputs on the second up() and cause unrelated diffs/replacements,
	// defeating the projectionLevel-only update assertion below.
	projName := resName("tt-cluster-project")
	netName := resName("tt-cluster-network")
	clusName := resName("tt-cluster")

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
			// Read-only replica sets require a multi-node topology (per the upstream
			// kurrentcloud_managed_cluster_replicaset example). A single-node cluster
			// rejects replica creation with "operation cannot be performed in the current state".
			Topology:        pulumi.String("three-node-multi-zone"),
			InstanceType:    pulumi.String(cp.InstanceType),
			DiskSize:        pulumi.Int(24),
			DiskType:        pulumi.String(cp.DiskType),
			DiskIops:        pulumi.Int(cp.DiskIops),
			DiskThroughput:  pulumi.Int(cp.DiskThroughput),
			ServerVersion:   pulumi.String(cp.ServerVersion),
			ProjectionLevel: pulumi.String(projectionLevel),
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
			// Create the backup after the replica so the two cluster mutations are
			// serialized rather than racing (defensive: cluster -> replica -> backup).
		}, pulumi.DependsOn([]pulumi.Resource{replica}))
		if err != nil {
			return err
		}

		ctx.Export("clusterId", cluster.ID())
		ctx.Export("clusterDnsName", cluster.DnsName)
		ctx.Export("clusterProjectionLevel", cluster.ProjectionLevel.Elem())
		ctx.Export("replicaId", replica.ID())
		ctx.Export("replicaStatus", replica.Status)
		ctx.Export("backupId", backup.ID())
		return nil
	})

	res := up(t, ctx, stack)
	clusterID := assertNonEmpty(t, res, "clusterId")
	assertNonEmpty(t, res, "clusterDnsName")
	assertNonEmpty(t, res, "replicaId")
	assertNonEmpty(t, res, "replicaStatus")
	assertNonEmpty(t, res, "backupId")
	if got := outString(t, res, "clusterProjectionLevel"); got != "off" {
		t.Errorf("projectionLevel = %q, want %q", got, "off")
	}

	// Update projectionLevel off -> system. As of kurrentcloud v2.1.0 this is an
	// in-place update, so nothing should be replaced.
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
