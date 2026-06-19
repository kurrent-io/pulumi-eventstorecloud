package live

import (
	"testing"

	"github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestProject covers Project create plus an in-place rename (must update, not replace).
func TestProject(t *testing.T) {
	requireCreds(t)
	base := resName("tt-project")
	renamed := false

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		name := base
		if renamed {
			name = base + "-renamed"
		}
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(name),
		})
		if err != nil {
			return err
		}
		ctx.Export("projectId", p.ID())
		ctx.Export("projectName", p.Name)
		return nil
	})

	res := up(t, ctx, stack)
	id := assertNonEmpty(t, res, "projectId")
	if got := outString(t, res, "projectName"); got != base {
		t.Errorf("project name = %q, want %q", got, base)
	}

	// Rename in place.
	renamed = true
	res2 := up(t, ctx, stack)
	assertNoReplacements(t, res2)
	if got := outString(t, res2, "projectName"); got != base+"-renamed" {
		t.Errorf("renamed project name = %q, want %q", got, base+"-renamed")
	}
	if id2 := assertNonEmpty(t, res2, "projectId"); id2 != id {
		t.Errorf("project id changed across rename (%q -> %q): unexpected replacement", id, id2)
	}
}

// TestNetwork covers Network create within a project.
func TestNetwork(t *testing.T) {
	requireCreds(t)
	cp := clusterParamsFromEnv()

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-net-project")),
		})
		if err != nil {
			return err
		}
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(resName("tt-network")),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String(cp.ResourceProvider),
			Region:           pulumi.String(cp.Region),
			CidrBlock:        pulumi.String(cp.CidrBlock),
		})
		if err != nil {
			return err
		}
		ctx.Export("projectId", p.ID())
		ctx.Export("networkId", net.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "projectId")
	assertNonEmpty(t, res, "networkId")
}

// TestAcl covers Acl create with a CIDR allow-list entry.
func TestAcl(t *testing.T) {
	requireCreds(t)

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-acl-project")),
		})
		if err != nil {
			return err
		}
		acl, err := kurrentcloud.NewAcl(ctx, "acl", &kurrentcloud.AclArgs{
			Name:      pulumi.String(resName("tt-acl")),
			ProjectId: p.ID().ToStringOutput(),
			CidrBlocks: pulumi.MapArray{
				pulumi.Map{
					"address": pulumi.String("10.10.0.0/16"),
					"comment": pulumi.String("kurrentcloud live test"),
				},
			},
		})
		if err != nil {
			return err
		}
		ctx.Export("aclId", acl.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "aclId")
}

// TestDataSources covers the getProject and getNetwork data sources by looking up
// resources created in the same program and asserting the looked-up identifiers
// match the created ones.
func TestDataSources(t *testing.T) {
	requireCreds(t)
	cp := clusterParamsFromEnv()

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-ds-project")),
		})
		if err != nil {
			return err
		}
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(resName("tt-ds-network")),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String(cp.ResourceProvider),
			Region:           pulumi.String(cp.Region),
			CidrBlock:        pulumi.String(cp.CidrBlock),
		})
		if err != nil {
			return err
		}

		// Output-form invokes defer until the resource names are known (post-create).
		lookedProject := kurrentcloud.LookupProjectOutput(ctx, kurrentcloud.LookupProjectOutputArgs{
			Name: p.Name,
		})
		lookedNetwork := kurrentcloud.LookupNetworkOutput(ctx, kurrentcloud.LookupNetworkOutputArgs{
			Name:      net.Name,
			ProjectId: p.ID().ToStringOutput(),
		})

		ctx.Export("createdProjectId", p.ID())
		ctx.Export("lookedProjectId", lookedProject.Id())
		ctx.Export("createdNetworkId", net.ID())
		ctx.Export("lookedNetworkId", lookedNetwork.Id())
		ctx.Export("lookedNetworkCidr", lookedNetwork.CidrBlock())
		return nil
	})

	res := up(t, ctx, stack)
	if a, b := outString(t, res, "createdProjectId"), outString(t, res, "lookedProjectId"); a != b {
		t.Errorf("getProject id mismatch: created %q vs looked up %q", a, b)
	}
	if a, b := outString(t, res, "createdNetworkId"), outString(t, res, "lookedNetworkId"); a != b {
		t.Errorf("getNetwork id mismatch: created %q vs looked up %q", a, b)
	}
	if got := outString(t, res, "lookedNetworkCidr"); got != cp.CidrBlock {
		t.Errorf("getNetwork cidr = %q, want %q", got, cp.CidrBlock)
	}
}
