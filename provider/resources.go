// Copyright 2024, Kurrent, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package eventstorecloud

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/EventStore/pulumi-eventstorecloud/provider/pkg/version"
	"github.com/kurrent-io/terraform-provider-kurrentcloud/v2/esc"
	"github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfbridge"
	shim "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim"
	shimv2 "github.com/pulumi/pulumi-terraform-bridge/v3/pkg/tfshim/sdk-v2"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

const (
	mainPkg = "kurrentcloud"
	mainMod = "index"
)

var namespaceMap = map[string]string{
	"kurrentcloud": "KurrentCloud",
}

// makeMember manufactures a type token for the package and the given module and type.
func makeMember(mod string, mem string) tokens.ModuleMember {
	moduleName := strings.ToLower(mod)
	namespaceMap[moduleName] = mod
	fn := string(unicode.ToLower(rune(mem[0]))) + mem[1:]
	token := moduleName + "/" + fn
	return tokens.ModuleMember(mainPkg + ":" + token + ":" + mem)
}

// makeType manufactures a type token for the package and the given module and type.
func makeType(mod string, typ string) tokens.Type {
	return tokens.Type(makeMember(mod, typ))
}

// makeDataSource manufactures a standard resource token given a module and resource name.  It
// automatically uses the main package and names the file by simply lower casing the data source's
// first character.
func makeDataSource(mod string, res string) tokens.ModuleMember {
	return makeMember(mod, res)
}

// makeResource manufactures a standard resource token given a module and resource name.  It
// automatically uses the main package and names the file by simply lower casing the resource's
// first character.
func makeResource(mod string, res string) tokens.Type {
	return makeType(mod, res)
}

// legacyAliases returns the historical eventstorecloud:index token for a resource type.
// Declaring it as an alias lets existing Pulumi stacks (created with the eventstorecloud
// package) refresh onto the renamed kurrentcloud token without destroying and recreating
// the underlying cloud resources.
func legacyAliases(typ string) []tfbridge.AliasInfo {
	fn := string(unicode.ToLower(rune(typ[0]))) + typ[1:]
	tok := "eventstorecloud:" + mainMod + "/" + fn + ":" + typ
	return []tfbridge.AliasInfo{{Type: &tok}}
}

// preConfigureCallback is called before the providerConfigure function of the underlying provider.
// It should validate that the provider can be configured, and provide actionable errors in the case
// it cannot be. Configuration variables can be read from `vars` using the `stringValue` function -
// for example `stringValue(vars, "accessKey")`.
func preConfigureCallback(vars resource.PropertyMap, c shim.ResourceConfig) error {
	return nil
}

