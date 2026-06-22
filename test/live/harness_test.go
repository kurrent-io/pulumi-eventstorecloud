package live

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const projectName = "kurrentcloud-live-tests"

// TestMain wires up a self-contained local Pulumi backend so the suite needs no
// Pulumi Cloud login. Kurrent Cloud credentials are read by the provider plugin
// from the inherited ESC_* environment variables (see requireCreds / TESTING.md).
func TestMain(m *testing.M) {
	if os.Getenv("PULUMI_BACKEND_URL") == "" {
		dir, err := os.MkdirTemp("", "kurrent-live-backend-")
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed to create temp Pulumi backend:", err)
			os.Exit(1)
		}
		_ = os.Setenv("PULUMI_BACKEND_URL", "file://"+dir)
	}
	// A passphrase (even an empty one) is required for the file backend's secret
	// provider; default it so the suite is non-interactive.
	if os.Getenv("PULUMI_CONFIG_PASSPHRASE") == "" && os.Getenv("PULUMI_CONFIG_PASSPHRASE_FILE") == "" {
		_ = os.Setenv("PULUMI_CONFIG_PASSPHRASE", "")
	}
	os.Exit(m.Run())
}

// requireCreds skips the test unless Kurrent Cloud credentials are present. The
// provider reads ESC_TOKEN/ESC_ORG_ID (and optionally ESC_URL) natively, so the
// default provider configures itself from the environment.
func requireCreds(t *testing.T) {
	t.Helper()
	if os.Getenv("ESC_TOKEN") == "" || os.Getenv("ESC_ORG_ID") == "" {
		t.Skip("skipping live test: set ESC_TOKEN and ESC_ORG_ID (see TESTING.md)")
	}
}

// skipUnlessEnv skips the test unless every listed env var is set, naming the
// missing ones so the operator knows exactly what to configure.
func skipUnlessEnv(t *testing.T, keys ...string) {
	t.Helper()
	var missing []string
	for _, k := range keys {
		if os.Getenv(k) == "" {
			missing = append(missing, k)
		}
	}
	if len(missing) > 0 {
		t.Skipf("skipping: set %s to run this test (see TESTING.md)", strings.Join(missing, ", "))
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

// clusterParams holds the cost-contained managed-cluster shape, overridable via
// KURRENT_TEST_* environment variables (defaults favor the smallest billable run).
type clusterParams struct {
	ResourceProvider string
	Region           string
	InstanceType     string
	Topology         string
	DiskType         string
	DiskSize         int
	DiskIops         int
	DiskThroughput   int
	ServerVersion    string
	CidrBlock        string
}

func clusterParamsFromEnv() clusterParams {
	return clusterParams{
		ResourceProvider: envOr("KURRENT_TEST_RESOURCE_PROVIDER", "aws"),
		Region:           envOr("KURRENT_TEST_REGION", "us-west-2"),
		InstanceType:     envOr("KURRENT_TEST_INSTANCE_TYPE", "F1"),
		Topology:         envOr("KURRENT_TEST_TOPOLOGY", "single-node"),
		DiskType:         envOr("KURRENT_TEST_DISK_TYPE", "gp3"),
		DiskSize:         envInt("KURRENT_TEST_DISK_SIZE", 16),
		DiskIops:         envInt("KURRENT_TEST_DISK_IOPS", 3000),
		DiskThroughput:   envInt("KURRENT_TEST_DISK_THROUGHPUT", 125),
		ServerVersion:    envOr("KURRENT_TEST_SERVER_VERSION", "24.10"),
		CidrBlock:        envOr("KURRENT_TEST_CIDR", "172.21.0.0/16"),
	}
}

const nameCharset = "abcdefghijklmnopqrstuvwxyz0123456789"

// randSuffix returns a short random suffix so resource names don't collide
// across runs or with leftovers from a previously failed teardown. Go's global
// rand is auto-seeded (Go 1.20+), so this varies per process.
func randSuffix() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = nameCharset[rand.Intn(len(nameCharset))]
	}
	return string(b)
}

// resName builds a unique, human-recognizable Kurrent Cloud resource name.
func resName(prefix string) string {
	return prefix + "-" + randSuffix()
}

func sanitizeStackName(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '-':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	return b.String()
}

func uniqueStackName(t *testing.T) string {
	return fmt.Sprintf("%s-%d", sanitizeStackName(t.Name()), os.Getpid()%100000)
}

// newStack creates an ephemeral inline-program stack and registers guaranteed
// teardown (destroy + remove) via t.Cleanup so no cloud resources leak even when
// assertions fail.
func newStack(t *testing.T, deploy pulumi.RunFunc) (context.Context, auto.Stack) {
	t.Helper()
	ctx := context.Background()
	name := uniqueStackName(t)
	stack, err := auto.UpsertStackInlineSource(ctx, name, projectName, deploy)
	if err != nil {
		t.Fatalf("create stack %q: %v", name, err)
	}
	t.Cleanup(func() {
		// Destroy with retries: a cluster mutation immediately before teardown can
		// leave the cluster briefly locked ("failed to obtain lock on cluster") or
		// otherwise busy, which a single destroy attempt would hit. Retry a few
		// times before giving up so we don't leak billable resources.
		var derr error
		for attempt := 1; attempt <= 4; attempt++ {
			if _, derr = stack.Destroy(ctx, optdestroy.ProgressStreams(os.Stderr)); derr == nil {
				break
			}
			t.Logf("destroy attempt %d/4 for stack %q failed: %v", attempt, name, derr)
			if attempt < 4 {
				time.Sleep(90 * time.Second)
			}
		}
		if derr != nil {
			t.Logf("WARNING: destroy failed for stack %q after retries: %v (manual cleanup may be required in Kurrent Cloud)", name, derr)
		}
		if err := stack.Workspace().RemoveStack(ctx, name); err != nil {
			t.Logf("WARNING: remove stack %q: %v", name, err)
		}
	})
	return ctx, stack
}

// up runs `pulumi up`, failing the test on error, and returns the result for
// assertions.
func up(t *testing.T, ctx context.Context, stack auto.Stack) auto.UpResult {
	t.Helper()
	res, err := stack.Up(ctx, optup.ProgressStreams(os.Stderr))
	if err != nil {
		t.Fatalf("pulumi up: %v", err)
	}
	return res
}

// outString returns a named stack output as a string, failing the test if it is
// missing or not a string.
func outString(t *testing.T, res auto.UpResult, key string) string {
	t.Helper()
	v, ok := res.Outputs[key]
	if !ok {
		t.Fatalf("missing stack output %q", key)
	}
	s, ok := v.Value.(string)
	if !ok {
		t.Fatalf("stack output %q is not a string: %T", key, v.Value)
	}
	return s
}

// assertNonEmpty fails the test if a named string output is empty.
func assertNonEmpty(t *testing.T, res auto.UpResult, key string) string {
	t.Helper()
	s := outString(t, res, key)
	if s == "" {
		t.Errorf("expected non-empty stack output %q", key)
	}
	return s
}

// assertNoReplacements asserts an update changed resources in place (no
// destroy/recreate) — the key invariant behind both in-place updates and the
// eventstorecloud->kurrentcloud alias migration.
func assertNoReplacements(t *testing.T, res auto.UpResult) {
	t.Helper()
	if res.Summary.ResourceChanges == nil {
		return
	}
	ch := *res.Summary.ResourceChanges
	replaced := ch["replace"] + ch["replaced"] + ch["create-replacement"] + ch["delete-replaced"]
	if replaced > 0 {
		t.Errorf("expected no replacements but update reported replacements; resource changes=%v", ch)
	}
}
