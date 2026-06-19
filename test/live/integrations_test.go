package live

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestPeering covers the Peering resource. It needs a real peer network to peer
// with, supplied via env. peerResourceProvider defaults to the test resource
// provider (KURRENT_TEST_RESOURCE_PROVIDER).
func TestPeering(t *testing.T) {
	requireCreds(t)
	skipUnlessEnv(t,
		"KURRENT_TEST_PEER_ACCOUNT_ID",
		"KURRENT_TEST_PEER_NETWORK_ID",
		"KURRENT_TEST_PEER_REGION",
		"KURRENT_TEST_PEER_ROUTES",
	)
	cp := clusterParamsFromEnv()
	routes := strings.Split(os.Getenv("KURRENT_TEST_PEER_ROUTES"), ",")

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-peer-project")),
		})
		if err != nil {
			return err
		}
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(resName("tt-peer-network")),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String(cp.ResourceProvider),
			Region:           pulumi.String(cp.Region),
			CidrBlock:        pulumi.String(cp.CidrBlock),
		})
		if err != nil {
			return err
		}
		peer, err := kurrentcloud.NewPeering(ctx, "peering", &kurrentcloud.PeeringArgs{
			Name:                 pulumi.String(resName("tt-peering")),
			ProjectId:            p.ID().ToStringOutput(),
			NetworkId:            net.ID().ToStringOutput(),
			PeerResourceProvider: pulumi.String(cp.ResourceProvider),
			PeerAccountId:        pulumi.String(os.Getenv("KURRENT_TEST_PEER_ACCOUNT_ID")),
			PeerNetworkId:        pulumi.String(os.Getenv("KURRENT_TEST_PEER_NETWORK_ID")),
			PeerNetworkRegion:    pulumi.String(os.Getenv("KURRENT_TEST_PEER_REGION")),
			Routes:               pulumi.ToStringArray(routes),
		})
		if err != nil {
			return err
		}
		ctx.Export("peeringId", peer.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "peeringId")
}

// TestIntegration covers the generic Integration resource. The free-form `data`
// map (sink type and its settings) is supplied as a JSON object via env, e.g.
// KURRENT_TEST_INTEGRATION_DATA_JSON='{"sink":"opsGenie","apiKey":"...","region":"us"}'.
func TestIntegration(t *testing.T) {
	requireCreds(t)
	skipUnlessEnv(t, "KURRENT_TEST_INTEGRATION_DATA_JSON")

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(os.Getenv("KURRENT_TEST_INTEGRATION_DATA_JSON")), &raw); err != nil {
		t.Fatalf("KURRENT_TEST_INTEGRATION_DATA_JSON is not valid JSON: %v", err)
	}
	data := pulumi.Map{}
	for k, v := range raw {
		data[k] = pulumi.Any(v)
	}

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-integration-project")),
		})
		if err != nil {
			return err
		}
		integ, err := kurrentcloud.NewIntegration(ctx, "integration", &kurrentcloud.IntegrationArgs{
			ProjectId:   p.ID().ToStringOutput(),
			Description: pulumi.String("kurrentcloud live test integration"),
			Data:        data,
		})
		if err != nil {
			return err
		}
		ctx.Export("integrationId", integ.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "integrationId")
}

// TestAWSCloudWatchLogsIntegration covers the AWS CloudWatch Logs integration.
// To contain cost it targets an existing cluster (KURRENT_TEST_CW_CLUSTER_ID in
// KURRENT_TEST_CW_PROJECT_ID) rather than provisioning a new one, and needs AWS
// IAM credentials with permission to create/write the log group.
func TestAWSCloudWatchLogsIntegration(t *testing.T) {
	requireCreds(t)
	skipUnlessEnv(t,
		"KURRENT_TEST_AWS_ACCESS_KEY_ID",
		"KURRENT_TEST_AWS_SECRET_ACCESS_KEY",
		"KURRENT_TEST_AWS_REGION",
		"KURRENT_TEST_CW_PROJECT_ID",
		"KURRENT_TEST_CW_CLUSTER_ID",
	)

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		integ, err := kurrentcloud.NewAWSCloudWatchLogsIntegration(ctx, "cwlogs", &kurrentcloud.AWSCloudWatchLogsIntegrationArgs{
			ProjectId:       pulumi.String(os.Getenv("KURRENT_TEST_CW_PROJECT_ID")),
			ClusterIds:      pulumi.StringArray{pulumi.String(os.Getenv("KURRENT_TEST_CW_CLUSTER_ID"))},
			Description:     pulumi.String("kurrentcloud live test cw logs"),
			GroupName:       pulumi.String(envOr("KURRENT_TEST_CW_LOG_GROUP", "kurrentcloud-live-test")),
			Region:          pulumi.String(os.Getenv("KURRENT_TEST_AWS_REGION")),
			AccessKeyId:     pulumi.String(os.Getenv("KURRENT_TEST_AWS_ACCESS_KEY_ID")),
			SecretAccessKey: pulumi.String(os.Getenv("KURRENT_TEST_AWS_SECRET_ACCESS_KEY")),
		})
		if err != nil {
			return err
		}
		ctx.Export("integrationId", integ.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "integrationId")
}

// TestAWSCloudWatchMetricsIntegration covers the AWS CloudWatch Metrics
// integration, targeting an existing cluster to contain cost (see the logs test).
func TestAWSCloudWatchMetricsIntegration(t *testing.T) {
	requireCreds(t)
	skipUnlessEnv(t,
		"KURRENT_TEST_AWS_ACCESS_KEY_ID",
		"KURRENT_TEST_AWS_SECRET_ACCESS_KEY",
		"KURRENT_TEST_AWS_REGION",
		"KURRENT_TEST_CW_PROJECT_ID",
		"KURRENT_TEST_CW_CLUSTER_ID",
	)

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		integ, err := kurrentcloud.NewAWSCloudWatchMetricsIntegration(ctx, "cwmetrics", &kurrentcloud.AWSCloudWatchMetricsIntegrationArgs{
			ProjectId:       pulumi.String(os.Getenv("KURRENT_TEST_CW_PROJECT_ID")),
			ClusterIds:      pulumi.StringArray{pulumi.String(os.Getenv("KURRENT_TEST_CW_CLUSTER_ID"))},
			Description:     pulumi.String("kurrentcloud live test cw metrics"),
			Namespace:       pulumi.String(envOr("KURRENT_TEST_CW_NAMESPACE", "kurrentcloud-live-test")),
			Region:          pulumi.String(os.Getenv("KURRENT_TEST_AWS_REGION")),
			AccessKeyId:     pulumi.String(os.Getenv("KURRENT_TEST_AWS_ACCESS_KEY_ID")),
			SecretAccessKey: pulumi.String(os.Getenv("KURRENT_TEST_AWS_SECRET_ACCESS_KEY")),
		})
		if err != nil {
			return err
		}
		ctx.Export("integrationId", integ.ID())
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "integrationId")
}
