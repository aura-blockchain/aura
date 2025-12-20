# Vault Dev Deployment

## Overview
HashiCorp Vault deployed in dev mode for testing network policies and secrets management.

## Deployed Resources
- **Namespace**: `vault`
- **Service**: `vault.vault.svc.cluster.local:8200`
- **Pod**: 1 replica, running as non-root user (UID 100)

## Security Context
- RunAsNonRoot: true (UID 100, fsGroup 1000)
- Drop all capabilities
- SeccompProfile: RuntimeDefault
- No privilege escalation

## Access
```bash
# Check status
kubectl exec -n vault deployment/vault -- vault status

# Health check
curl http://vault.vault.svc.cluster.local:8200/v1/sys/health

# From pod IP
VAULT_IP=$(kubectl get pod -n vault -l app=vault -o jsonpath='{.items[0].status.podIP}')
curl http://$VAULT_IP:8200/v1/sys/health
```

## Dev Mode Settings
- Root token: `root`
- In-memory storage (non-persistent)
- Unsealed by default

## Network Policy Testing
The deployment enables the `test_vault_access()` function in:
`k8s/network-policies/test-network-policies.sh`

## Deployment
```bash
kubectl apply -f k8s/vault-dev.yaml
```

**WARNING**: This is a dev-mode deployment. NOT for production use.
