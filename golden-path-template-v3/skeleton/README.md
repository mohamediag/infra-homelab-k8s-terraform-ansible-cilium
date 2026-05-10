# Generated Service

This service is managed by the IDP CI workflow.

## Required Repository Secret

- `GITOPS_TOKEN`: PAT with permission to push branches and open pull requests in `mohamediag/infra-homelab-k8s-terraform-ansible-cilium`.

This is the v1 auth model for speed. Replace it later with a GitHub App installation token for tighter permissions and easier rotation.

## Promotion

Normal pushes update dev first. If dev was directly updated, the same pipeline continues into approval-gated promotion jobs:

- `promote-to-staging`: uses the GitHub `staging` environment and defaults to a direct GitOps commit.
- `promote-to-prod`: uses the GitHub `prod` environment and defaults to opening a GitOps PR.

Configure promotion behavior in `service.yaml`:

```yaml
ci:
  autoCommitDev: true
  promotion:
    staging: commit
    prod: pr
```

Allowed promotion values are `commit` and `pr`.
