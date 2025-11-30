# Technical Gaps Report

This report catalogs integration and architecture gaps identified during an end-to-end scan of the backend, frontend, ML service, database, and deployment artifacts. Entries are grouped by type and reference specific locations.

---

### ML proxy only partially implemented (forecast 501)
**Location:** `internal/api/handlers/ml_proxy.go` (Forecast), `internal/api/routes/routes.go` (ML routes)
**Type:** integration gap
**Details:** The anomaly proxy is implemented and working against `/detect-anomalies`, but the forecast proxy returns 501. Frontend or docs expecting cost forecasting will fail.
**Suggested Fix:** Implement a forecast mapping. Either: (A) add `/api/v1/ml/forecast` in the FastAPI service with expected schema; or (B) transform the forecast request to call an existing ML route (`/predict-data-transfer`) and adapt response to documented model.

### Missing resource details endpoint vs docs
**Location:** `internal/api/routes/routes.go`, `internal/api/handlers/resources.go`, `API_DOCUMENTATION.md` (Get Resource Details)
**Type:** model/API contract mismatch
**Details:** Docs specify `GET /api/v1/resources/{resource_id}`, but router only exposes list and stats. No detail handler registered/implemented.
**Suggested Fix:** Add `GET /api/v1/resources/{resource_id}` handler returning full resource details and utilization history, or update docs to remove the endpoint until implemented.

### Onboarding UI vs backend routes partially duplicated
**Location:** `frontend/src/pages/Onboarding.tsx` (uses `/api/onboarding/*`), `internal/api/routes/routes.go` (onboarding routes)
**Type:** integration gap
**Details:** Onboarding routes are now registered, but SimpleOnboarding page still posts to `/api/customers` directly; `Onboarding.tsx` calls `/api/onboarding/*` endpoints. Ensure only one UX path is supported and fully wired.
**Suggested Fix:** Pick the `/api/onboarding/*` flow as canonical. Update `SimpleOnboarding.tsx` to use onboarding endpoints, or remove the duplicate page.

### Dynamic filters rely on current page data, not CMDB/tag datasets
**Location:** `frontend/src/pages/HiddenCosts.tsx` (filter source), backend lacks tags/categories endpoint
**Type:** UX issue / missing feature
**Details:** Filters are populated by mapping categories from the current findings list. This is not Datadog-like (should be global, derived from resource/tag datasets). No API exists to fetch distinct categories/tags per tenant.
**Suggested Fix:** Add `GET /api/filters?tenant_id=...` returning distinct values (categories, severities, services, tags). Update UI to fetch and populate filters on mount.

### No explicit API to list tags/CMDB dimensions
**Location:** `internal/api/handlers/*` (absent), `internal/services/*` (no tag exposure), DB schema not queried for tags
**Type:** missing feature
**Details:** Tag-based browsing requires an endpoint to enumerate tags (e.g., `env`, `team`, `project`) from resource inventory. Absent right now.
**Suggested Fix:** Add `GET /api/v1/tags?tenant_id=...` returning `{ key, values[] }`. Back it by an indexed table or a view over resources with tag JSON.

### Findings pagination missing
**Location:** `internal/api/handlers/customers.go` (GetFindings)
**Type:** performance issue
**Details:** The findings endpoint returns all rows for a tenant and allows filtering by category/severity, but provides no pagination. Large tenants will return big payloads.
**Suggested Fix:** Add `page`, `per_page` query params with `LIMIT/OFFSET`, and return a `meta` block with totals.

### Admin header expectations vs client behavior
**Location:** `frontend/src/pages/AdminDashboard.tsx`, `frontend/src/pages/AuditLogs.tsx`, `internal/api/middleware` (admin auth)
**Type:** integration gap (now mitigated)
**Details:** Admin endpoints require `X-Admin-Key` and `X-Admin-User`. The API client was updated to pass headers, but any other admin calls must follow suit.
**Suggested Fix:** Centralize admin headers through the API client for all admin routes and add dev configuration for these values.

### TESTING_GUIDE references `/api/scan` but missing route
**Location:** `TESTING_GUIDE.md` (Trigger Manual Scan), `internal/api/routes/routes.go` (absent)
**Type:** missing feature / doc mismatch
**Details:** Guide suggests `POST /api/scan` to trigger a scan. No such route is implemented.
**Suggested Fix:** Implement `POST /api/scan` that kicks off a background job (resource sync + detectors run) for a tenant, or remove from docs.

### Port inconsistencies across docs
**Location:** `API_DOCUMENTATION.md` (8080 corrected), `DEPLOYMENT_GUIDE.md` (8090/8091), `QUICK_START.md` (8085/8081), `docker-compose.yml` (8080/8000)
**Type:** documentation inconsistency
**Details:** Multiple ports documented for API/ML. Compose uses 8080/8000.
**Suggested Fix:** Standardize: Local API 8080, ML 8000 across all docs. Note alternative ports for k8s manifests where applicable.

### CORS not environment-driven
**Location:** `internal/api/server.go` (CORS setup)
**Type:** configuration gap
**Details:** CORS allows only `http://localhost:3000`. For staging/prod, origin needs to be configurable.
**Suggested Fix:** Read allowed origins from env `CORS_ALLOWED_ORIGINS` (comma-separated); fallback to localhost in dev.

