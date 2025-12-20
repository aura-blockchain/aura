# Linkerd Service Mesh Installation

## Installation Summary

Linkerd has been successfully installed on the KinD cluster for Aura blockchain with the following components:

### Components Installed

1. **Gateway API CRDs** (v1.4.0) - Required dependency
2. **Linkerd CRDs** - Custom resource definitions for Linkerd policies
3. **Linkerd Control Plane** - Core service mesh infrastructure
4. **Linkerd Viz Extension** - Observability and monitoring dashboard

### Control Plane Components

- **linkerd-identity**: Certificate authority for mTLS
- **linkerd-destination**: Service discovery and routing
- **linkerd-proxy-injector**: Automatic sidecar injection webhook

### Viz Extension Components

- **metrics-api**: Metrics aggregation
- **prometheus**: Time-series database
- **tap**: Real-time request inspection
- **web**: Linkerd dashboard UI

## Namespace Configuration

The `aura` namespace has been annotated for automatic sidecar injection:

```yaml
annotations:
  linkerd.io/inject: enabled
```

## Restarting Pods for Injection

Existing pods need to be restarted to get the Linkerd sidecar injected:

```bash
# Restart StatefulSet pods
export KUBECONFIG=/tmp/kind-aura.kubeconfig
kubectl rollout restart statefulset/aura-validator -n aura

# Restart individual pods
kubectl delete pod aura-validator-0 -n aura
kubectl delete pod aura-validator-1 -n aura
kubectl delete pod aura-validator-2 -n aura
```

After restart, pods will have 2 containers (main + linkerd-proxy).

## Verification Commands

```bash
# Check installation
linkerd check

# View mesh status
linkerd viz stat deployment -n aura

# View pod injection status
kubectl get pods -n aura -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .spec.containers[*]}{.name}{" "}{end}{"\n"}{end}'

# Access dashboard (port-forward)
linkerd viz dashboard &
```

## Features Enabled

- **Automatic mTLS**: All pod-to-pod communication encrypted
- **Traffic Metrics**: Golden metrics (success rate, latency, throughput)
- **Service Profiles**: Advanced routing and retry policies
- **Tap**: Live traffic inspection
- **Authorization Policies**: Fine-grained access control

## Configuration Files

Linkerd uses the following key resources:

- Control plane: `linkerd` namespace
- Viz extension: `linkerd-viz` namespace
- Config: `linkerd-config` ConfigMap
- Trust anchors: `linkerd-identity-trust-roots` ConfigMap

## Next Steps

1. Restart pods to inject sidecars
2. Verify mTLS with: `linkerd viz edges deployment -n aura`
3. Create ServiceProfiles for advanced routing
4. Set up AuthorizationPolicies for security
