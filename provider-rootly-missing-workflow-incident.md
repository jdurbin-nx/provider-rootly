# provider-rootly: add WorkflowIncident (wrap `rootly_workflow_incident`) — needed for SRE-2268

## What's missing

The provider (v0.4.0) wraps `rootly_workflow_alert` (kind `WorkflowAlert`,
group `workflow.rootly.crossplane.io`) but NOT `rootly_workflow_incident` —
the incident-triggered workflow type. Upstream resource in the rootlyhq/rootly
Terraform provider (5.20.1, same version already wrapped):

- TF resource: `rootly_workflow_incident`
- Docs: https://registry.terraform.io/providers/rootlyhq/rootly/latest/docs/resources/workflow_incident

Wrap it exactly like `workflow_alert` was wrapped: same group
(`workflow.rootly.crossplane.io/v1alpha1`), kind `WorkflowIncident`, same
external-name/config pattern. No new task kinds are needed — the existing
`WorkflowTaskPublishIncident` (and all other task kinds) attach to any
workflow via `workflow_id`, which already works.

## Why it's needed (observed, 2026-10-15)

A `publish_incident` task inside an **alert** workflow silently no-ops: its
`incident: { id: "{{ incident.id }}", name: "{{ incident.title }}" }` liquid
renders empty because alert workflows execute with **alert** context only —
the incident created by a preceding `create_incident` task in the same
workflow is NOT bound to `{{ incident }}` for later tasks. Verified live:
task config fully persisted and correct in Rootly (status_page_id, 
selected_component_keys, selected_component_statuses all stored), workflow ran
(page + create_incident tasks succeeded, incident created), publish produced
nothing — no status-page event, component status unchanged (null).

Rootly's model: alert workflow declares the incident; a separate
**incident-triggered** workflow (trigger `incident_created`, where
`{{ incident }}` IS bound) does publish/status-page work. That workflow type
is what the provider can't express today.

## Fields the consumer needs (minimum)

From the TF schema for `workflow_incident` — the usual workflow envelope:

- `name`, `description`, `enabled`, `position`
- `trigger_params` block:
  - `triggers` (list; will use `["incident_created"]`; also fires on
    `incident_updated` etc.)
  - `incident_condition` (ALL/ANY)
  - incident condition fields + id lists, at minimum the service pair:
    `incident_condition_service` (IS/ANY/CONTAINS/...) + `service_ids`
    (also in schema: environment/severity/kind/status/visibility pairs —
    include whatever upjet generates, no need to trim)

Intended consumer manifest once it lands (for scale):

```yaml
apiVersion: workflow.rootly.crossplane.io/v1alpha1
kind: WorkflowIncident
metadata:
  name: cdb-status-page
spec:
  managementPolicies: ["*"]
  deletionPolicy: Orphan
  forProvider:
    name: "cdb incident -> internal status page"
    enabled: true
    triggerParams:
      - triggers: ["incident_created"]
        incidentConditionService: "IS"
        serviceIds: ["db2203ec-c76f-41a4-b9a7-6adf1feebf54"]
  providerConfigRef:
    name: rootly
```

…then the existing `WorkflowTaskPublishIncident` CR just repoints its
`workflowId` at this workflow's id.

## Gotchas from wrapping the previous kinds (apply here too)

- Flat `trigger_params` fields persist fine; `alert_payload_conditions` on
  the alert kind did NOT (write path drops it — separate known gap, not
  needed for this work).
- Cross-resource reference resolvers have burned us twice (publish task's
  `workflowIdRef` SSA patch rejected: "taskParams is a required parameter";
  AlertRoute `targetIdRef` resolves against the Service kind only). Literal
  IDs are the working pattern — refs are nice-to-have, don't block on them.
- Condition enum strings are `IS`, `IS NOT`, `ANY`, `CONTAINS`,
  `CONTAINS_ALL`, `CONTAINS_NONE`, `NONE`, `SET`, `UNSET` on the alert kind;
  expect the same upstream enums on incident conditions.
