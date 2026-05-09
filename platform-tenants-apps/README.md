# platform-tenants-apps

Tenant-facing GitOps folder for application **App XRs** (Crossplane v2 namespaced composite resources).

## Layout

- `argocd/applicationset.yaml`: Generates one ArgoCD Application per service/environment folder under `apps/<team>/<service>/<env>/`.
- `apps/<team>/<service>/{dev,staging,prod}/app.yaml`: Per-environment `App` XR consumed by Crossplane.

## Ownership model

- `gitops/`: platform and infrastructure-level resources managed by platform operators.
- `platform-tenants-apps/`: tenant app intent (App XRs) managed through app/team workflows.

## Namespace model

Each `App` XR declares `metadata.namespace: <team>-<env>` (e.g. `team-platform-dev`). The XR **and** all its composed Kubernetes resources (Deployment, Service, ConfigMap, ExternalSecret, HTTPRoute, plus the wrapping provider-kubernetes `Object` MRs) all land in that single team-env namespace.

The ApplicationSet derives the destination namespace from the folder path (`<team>-<env>`) and ArgoCD's `CreateNamespace=true` syncOption mints it on first sync. No separate Namespace YAML is required.

## Authoring an App XR (manual)

```yaml
apiVersion: platform.homelab.io/v1alpha1
kind: App
metadata:
  name: my-service              # one XR per env; uniqueness comes from namespace
  namespace: team-backend-dev   # team-env namespace (must exist)
spec:
  compositionRef:
    name: app-kubernetes
  parameters:
    name: my-service
    owner: team-backend
    environment: dev
    image: ghcr.io/example/my-service:abc123
    port: 8080
    replicas: 1
    resources:
      requests: { cpu: 25m, memory: 32Mi }
      limits:   { cpu: 100m, memory: 64Mi }
    config:
      LOG_LEVEL: debug
    secrets: []                 # names → ExternalSecret @ <name>/<env>/<SECRET>
    exposure:
      type: none                # public | private | none
      host: my-service.example.com
    healthCheck:
      liveness: /healthz
      readiness: /readyz
    metrics:
      enabled: true
      path: /metrics
```

In production, the generator (`tools/generator/`, Step 5) writes these files from a developer-authored `service.yaml` in the app repo.

## Debugging

```sh
# See the XR and all composed resources
crossplane beta trace app/<name> -n <team-env-ns>

# Inspect the wrapping Object MRs
kubectl get objects.kubernetes.m.crossplane.io -n <team-env-ns>

# Inspect the actual workload
kubectl get all -n <team-env-ns>
```
