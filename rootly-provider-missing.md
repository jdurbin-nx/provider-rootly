# provider-rootly — resources missing for alert → incident → status-page

Goal: drive a **Rootly status page** from a Prometheus/Alertmanager alert, fully git/Crossplane-managed.
Flow needed: alert hits the alertmanager **AlertsSource** (routably, not `/notify/<type>/<id>`) → an
**Alert Route rule** matches it → pages SRE **and** attaches a **Service** / creates an **Incident** →
a **Status Page Component** mapped to that Service reflects status → (optional) a **Workflow**
auto-publishes the update.

## What the installed provider already has
`teams`, `escalationpolicies`, `escalationlevels`, `schedules`, `schedulerotations`,
`schedulerotationusers`, `heartbeats`, `alertssources`, `providerconfig(s)`.
(On-call / alerting primitives only.)

## Missing — REQUIRED for this feature
Mapped to the Rootly API objects (all confirmed present in the account via the API):

1. **Service** — `rootly_service` / API `services`. The pivot object: alert routes target it,
   incidents attach to it, status-page components map to it. (13 services already exist in the
   account, e.g. "Authorization server (aka cloud_db)" — we just can't reference/manage them.)

2. **Alert Route + Alert Route Rules** — API `alert_routes` + nested `alert_route_rules`.
   A route binds to one or more `alertssources`; each rule has a condition (`condition_type`,
   field matchers) and a `destination` (`target_type` = Group | Team | Service | EscalationPolicy,
   `target_id`). This is the "route on Rootly's side" piece. The AlertsSource CRD exists but there
   is **no route/rule CRD**, so routing can't be expressed in git today.

3. **Workflow + Workflow Tasks** — `rootly_workflow` (+ task resources). Genres `alert` / `incident`.
   Needed tasks: **create_incident** (turn a matched alert into an incident with the Service
   attached) and **publish status-page update** (auto-post, if we don't want manual publishing).
   Without this, incident creation and auto-publish can't be git-managed.

4. **Status Page** — `rootly_status_page` / API `status_pages`. The page itself (title, public vs
   internal, custom domain, etc.). **Zero exist in the account** — this creates the first one.

5. **Status Page Component (component → service/functionality mapping)** — the items shown under
   "System Status". Each maps to a Service, Functionality, or third-party service. In the Rootly TF
   provider this is a sub-resource of the status page (or a dedicated component resource) — needs
   coverage either way.

## Missing — OPTIONAL / supporting
- **Functionality** — `rootly_functionality` / API `functionalities`. Higher-level customer-facing
  grouping ("Login", "Checkout") if components should map to functionalities rather than raw
  services. (Only 3 Rootly demo defaults exist today.)
- **Incident Type** (`rootly_incident_type`) and **Severity** (`rootly_severity`) — for well-formed
  auto-created incidents (so the workflow's create_incident has a type/severity to set).

## Priority order to unblock the cdb-auth prod → status-page build
`Service` → `AlertRoute` (+rules) → `Workflow` (+create_incident / publish tasks) → `StatusPage` →
`Component`. Functionality / IncidentType / Severity are polish.

## Install / rollout note
All of the above land **after** the new provider version ships and is installed on sre-prod.
Until then: **do not push any Crossplane manifest that references these new kinds** — Flux would
dry-run-fail against the missing CRDs. Hold the Rootly-side manifests (and the paired Alertmanager
"routable feed" change) until the provider is live; the git branch is parked, not merged.
