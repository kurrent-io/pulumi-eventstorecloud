package live

import (
	"os"
	"testing"

	"github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// TestPeeringAWS validates the Peering resource end-to-end by self-provisioning the
// peer side in AWS. It stands up a throwaway VPC with the Pulumi AWS provider,
// creates the Kurrent peering against it, accepts the AWS-side connection using the
// peering's provider_metadata["aws_peering_link_id"], and asserts the peering
// reaches "active". Everything is torn down by the harness, so nothing leaks. A VPC
// and a peering connection are free, so the AWS cost is ~0.
//
// Gated on the KURRENT_TEST_AWS_* credentials (the same set the CloudWatch tests
// use). The IAM identity needs ec2 VPC + peering permissions and
// sts:GetCallerIdentity (see TESTING.md). KURRENT_TEST_AWS_SESSION_TOKEN is only
// needed for temporary (STS) credentials. The peer account id is derived from the
// credentials via GetCallerIdentity, but can be overridden with
// KURRENT_TEST_PEER_ACCOUNT_ID.
//
// The peer VPC and the Kurrent network are created in the same region
// (KURRENT_TEST_AWS_REGION) for same-region peering.
func TestPeeringAWS(t *testing.T) {
	requireCreds(t)
	skipUnlessEnv(t,
		"KURRENT_TEST_AWS_ACCESS_KEY_ID",
		"KURRENT_TEST_AWS_SECRET_ACCESS_KEY",
		"KURRENT_TEST_AWS_REGION",
	)

	awsRegion := os.Getenv("KURRENT_TEST_AWS_REGION")
	awsAccessKey := os.Getenv("KURRENT_TEST_AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("KURRENT_TEST_AWS_SECRET_ACCESS_KEY")
	awsToken := os.Getenv("KURRENT_TEST_AWS_SESSION_TOKEN") // set only for temporary STS creds

	// The peer VPC CIDR must not overlap the Kurrent network CIDR (172.21.0.0/16).
	const peerCidr = "10.99.0.0/16"
	const kurrentCidr = "172.21.0.0/16"

	ctx, stack := newStack(t, func(ctx *pulumi.Context) error {
		// Configure the AWS provider explicitly from KURRENT_TEST_AWS_* so the test
		// does not depend on ambient AWS_* environment or a shared credentials file.
		provArgs := &aws.ProviderArgs{
			Region:    pulumi.String(awsRegion),
			AccessKey: pulumi.String(awsAccessKey),
			SecretKey: pulumi.String(awsSecretKey),
		}
		if awsToken != "" {
			provArgs.Token = pulumi.String(awsToken)
		}
		awsProv, err := aws.NewProvider(ctx, "aws", provArgs)
		if err != nil {
			return err
		}
		awsOpt := pulumi.Provider(awsProv)

		// Derive the peer AWS account id from the credentials (overridable via env).
		accountID := os.Getenv("KURRENT_TEST_PEER_ACCOUNT_ID")
		if accountID == "" {
			ident, err := aws.GetCallerIdentity(ctx, &aws.GetCallerIdentityArgs{}, awsOpt)
			if err != nil {
				return err
			}
			accountID = ident.AccountId
		}

		// Throwaway peer VPC in the caller's AWS account.
		vpc, err := ec2.NewVpc(ctx, "peer-vpc", &ec2.VpcArgs{
			CidrBlock: pulumi.String(peerCidr),
			Tags:      pulumi.StringMap{"Name": pulumi.String(resName("kurrentcloud-peer-test"))},
		}, awsOpt)
		if err != nil {
			return err
		}

		p, err := kurrentcloud.NewProject(ctx, "project", &kurrentcloud.ProjectArgs{
			Name: pulumi.String(resName("tt-peer-project")),
		})
		if err != nil {
			return err
		}
		// The Kurrent network must be in the same region as the peer VPC.
		net, err := kurrentcloud.NewNetwork(ctx, "network", &kurrentcloud.NetworkArgs{
			Name:             pulumi.String(resName("tt-peer-network")),
			ProjectId:        p.ID().ToStringOutput(),
			ResourceProvider: pulumi.String("aws"),
			Region:           pulumi.String(awsRegion),
			CidrBlock:        pulumi.String(kurrentCidr),
		})
		if err != nil {
			return err
		}

		peer, err := kurrentcloud.NewPeering(ctx, "peering", &kurrentcloud.PeeringArgs{
			Name:                 pulumi.String(resName("tt-peering")),
			ProjectId:            p.ID().ToStringOutput(),
			NetworkId:            net.ID().ToStringOutput(),
			PeerResourceProvider: pulumi.String("aws"),
			PeerAccountId:        pulumi.String(accountID),
			PeerNetworkId:        vpc.ID().ToStringOutput(),
			PeerNetworkRegion:    pulumi.String(awsRegion),
			Routes:               pulumi.StringArray{vpc.CidrBlock},
		})
		if err != nil {
			return err
		}

		// Accept the Kurrent-initiated peering on the AWS side. The Kurrent provider
		// surfaces the AWS VPC peering connection id (pcx-...) as the
		// "aws_peering_link_id" key of provider_metadata. The accepter waits for the
		// connection to become active.
		accepter, err := ec2.NewVpcPeeringConnectionAccepter(ctx, "accepter", &ec2.VpcPeeringConnectionAccepterArgs{
			VpcPeeringConnectionId: peer.ProviderMetadata.MapIndex(pulumi.String("aws_peering_link_id")),
			AutoAccept:             pulumi.Bool(true),
			Tags:                   pulumi.StringMap{"Name": pulumi.String(resName("kurrentcloud-peer-test"))},
		}, awsOpt)
		if err != nil {
			return err
		}

		ctx.Export("peeringId", peer.ID())
		ctx.Export("peerAccountId", pulumi.String(accountID))
		ctx.Export("acceptStatus", accepter.AcceptStatus)
		return nil
	})

	res := up(t, ctx, stack)
	assertNonEmpty(t, res, "peeringId")
	if got := outString(t, res, "acceptStatus"); got != "active" {
		t.Errorf("peering accept status = %q, want %q (the AWS side did not establish)", got, "active")
	}
}
