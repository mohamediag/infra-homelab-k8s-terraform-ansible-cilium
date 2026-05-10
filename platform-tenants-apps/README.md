# platform-tenants-apps

Tenant-facing GitOps control folder. It owns the ApplicationSet that generates one ArgoCD Application per application/environment folder.

## Layout

- `platform-tenants-apps/argocd/applicationset.yaml`: Generates one ArgoCD Application per `<env>/<team>/<application>` folder.
- `platform-tenants-apps-<env>/apps/<team>/<application>/namespace.yaml`: Explicit namespace manifest for that application/environment.
- `platform-tenants-apps-<env>/apps/<team>/<application>/<service>.yaml`: One Crossplane v2 namespaced `App` XR per service.

Example:

```text
platform-tenants-apps-dev/
└── apps/
    └── team-platform/
        └── sample-application/
            ├── namespace.yaml
            ├── sample-service.yaml
            └── another-service.yaml
```

## Ownership model

- `gitops/`: platform and infrastructure-level resources managed by platform operators.
- `platform-tenants-apps/`: ArgoCD ApplicationSet owned by platform operators.
- `platform-tenants-apps-{dev,staging,prod}/`: tenant app intent managed through app/team workflows.

## Namespace model

Namespace convention: `<application>-<env>`.

Example: `sample-application-dev`.

Team ownership is stored as namespace/app labels, not encoded in the runtime namespace name.

An application can contain multiple services. All service XRs and their composed workloads live in that one application/environment namespace:

```text
sample-application-dev
├── App/sample-service
├── App/another-service
├── Object MRs
├── Deployment/sample-service
├── Deployment/another-service
├── Service/sample-service
└── Service/another-service
```

The namespace is an explicit manifest in the same folder as the service XRs. ArgoCD sync waves guarantee ordering:

- `namespace.yaml`: `argocd.argoproj.io/sync-wave: "-10"`
- service `App` XRs: `argocd.argoproj.io/sync-wave: "0"`

`CreateNamespace=true` is intentionally not used; the namespace is Git-owned and can carry labels, annotations, quotas, and policies later.

## Authoring a Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: sample-application-dev
  annotations:
    argocd.argoproj.io/sync-wave: "-10"
  labels:
    app.platform.homelab.io/team: team-platform
    app.platform.homelab.io/application: sample-application
    app.platform.homelab.io/environment: dev
```

## Authoring an App XR

```yaml
apiVersion: platform.homelab.io/v1alpha1
kind: App
metadata:
  name: sample-service
  namespace: sample-application-dev
  annotations:
    argocd.argoproj.io/sync-wave: "0"
spec:
  crossplane:
    compositionRef:
      name: app-kubernetes
  parameters:
    name: sample-service
    owner: team-platform
    environment: dev
    image: ghcr.io/example/sample-service:abc123
    port: 8080
    replicas: 1
    resources:
      requests: { cpu: 25m, memory: 32Mi }
      limits: { cpu: 100m, memory: 64Mi }
    config:
      LOG_LEVEL: debug
    secrets: []
    exposure:
      type: none
      host: sample-service.example.com
    healthCheck:
      liveness: /healthz
      readiness: /readyz
    metrics:
      enabled: true
      path: /metrics
```

In production, the generator (`tools/generator/`, Step 5) writes these files from developer-authored service/application metadata in app repos.

## Debugging

```sh
# See the XR and all composed resources
crossplane beta trace app/<service-name> -n <application>-<env>

# Inspect the wrapping Object MRs
kubectl get objects.kubernetes.m.crossplane.io -n <application>-<env>

# Inspect the actual workload
kubectl get all -n <application>-<env>
```