// Provider returns additional overlaid schema and metadata associated with the provider..
func Provider() tfbridge.ProviderInfo {
	// Instantiate the Terraform provider
	p := shimv2.NewProvider(esc.New("")())

	// Create a Pulumi provider mapping
	prov := tfbridge.ProviderInfo{
		P:                    p,
		Name:                 "kurrentcloud",
		DisplayName:          "Kurrent Cloud",
		Publisher:            "Kurrent",
		Description:          "A Pulumi package for creating and managing Kurrent Cloud resources.",
		Keywords:             []string{"pulumi", "kurrentcloud", "kurrent", "eventstore", "eventstorecloud"},
		License:              "Apache-2.0",
		Homepage:             "https://www.kurrent.io",
		Repository:           "https://github.com/kurrent-io/pulumi-eventstorecloud",
		PluginDownloadURL:    "github://api.github.com/kurrent-io",
		GitHubOrg:            "kurrent-io",
		Config:               map[string]*tfbridge.SchemaInfo{},
		PreConfigureCallback: preConfigureCallback,
		// The underlying Terraform provider (kurrentcloud v2) registers every resource under
		// both the preferred kurrentcloud_* name and the deprecated eventstorecloud_* alias.
		// We map the kurrentcloud_* names and attach Pulumi aliases back to the historical
		// eventstorecloud:index:* tokens, so existing Pulumi stacks refresh onto the renamed
		// tokens without resource replacement. The duplicate eventstorecloud_* TF names are
		// ignored to keep a single, unambiguous Pulumi resource per concept.
		IgnoreMappings: []string{
			"eventstorecloud_project",
			"eventstorecloud_acl",
			"eventstorecloud_network",
			"eventstorecloud_peering",
			"eventstorecloud_managed_cluster",
			"eventstorecloud_scheduled_backup",
			"eventstorecloud_integration",
			"eventstorecloud_integration_awscloudwatch_logs",
			"eventstorecloud_integration_awscloudwatch_metrics",
		},
		Resources: map[string]*tfbridge.ResourceInfo{
			"kurrentcloud_project":                           {Tok: makeResource(mainMod, "Project"), Aliases: legacyAliases("Project")},
			"kurrentcloud_acl":                               {Tok: makeResource(mainMod, "Acl"), Aliases: legacyAliases("Acl")},
			"kurrentcloud_network":                           {Tok: makeResource(mainMod, "Network"), Aliases: legacyAliases("Network")},
			"kurrentcloud_peering":                           {Tok: makeResource(mainMod, "Peering"), Aliases: legacyAliases("Peering")},
			"kurrentcloud_managed_cluster":                   {Tok: makeResource(mainMod, "ManagedCluster"), Aliases: legacyAliases("ManagedCluster")},
			"kurrentcloud_managed_cluster_replicaset":        {Tok: makeResource(mainMod, "ManagedClusterReplicaset")},
			"kurrentcloud_scheduled_backup":                  {Tok: makeResource(mainMod, "ScheduledBackup"), Aliases: legacyAliases("ScheduledBackup")},
			"kurrentcloud_integration":                       {Tok: makeResource(mainMod, "Integration"), Aliases: legacyAliases("Integration")},
			"kurrentcloud_integration_awscloudwatch_logs":    {Tok: makeResource(mainMod, "AWSCloudWatchLogsIntegration"), Aliases: legacyAliases("AWSCloudWatchLogsIntegration")},
			"kurrentcloud_integration_awscloudwatch_metrics": {Tok: makeResource(mainMod, "AWSCloudWatchMetricsIntegration"), Aliases: legacyAliases("AWSCloudWatchMetricsIntegration")},
		},
		DataSources: map[string]*tfbridge.DataSourceInfo{
			"kurrentcloud_project": {Tok: makeDataSource(mainMod, "getProject")},
			"kurrentcloud_network": {Tok: makeDataSource(mainMod, "getNetwork")},
		},
		JavaScript: &tfbridge.JavaScriptInfo{
			Dependencies: map[string]string{
				"@pulumi/pulumi": "^3.0.0",
			},
			DevDependencies: map[string]string{
				"@types/node": "^10.0.0", // so we can access strongly typed node definitions.
				"@types/mime": "^2.0.0",
			},
			PackageName: "@kurrent-io/pulumi-kurrentcloud",
		},
		Python: &tfbridge.PythonInfo{
			Requires: map[string]string{
				"pulumi": ">=3.0.0,<4.0.0",
			},
		},
		Golang: &tfbridge.GolangInfo{
			// NOTE: the SDK Go module path (sdk/go.mod) is still rooted at the historical
			// EventStore/pulumi-eventstorecloud repository. Only the leaf package is renamed to
			// kurrentcloud here. Renaming the repository + Go module path is a tracked follow-up.
			ImportBasePath: filepath.Join(
				"github.com/EventStore/pulumi-eventstorecloud/sdk",
				tfbridge.GetModuleMajorVersion(version.Version),
				"go",
				mainPkg,
			),
			GenerateResourceContainerTypes: true,
		},
		CSharp: &tfbridge.CSharpInfo{
			PackageReferences: map[string]string{
				"Pulumi":                       "3.*",
				"System.Collections.Immutable": "5.0.0",
			},
			Namespaces: namespaceMap,
		},
	}

	prov.SetAutonaming(255, "-")

	return prov
}
