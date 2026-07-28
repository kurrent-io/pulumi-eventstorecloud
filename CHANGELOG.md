CHANGELOG
=========

## 1.0.0 (unreleased)

**Rebranded from `eventstorecloud` to `kurrentcloud`** to match the Kurrent Cloud Terraform provider.

- **Service Account authentication**: new optional `clientSecret` (secret) and `identityKitUrl`
  provider config. When both `clientId` and `clientSecret` are set (or `ESC_CLIENT_ID` /
  `ESC_CLIENT_SECRET` via the environment), the provider authenticates with the OAuth2
  client-credentials grant instead of the `token` refresh-token flow. Inherited from the upstream
  Terraform provider (`features/pkce-and-sa-auth`).
- Re-pointed the Terraform bridge to `kurrent-io/terraform-provider-kurrentcloud` v2.1.1 (was
  `EventStore/terraform-provider-eventstorecloud` v1.6.0).
- **New resource:** `ManagedClusterReplicaset` — read-only replica sets attached to a managed cluster.
- Renamed the Pulumi package, resource tokens, namespaces, and SDK packages to `kurrentcloud` /
  `@kurrent-io/pulumi-kurrentcloud` / `pulumi_kurrentcloud` / `Pulumi.KurrentCloud`. Existing stacks
  migrate without resource replacement via Pulumi aliases — see [MIGRATION.md](./MIGRATION.md). (breaking)
- Inherited the upstream in-place `projectionLevel` update behavior (no longer forces cluster replacement).

## 0.1.2 (Initial release)

- First release using the Terraform bridge and ESC Terraform provider
- SDK packages are on GitHub package registry, check `README` for instructions

## 0.2.0

- Update to TF provider 1.5.9

## 0.3.0

- Renamed the namespace to `EventStoreCloud` (breaking)

---
