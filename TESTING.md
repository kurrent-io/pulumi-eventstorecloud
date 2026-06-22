# Testing the Kurrent Cloud Pulumi provider

This repository has three layers of tests:

| Layer | Location | Needs cloud? | What it covers |
|---|---|---|---|
| **Offline alias guard** | `test/live` (`TestProviderAliasesInSchema`) | No | Every renamed resource still declares its `eventstorecloud:index/<resource>:<Type>` alias (so existing stacks aren't replaced on upgrade). Runs on every PR. |
| **Live end-to-end suite** | `test/live` | Yes | create → update → destroy for **every** resource and data source against a real Kurrent Cloud org, using the Pulumi Automation API. |
| **Live alias migration** | `test/alias-migration/run.sh` | Yes | Empirically proves an `eventstorecloud` stack migrates to `kurrentcloud` with **zero replacements**. |

> ⚠️ **The live suite provisions real, billable infrastructure** (networks, and — unless skipped — a managed cluster). It always tears resources down, but a killed run can leak; see [Cleanup](#cleanup).

## What gets tested

Resources: `Project`, `Network`, `Acl`, `Peering`, `ManagedCluster`, `ManagedClusterReplicaset`, `ScheduledBackup`, `Integration`, `AWSCloudWatchLogsIntegration`, `AWSCloudWatchMetricsIntegration`. Data sources: `getProject`, `getNetwork`. Plus an in-place `projectionLevel` update (asserts no replacement) and an in-place project rename.

Tests whose external prerequisites aren't configured **skip themselves** (and say which env vars are missing), so you can run as much or as little as your environment supports.

## Prerequisites

- **Go** ≥ 1.25, **Pulumi CLI**, and **pulumictl** (`brew install pulumi pulumi/tap/pulumictl`).
- The **provider plugin built and on `PATH`**:
  ```bash
  make provider                 # builds ./bin/pulumi-resource-kurrentcloud
  export PATH="$PWD/bin:$PATH"
  ```
- **Kurrent Cloud credentials** (see below).

No Pulumi Cloud account is needed — the suite uses a throwaway local file backend automatically.

## Environment variables

### Required — Kurrent Cloud credentials

| Variable | Required | Where the value comes from |
|---|---|---|
| `ESC_TOKEN` | **Yes** | A Kurrent Cloud **refresh token** for an account with admin access to the org. Generate it in the Kurrent Cloud console (Access / API tokens), or reuse the token your Terraform/`kurrent` CLI setup uses — it is the same token the provider reads natively. |
| `ESC_ORG_ID` | **Yes** | Your Kurrent Cloud **organization ID** (Org settings in the console, or the `…/organizations/<id>/…` segment of the console URL). |
| `ESC_URL` | No | API endpoint. Defaults to `https://api.eventstore.cloud`. Set only if targeting a non-default (e.g. staging) control plane. |

If `ESC_TOKEN`/`ESC_ORG_ID` are unset, **all live tests skip** and only the offline alias guard runs.

### Optional — managed-cluster shape (cost containment)

Sensible, cost-contained defaults are used; override only if needed.

| Variable | Default | Notes |
|---|---|---|
| `KURRENT_TEST_RESOURCE_PROVIDER` | `aws` | `aws`, `gcp`, or `azure`. |
| `KURRENT_TEST_REGION` | `us-west-2` | Region for the network/cluster. |
| `KURRENT_TEST_INSTANCE_TYPE` | `F1` | Smallest instance. |
| `KURRENT_TEST_TOPOLOGY` | `single-node` | Cheapest topology. |
| `KURRENT_TEST_DISK_TYPE` | `gp3` | |
| `KURRENT_TEST_DISK_SIZE` | `16` | GiB. |
| `KURRENT_TEST_DISK_IOPS` | `3000` | Required for `gp3`. |
| `KURRENT_TEST_DISK_THROUGHPUT` | `125` | Required for `gp3`. |
| `KURRENT_TEST_SERVER_VERSION` | `24.10` | |
| `KURRENT_TEST_CIDR` | `172.21.0.0/16` | Network address space. |
| `KURRENT_TEST_SKIP_CLUSTER` | _(unset)_ | Set to `1` to skip the billable cluster test (`TestManagedClusterLifecycle`). |

### Optional — unlock the externally-dependent tests

Each group is **all-or-nothing**: set every variable in a group to run its test, or leave them unset to skip it.

**`TestPeering`** — needs a real network to peer with:

| Variable | Where it comes from |
|---|---|
| `KURRENT_TEST_PEER_ACCOUNT_ID` | Cloud account ID that owns the peer network (e.g. AWS account number). |
| `KURRENT_TEST_PEER_NETWORK_ID` | The peer VPC/VNet/network ID. |
| `KURRENT_TEST_PEER_REGION` | Region of the peer network. |
| `KURRENT_TEST_PEER_ROUTES` | Comma-separated CIDR routes to the peer (e.g. `10.2.0.0/16`). |

**`TestPeeringAWS`** — the self-provisioning peering test: it stands up a throwaway AWS VPC (via Pulumi's AWS provider), peers the Kurrent network with it, **accepts** the connection on the AWS side, asserts it reaches `active`, and tears it all down. Prefer this over `TestPeering` when you don't already have a peer VPC — you only supply AWS credentials and it derives the account id / VPC / CIDR itself. A VPC and a peering connection are free, so the AWS cost is ≈ $0.

| Variable | Where it comes from |
|---|---|
| `KURRENT_TEST_AWS_ACCESS_KEY_ID` | IAM access key (see the policy below). Same var the CloudWatch tests use. |
| `KURRENT_TEST_AWS_SECRET_ACCESS_KEY` | The matching IAM secret. |
| `KURRENT_TEST_AWS_REGION` | Region for both the peer VPC and the Kurrent network (same-region peering), e.g. `us-west-2`. |
| `KURRENT_TEST_AWS_SESSION_TOKEN` | _(optional)_ only for temporary/STS credentials. |
| `KURRENT_TEST_PEER_ACCOUNT_ID` | _(optional)_ overrides the account id otherwise derived via `sts:GetCallerIdentity`. |

Minimal IAM policy for that key (a test user; `Resource: "*"` is fine):

```json
{ "Version": "2012-10-17", "Statement": [{ "Effect": "Allow", "Resource": "*", "Action": [
  "sts:GetCallerIdentity",
  "ec2:CreateVpc", "ec2:DeleteVpc", "ec2:DescribeVpcs", "ec2:DescribeVpcAttribute", "ec2:ModifyVpcAttribute",
  "ec2:CreateTags", "ec2:DeleteTags",
  "ec2:DescribeVpcPeeringConnections", "ec2:AcceptVpcPeeringConnection",
  "ec2:DeleteVpcPeeringConnection", "ec2:ModifyVpcPeeringConnectionOptions"
]}]}
```

> Uses Pulumi's AWS provider, which is added to the `test/live` Go module only — it does not affect the shipped provider or SDKs. Pulumi fetches the `aws` plugin automatically on first run.

**`TestIntegration`** — generic integration sink:

| Variable | Where it comes from |
|---|---|
| `KURRENT_TEST_INTEGRATION_DATA_JSON` | A JSON object for the sink's `data` map, e.g. `{"sink":"opsGenie","apiKey":"…","region":"us"}`. Shape depends on the sink type. |

**`TestAWSCloudWatchLogsIntegration` / `TestAWSCloudWatchMetricsIntegration`** — need AWS IAM creds and an **existing** cluster (so the tests don't provision one):

| Variable | Where it comes from |
|---|---|
| `KURRENT_TEST_AWS_ACCESS_KEY_ID` | IAM access key with CloudWatch logs/metrics write permission. |
| `KURRENT_TEST_AWS_SECRET_ACCESS_KEY` | The matching IAM secret. |
| `KURRENT_TEST_AWS_REGION` | AWS region for the log group / metrics namespace. |
| `KURRENT_TEST_CW_PROJECT_ID` | An existing Kurrent Cloud project ID. |
| `KURRENT_TEST_CW_CLUSTER_ID` | An existing cluster ID in that project to attach the integration to. |
| `KURRENT_TEST_CW_LOG_GROUP` | _(optional)_ CloudWatch log group name. Default `kurrentcloud-live-test`. |
| `KURRENT_TEST_CW_NAMESPACE` | _(optional)_ CloudWatch metrics namespace. Default `kurrentcloud-live-test`. |

## Running locally

```bash
# from the repo root
make provider && export PATH="$PWD/bin:$PATH"   # build + expose the plugin
export ESC_TOKEN="<your refresh token>"
export ESC_ORG_ID="<your org id>"

# everything your environment supports:
make test-live

# just the cheap, credential-free alias guard:
make test-aliases

# a single test, or skip the billable cluster:
cd test/live
go test -v -run TestProject ./...
KURRENT_TEST_SKIP_CLUSTER=1 go test -v ./...
```

### Live alias-migration check

```bash
make build_nodejs                 # build the local Node SDK the script links
make provider && export PATH="$PWD/bin:$PATH"
ESC_TOKEN=… ESC_ORG_ID=… ./test/alias-migration/run.sh
```

## Continuous integration

`.github/workflows/live-tests.yml`:

- **`offline` job** — runs on every PR. No secrets. Runs the alias guard and compiles the harness.
- **`live` job** — runs on **manual dispatch** (Actions → Live tests → Run workflow, with an optional "skip cluster" toggle) and on a **weekly schedule**. Configure these in the repo's **Settings → Secrets and variables → Actions**:
  - **Secrets:** `ESC_TOKEN`, `ESC_ORG_ID`, and (optional) `KURRENT_TEST_AWS_ACCESS_KEY_ID`, `KURRENT_TEST_AWS_SECRET_ACCESS_KEY`, `KURRENT_TEST_INTEGRATION_DATA_JSON`.
  - **Variables:** any `KURRENT_TEST_*` overrides and the `KURRENT_TEST_PEER_*` / `KURRENT_TEST_CW_*` values you want to exercise. Anything left unset just skips its test.

## Cleanup

Every test destroys its stack via `t.Cleanup`. Resource names are prefixed `tt-` with a random suffix. If a run is force-killed and leaks resources, list and remove anything named `tt-*` in the Kurrent Cloud console for the test org.
