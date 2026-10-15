# provider-rootly

A [Crossplane](https://crossplane.io) provider for [Rootly](https://rootly.com),
generated with [Upjet](https://github.com/crossplane/upjet) from the official
[rootlyhq/terraform-provider-rootly](https://github.com/rootlyhq/terraform-provider-rootly).

Incident-management substrate as managed resources: drive Rootly from git,
alongside the rest of your infrastructure, with the same reconcile loop.

## Managed Resources

The on-call set, and the kinds that take an alert through to a status page:

| Kind | Group | Terraform resource |
|---|---|---|
| `Team` | `team.rootly.crossplane.io` | `rootly_team` |
| `EscalationPolicy` | `escalation.rootly.crossplane.io` | `rootly_escalation_policy` |
| `EscalationLevel` | `escalation.rootly.crossplane.io` | `rootly_escalation_level` |
| `Schedule` | `schedule.rootly.crossplane.io` | `rootly_schedule` |
| `ScheduleRotation` | `schedule.rootly.crossplane.io` | `rootly_schedule_rotation` |
| `ScheduleRotationUser` | `schedule.rootly.crossplane.io` | `rootly_schedule_rotation_user` |
| `Heartbeat` | `heartbeat.rootly.crossplane.io` | `rootly_heartbeat` |
| `AlertsSource` | `alerts.rootly.crossplane.io` | `rootly_alerts_source` |
| `AlertRoute` | `alerts.rootly.crossplane.io` | `rootly_alert_route` |
| `AlertUrgency` | `alerts.rootly.crossplane.io` | `rootly_alert_urgency` |
| `Service` | `service.rootly.crossplane.io` | `rootly_service` |
| `Functionality` | `service.rootly.crossplane.io` | `rootly_functionality` |
| `IncidentType` | `incident.rootly.crossplane.io` | `rootly_incident_type` |
| `Severity` | `incident.rootly.crossplane.io` | `rootly_severity` |
| `WorkflowAlert` | `workflow.rootly.crossplane.io` | `rootly_workflow_alert` |
| `WorkflowIncident` | `workflow.rootly.crossplane.io` | `rootly_workflow_incident` |
| `WorkflowTaskCreateIncident` | `workflow.rootly.crossplane.io` | `rootly_workflow_task_create_incident` |
| `WorkflowTaskPublishIncident` | `workflow.rootly.crossplane.io` | `rootly_workflow_task_publish_incident` |
| `WorkflowTaskPageRootlyOnCallResponders` | `workflow.rootly.crossplane.io` | `rootly_workflow_task_page_rootly_on_call_responders` |
| `StatusPage` | `statuspage.rootly.crossplane.io` | `rootly_status_page` |
| `StatusPageComponentGroup` | `statuspage.rootly.crossplane.io` | `rootly_status_page_component_group` |
| `StatusPageComponent` | `statuspage.rootly.crossplane.io` | `rootly_status_page_component` |

Every kind also exists namespaced under `*.rootly.m.crossplane.io` for
Crossplane v2 namespaced composition. Rootly assigns every resource its ID, so
external names are always learned from the API after creation -- never set one
by hand.

A `ScheduleRotation` takes its members one of two ways. Use one per rotation,
never both -- each would reconcile the other's members away:

- `ScheduleRotationUser`, one per user and position. Upstream deprecates
  `rootly_schedule_rotation_user` and will drop it in its next major version, so
  this kind goes away with that upgrade.
- `scheduleRotationMembers`, inline on the rotation. This is the upstream
  replacement, and it requires the schedule-nesting feature on the Rootly
  account.

An `AlertsSource` takes either `alertTemplateAttributes` or
`alertSourceFieldsAttributes`, never both; the upstream provider rejects a
configuration with both before it talks to Rootly. Neither is
late-initialized, so the spec carries only the one you declare. If an
`AlertsSource` reports `Conflicting configuration arguments`, an earlier
provider release late-initialized the other block into its spec: remove the
block you do not manage from `spec.forProvider` and it reconciles again.

### From An Alert To A Status Page

The kinds compose into one flow, every link by name:

1. An `AlertsSource` receives the alert (for example from Alertmanager).
2. An `AlertRoute` lists that source in `alertsSourceIdsRefs`; each rule
   matches on alert fields and sends to a destination, a `Service` named in
   `targetIdRef`.
3. A `WorkflowAlert` fires on `alert_created`. Its steps are separate kinds that
   name it in `workflowIdRef`: a `WorkflowTaskCreateIncident` opens the
   incident against the `Service`, with a `Severity` and `IncidentType`, and
   a `WorkflowTaskPageRootlyOnCallResponders` pages the on-call responders of
   a target at an `AlertUrgency`.
4. A `WorkflowIncident` fires on `incident_created` for that `Service`. Its
   `WorkflowTaskPublishIncident` posts the incident to the `StatusPage`.
   Publishing acts on the incident, and only an incident-triggered workflow
   runs its steps with the incident in scope: in an alert workflow the same
   step runs and posts nothing, even after a create-incident step.
5. A `StatusPageComponent` mirrors the `Service` on the page, so what the
   incident touches is what the page shows.

`examples/cluster/` holds one manifest per kind, written as that flow.

Four fields point at a kind chosen by a sibling type field, and a reference
can only resolve one kind. Each resolves the kind the flow above needs; the
others take the Rootly id directly:

| Field | Type field | Resolved by reference | Raw id for |
|---|---|---|---|
| `AlertRoute` `rules[].destinations[].targetId` | `targetType` | `Service` | `Group`, `Functionality`, `EscalationPolicy` |
| `StatusPageComponent` `sourceId` | `sourceType` | `Service` | `Functionality` |
| `WorkflowTaskCreateIncident`, `WorkflowTaskPageRootlyOnCallResponders` `workflowId` | the workflow's genre | `WorkflowAlert` | `WorkflowIncident` |
| `WorkflowTaskPublishIncident` `workflowId` | the workflow's genre | `WorkflowIncident` | `WorkflowAlert` |

A `WorkflowTaskPageRootlyOnCallResponders` names who it pages in one of five
maps, `serviceTarget`, `groupTarget`, `functionalityTarget`,
`escalationPolicyTarget` or `userTarget`, each holding the target's Rootly
`id` and `name`. A reference resolves a single id field, not a map, so these
take the id as written.

A `StatusPage` with `authenticationEnabled` takes its password from a secret
through `authenticationPasswordSecretRef`; upstream does not flag the field as
sensitive, this provider does, so it never appears in the spec or the status.

The upstream provider covers ~229 resources; widening the scope is one entry in
`config/external_name.go` and a regenerate.

## Install

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-rootly
spec:
  package: ghcr.io/nikogura/provider-rootly:v0.1.0
```

Pin by digest in anything you operate; tags are for humans.

Every green merge to main is a release: CI computes the next semantic version
from conventional commits, publishes the multi-arch image and package to ghcr
under that tag, and cuts the matching GitHub Release. Nothing is tagged by hand.

## Authenticate

One bearer token. The credentials Secret holds a JSON document:

```bash
kubectl create secret generic rootly-creds -n crossplane-system --from-literal=credentials='{"api_token":"<ROOTLY_API_TOKEN>"}'
```

```yaml
apiVersion: rootly.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  credentials:
    source: Secret
    secretRef:
      name: rootly-creds
      namespace: crossplane-system
      key: credentials
```

`api_host` is an optional second key, only for self-hosted or regional
endpoints; it defaults to `https://api.rootly.com`.

## Adopting An Existing Rootly Organization

Start observe-only: create MRs with
`managementPolicies: ["Observe"]` and the `crossplane.io/external-name`
annotation set to the resource's Rootly ID, watch them all come up Synced and
Ready against reality, and only then widen the policies. A wrong external name
does not fail -- full management would build a duplicate beside the real thing
and report success.

## Developing

```bash
make submodules   # once, after clone
make generate     # regenerate from config/schema.json + upstream docs
make build        # binary + package
```

The generation inputs are pinned in the Makefile
(`TERRAFORM_PROVIDER_VERSION`); bumping the upstream provider version is a
version bump plus `make generate`, then read the diff.

## License

Apache-2.0. At runtime the provider drives the MPL-2.0
`terraform-provider-rootly` binary, which is downloaded unmodified at image
build; see NOTICE.
