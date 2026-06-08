# Migration Guide — `eventstorecloud` → `kurrentcloud` (Pulumi)

The Pulumi provider has been renamed from **`eventstorecloud`** to **`kurrentcloud`** to match the
rebranded [Kurrent Cloud Terraform provider](https://github.com/kurrent-io/terraform-provider-kurrentcloud)
that it is generated from. This guide covers upgrading an existing Pulumi program.

> **TL;DR** — Update the package import and run `pulumi up`. Every resource type carries an alias to
> its old `eventstorecloud:index:*` token, so **no cloud resources are replaced**. Always confirm
> with `pulumi preview` first (expect `0 to replace`).

## What changed

| | Before (`eventstorecloud`) | After (`kurrentcloud`) |
|---|---|---|
| Plugin | `eventstorecloud` | `kurrentcloud` |
| Resource type tokens | `eventstorecloud:index:Project`, … | `kurrentcloud:index:Project`, … |
| npm package | `@eventstore/pulumi-eventstorecloud` | `@kurrent-io/pulumi-kurrentcloud` |
| Python package | `pulumi_eventstorecloud` | `pulumi_kurrentcloud` |
| Go SDK package | `.../sdk/go/eventstorecloud` | `.../sdk/go/kurrentcloud` |
| .NET package / namespace | `Pulumi.EventStoreCloud` | `Pulumi.KurrentCloud` |
| Config namespace | `eventstorecloud:token`, … | `kurrentcloud:token`, … |

The provider configuration **keys** (`token`, `organizationId`, …) and the `ESC_*` environment
variables are unchanged. The new `ManagedClusterReplicaset` resource (read-only replica sets) is now
available in every language.

## Why it is non-destructive

Each renamed resource declares a Pulumi [alias](https://www.pulumi.com/docs/concepts/options/aliases/)
to its historical `eventstorecloud:index:*` type. When you upgrade and run `pulumi up`, the engine
recognises the existing state objects as the same resources under their new type and simply updates
the type in state — it does **not** destroy and recreate the underlying Kurrent Cloud infrastructure.

## Upgrade steps

### 1. Install the plugin (for Go / .NET projects)

```bash
pulumi plugin install resource kurrentcloud --server github://api.github.com/kurrent-io
```

### 2. Swap the SDK package

**TypeScript / JavaScript**
```bash
npm uninstall @eventstore/pulumi-eventstorecloud
npm install @kurrent-io/pulumi-kurrentcloud
```
```diff
- import * as esc from "@eventstore/pulumi-eventstorecloud";
+ import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";
```

**Python**
```bash
pip uninstall pulumi_eventstorecloud && pip install pulumi_kurrentcloud
```
```diff
- import pulumi_eventstorecloud as esc
+ import pulumi_kurrentcloud as kurrent
```

**Go**
```diff
- esc "github.com/EventStore/pulumi-eventstorecloud/sdk/go/eventstorecloud"
+ kurrent "github.com/EventStore/pulumi-eventstorecloud/sdk/go/kurrentcloud"
```

**.NET**
```bash
dotnet remove package Pulumi.EventStoreCloud
dotnet add package Pulumi.KurrentCloud
```

### 3. Update config namespace (if you use `pulumi config`)

```bash
pulumi config set kurrentcloud:organizationId <org-id>
pulumi config set --secret kurrentcloud:token <token>
# remove the old keys once migrated
pulumi config rm eventstorecloud:organizationId
pulumi config rm eventstorecloud:token
```

### 4. Preview, then apply

```bash
pulumi preview   # MUST show "0 to replace" — only type/in-place updates
pulumi up
```

If `pulumi preview` shows any resource being **replaced** or **deleted**, stop and open an issue —
the aliases should make this a no-op migration.

## Rollback

The previous `@eventstore/pulumi-eventstorecloud` package and `eventstorecloud` plugin continue to
work. If you need to roll back before running `pulumi up`, restore the old import/package and your
state is untouched. (After a successful `pulumi up` the state holds the new `kurrentcloud:*` types;
rolling back then would itself be an alias-style migration in reverse.)
