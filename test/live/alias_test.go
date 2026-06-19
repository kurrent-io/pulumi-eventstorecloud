package live

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

// TestProviderAliasesInSchema is an offline regression guard for the
// eventstorecloud -> kurrentcloud rename. Every renamed resource must declare a
// Pulumi alias back to its historical eventstorecloud:index token; otherwise an
// existing stack would replace (destroy + recreate) real cloud resources on
// upgrade. The new ManagedClusterReplicaset must NOT carry an alias.
//
// This needs no credentials and runs on every CI build. The live, empirical
// "0 replacements" proof lives in test/alias-migration/.
func TestProviderAliasesInSchema(t *testing.T) {
	const schemaPath = "../../provider/cmd/pulumi-resource-kurrentcloud/schema.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Skipf("schema not found at %s (run `make tfgen` first): %v", schemaPath, err)
	}

	var schema struct {
		Resources map[string]struct {
			Aliases []struct {
				Type string `json:"type"`
			} `json:"aliases"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("parse schema: %v", err)
	}
	if len(schema.Resources) == 0 {
		t.Fatal("schema contains no resources")
	}

	const newPrefix = "kurrentcloud:"
	const oldPrefix = "eventstorecloud:"
	const replicaset = "kurrentcloud:index/managedClusterReplicaset:ManagedClusterReplicaset"

	sawReplicaset := false
	for tok, r := range schema.Resources {
		if !strings.HasPrefix(tok, newPrefix) {
			t.Errorf("unexpected non-kurrentcloud resource token in schema: %s", tok)
			continue
		}
		if tok == replicaset {
			sawReplicaset = true
			if len(r.Aliases) != 0 {
				t.Errorf("%s is new in kurrentcloud and must have no legacy alias, got %v", tok, r.Aliases)
			}
			continue
		}
		want := oldPrefix + strings.TrimPrefix(tok, newPrefix)
		found := false
		for _, a := range r.Aliases {
			if a.Type == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s is missing legacy alias %q (has %v): existing eventstorecloud stacks would be replaced on upgrade", tok, want, r.Aliases)
		}
	}
	if !sawReplicaset {
		t.Errorf("expected %s in schema (the new parity resource)", replicaset)
	}
}
