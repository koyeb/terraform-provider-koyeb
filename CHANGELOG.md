## 0.1.0 (Unreleased)

BACKWARDS INCOMPATIBILITIES / NOTES:

* `koyeb_service`, `koyeb_service_pool` and `koyeb_database` create and update now wait for the resource to become ready (up to 10 minutes for services and databases, 5 minutes for pools) before reporting success. A WEB service without a `ports` block never becomes healthy and will now fail the apply; add ports or use type `WORKER`.
* `koyeb_service_pool_claim` create now waits until the claim is `FULFILLED` and the claimed service is ready (up to 300 seconds each), and fails fast when the claim reports `FAILED` or `RELEASED`.
* Sandbox pools: set `definition.type = "SANDBOX"` explicitly — the provider defaults to `WEB` like the service resource, unlike the Koyeb SDKs which default pool definitions to `SANDBOX`. Services built from the `koyeb/sandbox` image also require a `SANDBOX_SECRET` environment variable, which the SDKs generate for you; supply one via `definition.env` (e.g. from a `random_password` resource).

FEATURES:

* The shared service definition schema now accepts `type = "SANDBOX"`, matching the service pool support of the Koyeb SDKs and CLI.