### Resource utilization retrieval not surfaced via API
**Location:** `internal/plugins/aws/services/*` (utilization logic), `internal/api/handlers/*` (no endpoints), `frontend/src/hooks/useLiveData.ts` (points to non-existent `/api/live`)
**Type:** integration gap
**Details:** Hooks attempt to fetch live utilization from `/api/live/...` which doesn't exist. There is no API exposing CloudWatch/metrics utilization time series.
**Suggested Fix:** Add `/api/v1/resources/{id}/metrics` and `/api/v1/resources/{id}/cost` that return utilization/cost series; retire unused `useLiveData` hooks or rewire them.

### Missing resource drill-down navigation wiring
**Location:** `frontend/src/components/ResourceInventory.tsx`, `frontend/src/components/ResourceDetails.tsx` (present), `frontend/src/App.tsx` (routing)
**Type:** UI integration gap
**Details:** Components for inventory and detail exist but are not clearly wired into routes with stable URLs (`/resources`, `/resources/:id`).
**Suggested Fix:** Add routes and links from dashboards/tables to resource detail pages; ensure deep-linking works.

### Pricing sync and RI/SP optimizer not exposed as jobs/endpoints
**Location:** `internal/services/aws_pricing_service.go`, `internal/aws/cost_explorer.go`, `internal/services/pricing.go` (present), router (absent)
**Type:** integration gap
**Details:** Pricing and RI/SP logic exist but there’s no endpoint/cron wiring to trigger syncs or refresh pricing data.
**Suggested Fix:** Add `POST /api/admin/sync/pricing` and `POST /api/admin/sync/inventory` (admin-only) and document a cron or background worker.

### ML client paths are inconsistent with FastAPI
**Location:** `internal/ml/client.go` (expects `/api/v1/ml/*`), `ml-service/api_ml.py` (exposes `/detect-anomalies`, `/predict-data-transfer`)
**Type:** integration gap (partially mitigated by proxy)
**Details:** Without the proxy, direct client usage would fail. Ensure all server-side ML usage routes through the proxy or update client paths.
**Suggested Fix:** Use proxy exclusively on server-side. Remove/disable direct client paths or align paths in the client to proxy endpoints.

### Security posture: wildcard headers
**Location:** `internal/api/server.go` (AllowedHeaders: `*`)
**Type:** security issue (minor in dev)
**Details:** Allowing all headers is convenient for dev but should be restricted in production.
**Suggested Fix:** Limit headers to known set in production via env toggle, e.g., `Authorization, Content-Type, X-Admin-Key, X-Admin-User`.

### Lack of pagination on admin customers
**Location:** `internal/api/handlers/admin.go` (GetCustomers)
**Type:** performance issue
**Details:** Returns all customers at once; will not scale.
**Suggested Fix:** Add pagination and optional sorting/search server-side.

### Missing error normalization for frontend
**Location:** Multiple handlers return different error shapes
**Type:** UX consistency issue
**Details:** Some handlers return `{ error: string }`, others `{ success: false, error: string }`. Frontend expects consistent `success` flag.
**Suggested Fix:** Standardize response envelope with `success`, `data`, `error`, `meta` across all handlers.

### Incomplete whitelist REST semantics
**Location:** `internal/api/handlers/whitelists.go`
**Type:** API design nuance
**Details:** Revoke uses `DELETE /api/whitelists?id=...` via query param. Path param is more idiomatic.
**Suggested Fix:** Support `DELETE /api/whitelists/{id}` while preserving query for backward compatibility.

### Seed/test data and CMDB exposure
**Location:** `scripts/seed_data.sql`, backend handlers
**Type:** missing feature
**Details:** CMDB-like inventory and tags not exposed via APIs for UI filter population.
**Suggested Fix:** Add `GET /api/v1/inventory/*` and `GET /api/v1/tags` endpoints with efficient queries and indexes.

---

## Larger Design Recommendations

1) Filters and Tag Service
- Add a dedicated filter service that queries distinct dimensions per tenant and caches results.
- API: `GET /api/filters?tenant_id=...` → `{ categories[], severities[], services[], tags: { key, values[] } }`.

2) Utilization and Cost Timeseries APIs
- For each resource, expose `/metrics` and `/cost` series (p95 latency <500ms). Back by pre-aggregated tables or views.

3) Scan Orchestration
- Implement `/api/scan` to orchestrate: inventory sync → utilization fetch → detectors run → store findings → emit metrics.

4) Documentation Harmonization
- Normalize ports across all docs (8080 for API, 8000 for ML in local compose). Document env overrides.

5) Admin Sync Endpoints and Cron
- Add admin-only endpoints to trigger pricing/inventory sync and describe background schedules (every 6 hours) in docs.

---

## Straightforward (<20 lines) Fix Candidates
- Add pagination parameters to `GetFindings` and `admin.GetCustomers`.
- Add `CORS_ALLOWED_ORIGINS` env parsing in `server.go`.
- Add ML proxy forecast handler once the ML path is finalized.
- Add `/api/scan` stub that enqueues a background task (no-op initially) to satisfy docs.

## Larger Fix Candidates
- New filters/tags endpoints with DB support and indexes.
- Resource details/metrics/cost timeseries endpoints and UI drill-down wiring.
- Pricing/inventory sync orchestration endpoints and cron.
- Error response normalization across all handlers.

---

Status: Initial scan complete. Awaiting approval to proceed with quick fixes and create follow-up design docs/implementations for larger items.




