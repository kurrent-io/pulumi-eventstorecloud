# coding=utf-8
# *** WARNING: this file was generated by the Pulumi Terraform Bridge (tfgen) Tool. ***
# *** Do not edit by hand unless you're certain you know what you are doing! ***

from . import _utilities
import typing
# Export this package's modules as members:
from .acl import *
from .aws_cloud_watch_logs_integration import *
from .aws_cloud_watch_metrics_integration import *
from .get_network import *
from .get_project import *
from .integration import *
from .managed_cluster import *
from .network import *
from .peering import *
from .project import *
from .provider import *
from .scheduled_backup import *

# Make subpackages available:
if typing.TYPE_CHECKING:
    import pulumi_eventstorecloud.config as __config
    config = __config
else:
    config = _utilities.lazy_import('pulumi_eventstorecloud.config')

_utilities.register(
    resource_modules="""
[
 {
  "pkg": "eventstorecloud",
  "mod": "index/aWSCloudWatchLogsIntegration",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/aWSCloudWatchLogsIntegration:AWSCloudWatchLogsIntegration": "AWSCloudWatchLogsIntegration"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/aWSCloudWatchMetricsIntegration",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/aWSCloudWatchMetricsIntegration:AWSCloudWatchMetricsIntegration": "AWSCloudWatchMetricsIntegration"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/acl",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/acl:Acl": "Acl"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/integration",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/integration:Integration": "Integration"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/managedCluster",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/managedCluster:ManagedCluster": "ManagedCluster"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/network",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/network:Network": "Network"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/peering",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/peering:Peering": "Peering"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/project",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/project:Project": "Project"
  }
 },
 {
  "pkg": "eventstorecloud",
  "mod": "index/scheduledBackup",
  "fqn": "pulumi_eventstorecloud",
  "classes": {
   "eventstorecloud:index/scheduledBackup:ScheduledBackup": "ScheduledBackup"
  }
 }
]
""",
    resource_packages="""
[
 {
  "pkg": "eventstorecloud",
  "token": "pulumi:providers:eventstorecloud",
  "fqn": "pulumi_eventstorecloud",
  "class": "Provider"
 }
]
"""
)
