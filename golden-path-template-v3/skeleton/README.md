# Generated Service

This service is managed by the IDP CI workflow.

## Required Repository Secret

- `GITOPS_TOKEN`: PAT with permission to push branches and open pull requests in `mohamediag/infra-homelab-k8s-terraform-ansible-cilium`.

This is the v1 auth model for speed. Replace it later with a GitHub App installation token for tighter permissions and easier rotation.

## Promotion

Normal pushes update dev only. To promote a validated image to staging or prod, run the `promote` workflow manually with:

- `environment`: `staging` or `prod`
- `image`: full image ref, for example `ghcr.io/mohamediag/sample-service:e42bf0d`
