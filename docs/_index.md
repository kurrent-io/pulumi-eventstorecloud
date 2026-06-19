---
title: Kurrent Cloud
meta_desc: Learn how you can use the Kurrent Cloud Provider for Pulumi to provision and manage Kurrent Cloud resources.
layout: package
---

The Kurrent Cloud provider for Pulumi can provision many of the cloud resources available in [Kurrent Cloud](https://www.kurrent.io/kurrent-cloud). It uses the Kurrent Cloud API to manage and provision resources.

The Kurrent Cloud provider must be configured with credentials to deploy and update resources; see [Installation & Configuration](./installation-configuration) for instructions.

## Example

```typescript
import * as kurrent from "@kurrent-io/pulumi-kurrentcloud";

const project = new kurrent.Project("sample-project", {
    name: "Improved Chicken Window",
});

const network = new kurrent.Network("sample-network", {
    name: "Chicken Window Net",
    projectId: project.id,
    resourceProvider: "aws",
    region: "eu-west1",
    cidrBlock: "172.21.0.0/16",
});
```
