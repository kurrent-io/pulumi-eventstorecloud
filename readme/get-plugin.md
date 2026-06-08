### Get the plugin

For projects that use .NET and Go Pulumi SDK you have to install the provider before trying to update the stack.

Use the following command to add the plugin to your environment:

```
pulumi plugin install resource kurrentcloud [version] \
  --server https://github.com/kurrent-io/pulumi-eventstorecloud/releases/download/[version]
```

Example:

```
pulumi plugin install resource kurrentcloud v0.2.3 \
  --server https://github.com/kurrent-io/pulumi-eventstorecloud/releases/download/v0.2.3
```
