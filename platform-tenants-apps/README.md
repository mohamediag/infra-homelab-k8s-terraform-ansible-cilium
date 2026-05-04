# platform-tenants-apps

Tenant-facing GitOps folder for application Claims.

## Layout

- `argocd/applicationset.yaml`: Generates one ArgoCD Application per service/environment folder under `apps/`.
- `apps/<service>/{dev,staging,prod}/app.yaml`: Per-environment App Claim consumed by Crossplane.

## Ownership model

- `gitops/`: platform and infrastructure-level resources managed by platform operators.
- `platform-tenants-apps/`: tenant app intent (Claims) managed through app/team workflows.
