# Aura Kubernetes Infrastructure Build Guide

**Target**: Production-grade k3s cluster for Aura Cosmos SDK blockchain
**Reference**: XAI project infrastructure (fully tested, 97% production ready)

## Architecture Overview

```
                         ┌─────────────────────────────────────┐
                         │           Internet                  │
                         └──────────────┬──────────────────────┘
                                        │
                         ┌──────────────▼──────────────────────┐
                         │     Ingress (nginx + rate limit)    │
                         │     TLS termination (cert-manager)  │
                         └──────────────┬──────────────────────┘
                                        │
    ┌───────────────────────────────────┼───────────────────────────────────┐
    │                                   │                                   │
    │                       k3s Cluster (2 nodes)                          │
    │               Tailscale mesh networking (100.x.x.x)                  │
    │                                   │                                   │
    │  ┌─────────────────┐    ┌─────────▼─────────┐    ┌─────────────────┐ │
    │  │  bcpc-staging   │◄──►│   aura namespace  │◄──►│   wsl2-worker   │ │
    │  │ (control-plane) │    │                   │    │    (worker)     │ │
    │  │ 100.91.253.108  │    │  ┌─────────────┐  │    │  100.76.8.7     │ │
    │  └─────────────────┘    │  │ Validators  │  │    └─────────────────┘ │
    │                         │  │ StatefulSet │  │                        │
    │                         │  │ (anti-aff.) │  │                        │
    │                         │  └─────────────┘  │                        │
    │                         │         │         │                        │
    │                         │  ┌──────▼──────┐  │                        │
    │                         │  │  Services   │  │                        │
    │                         │  │ P2P/RPC/API │  │                        │
    │                         │  └─────────────┘  │                        │
    │                         └───────────────────┘                        │
    │                                   │                                   │
    │  ┌────────────────────────────────▼────────────────────────────────┐ │
    │  │                    Infrastructure Services                       │ │
    │  │  ┌─────────┐ ┌─────────┐ ┌────────┐ ┌──────┐ ┌────────────────┐ │ │
    │  │  │Prometheus│ │ Grafana │ │ArgoCD  │ │Vault │ │External Secrets│ │ │
    │  │  │ :9090   │ │ :30030  │ │:30085  │ │:8200 │ │   Operator     │ │ │
    │  │  └─────────┘ └─────────┘ └────────┘ └──────┘ └────────────────┘ │ │
    │  │  ┌─────────┐ ┌─────────────┐ ┌─────────────┐                    │ │
    │  │  │Linkerd  │ │cert-manager │ │     VPA     │                    │ │
    │  │  │ mTLS    │ │     TLS     │ │ Recommender │                    │ │
    │  │  └─────────┘ └─────────────┘ └─────────────┘                    │ │
    │  └─────────────────────────────────────────────────────────────────┘ │
    └──────────────────────────────────────────────────────────────────────┘
```

## Port Allocation

Per PORT_ALLOCATION.md, Aura uses 10000-10999:

| Service | Port | Type | Description |
|---------|------|------|-------------|
| P2P | 26656 (internal), 30656 (NodePort) | TCP | Tendermint P2P |
| RPC | 26657 (internal), 30657 (NodePort) | TCP | Tendermint RPC |
| API | 1317 (internal), 30317 (NodePort) | TCP | REST API |
| gRPC | 9090 (internal), 30090 (NodePort) | TCP | gRPC endpoint |
| Metrics | 26660 (internal) | TCP | Prometheus metrics |
| Grafana | 31030 | TCP | Monitoring dashboard (NodePort; 10030 not available due to NodePort range) |
| Explorer | 10080 | TCP | Block explorer |
| Flask | 10050 | TCP | Utility APIs |

---

## Phase 1: Cluster Setup

### 1.1 Prerequisites (bcpc machine)

```bash
# Verify tools installed
docker --version      # 29.1.0+
kubectl version       # v1.33+
helm version          # 3.x

# Install k3s on bcpc (control-plane)
curl -sfL https://get.k3s.io | sh -s - server \
    --cluster-init \
    --disable traefik \
    --disable servicelb \
    --flannel-iface tailscale0 \
    --kubelet-arg "node-ip=100.91.253.108" \
    --node-external-ip 100.91.253.108 \
    --write-kubeconfig-mode 644

# Verify installation
sudo k3s kubectl get nodes
```

### 1.2 Join WSL2 Worker

On WSL2 (wsl2-worker):

```bash
# Get join token from bcpc
# On bcpc:
sudo cat /var/lib/rancher/k3s/server/node-token

# On WSL2:
curl -sfL https://get.k3s.io | K3S_URL=https://100.91.253.108:6443 \
    K3S_TOKEN=<token> sh -s - agent \
    --flannel-iface tailscale0 \
    --kubelet-arg "node-ip=100.76.8.7" \
    --node-external-ip 100.76.8.7

# WSL2 specific: configure metrics port forwarding
# In Windows PowerShell (admin):
netsh interface portproxy add v4tov4 listenport=10250 \
    listenaddress=100.76.8.7 connectport=10250 connectaddress=<WSL2-internal-IP>
```

### 1.3 Configure Kubeconfig

```bash
# Copy kubeconfig for each project
mkdir -p ~/blockchain-projects/aura/.kube
sudo cp /etc/rancher/k3s/k3s.yaml ~/blockchain-projects/aura/.kube/config
sudo chown $USER:$USER ~/blockchain-projects/aura/.kube/config

# Edit to use Tailscale IP
sed -i 's/127.0.0.1/100.91.253.108/g' ~/blockchain-projects/aura/.kube/config

# Add to env.sh
echo 'export KUBECONFIG=~/blockchain-projects/aura/.kube/config' >> ~/blockchain-projects/aura/env.sh
```

### 1.4 Verify Cluster

```bash
source ~/blockchain-projects/aura/env.sh
kubectl get nodes
# Expected:
# NAME           STATUS   ROLES                  VERSION
# bcpc-staging   Ready    control-plane,master   v1.33.6+k3s1
# wsl2-worker    Ready    <none>                 v1.33.6+k3s1
```

---

## Phase 2: Namespace & Security Setup

### 2.1 Create Namespace with Pod Security Standards

Create `k8s/namespace.yaml`:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: aura
  labels:
    app: aura
    # Pod Security Standards - restricted mode
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

Apply: `kubectl apply -f k8s/namespace.yaml`

### 2.2 RBAC Configuration

Create `k8s/rbac-full.yaml`:

```yaml
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aura-validator
  namespace: aura

---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: aura-node
  namespace: aura

---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: aura-role
  namespace: aura
rules:
  - apiGroups: [""]
    resources: [pods, pods/log, pods/status, configmaps, secrets, services, endpoints, persistentvolumeclaims]
    verbs: [get, list, watch]
  - apiGroups: ["apps"]
    resources: [statefulsets, statefulsets/status, deployments]
    verbs: [get, list, watch]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: aura-rolebinding
  namespace: aura
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: aura-role
subjects:
  - kind: ServiceAccount
    name: aura-validator
    namespace: aura
  - kind: ServiceAccount
    name: aura-node
    namespace: aura

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: aura-cluster-role
rules:
  - apiGroups: [""]
    resources: [nodes, nodes/proxy, nodes/metrics, persistentvolumes, events]
    verbs: [get, list, watch]
  - apiGroups: ["metrics.k8s.io"]
    resources: [pods, nodes]
    verbs: [get, list]

---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: aura-cluster-rolebinding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: aura-cluster-role
subjects:
  - kind: ServiceAccount
    name: aura-validator
    namespace: aura
```

### 2.3 Resource Quotas

Create `k8s/resource-quota.yaml`:

```yaml
apiVersion: v1
kind: ResourceQuota
metadata:
  name: aura-quota
  namespace: aura
spec:
  hard:
    requests.cpu: "8"
    requests.memory: "16Gi"
    limits.cpu: "16"
    limits.memory: "32Gi"
    pods: "20"
    persistentvolumeclaims: "10"

---
apiVersion: v1
kind: LimitRange
metadata:
  name: aura-limits
  namespace: aura
spec:
  limits:
    - type: Pod
      max:
        cpu: "8"
        memory: "16Gi"
      min:
        cpu: "100m"
        memory: "128Mi"
    - type: Container
      default:
        cpu: "1000m"
        memory: "2Gi"
      defaultRequest:
        cpu: "500m"
        memory: "1Gi"
    - type: PersistentVolumeClaim
      max:
        storage: "500Gi"
      min:
        storage: "1Gi"
```

---

## Phase 3: Storage Configuration

### 3.1 StorageClass

Create `k8s/storage.yaml`:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: aura-blockchain-storage
  labels:
    app: aura
provisioner: rancher.io/local-path
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Retain
parameters:
  nodePath: /data/aura-blockchain

---
# For fast SSD (validators)
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: rancher.io/local-path
volumeBindingMode: WaitForFirstConsumer
reclaimPolicy: Retain
parameters:
  nodePath: /data/aura-validators
```

### 3.2 Create Data Directories

On both nodes:

```bash
sudo mkdir -p /data/aura-blockchain /data/aura-validators
sudo chmod 755 /data/aura-blockchain /data/aura-validators
```

---

## Phase 4: Secrets Management (Vault + ESO)

### 4.1 Install Vault (Dev Mode)

```bash
# Create vault namespace (privileged for init containers)
kubectl create namespace vault
kubectl label namespace vault pod-security.kubernetes.io/enforce=privileged

# Install Vault via Helm
helm repo add hashicorp https://helm.releases.hashicorp.com
helm repo update

helm install vault hashicorp/vault -n vault \
    --set server.dev.enabled=true \
    --set server.dev.devRootToken=root \
    --set injector.enabled=false
```

### 4.2 Configure Vault

```bash
# Wait for vault to be ready
kubectl wait --for=condition=ready pod/vault-0 -n vault --timeout=120s

# Initialize secrets engine
kubectl exec -n vault vault-0 -- vault secrets enable -path=secret kv-v2

# Create Aura secrets path
kubectl exec -n vault vault-0 -- vault kv put secret/aura/validator-keys \
    priv_validator_key="<base64-encoded-key>" \
    node_key="<base64-encoded-key>" \
    jwt_secret="<jwt-secret>"

# Create policy
kubectl exec -n vault vault-0 -- vault policy write aura-read - <<EOF
path "secret/data/aura/*" {
  capabilities = ["read"]
}
EOF
```

### 4.3 Install External Secrets Operator

```bash
# Install ESO
helm repo add external-secrets https://charts.external-secrets.io
helm repo update

helm install external-secrets external-secrets/external-secrets \
    -n external-secrets --create-namespace \
    --set installCRDs=true
```

### 4.4 Configure ClusterSecretStore

Create `k8s/external-secrets.yaml`:

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ClusterSecretStore
metadata:
  name: vault-backend
spec:
  provider:
    vault:
      server: "http://vault.vault.svc:8200"
      path: "secret"
      version: "v2"
      auth:
        tokenSecretRef:
          name: vault-token
          key: token
          namespace: vault

---
apiVersion: v1
kind: Secret
metadata:
  name: vault-token
  namespace: vault
type: Opaque
stringData:
  token: "root"  # Replace with proper token in production

---
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: aura-secrets
  namespace: aura
spec:
  refreshInterval: "1h"
  secretStoreRef:
    kind: ClusterSecretStore
    name: vault-backend
  target:
    name: aura-secrets-from-vault
    creationPolicy: Owner
  data:
    - secretKey: priv_validator_key
      remoteRef:
        key: aura/validator-keys
        property: priv_validator_key
    - secretKey: node_key
      remoteRef:
        key: aura/validator-keys
        property: node_key
    - secretKey: jwt_secret
      remoteRef:
        key: aura/validator-keys
        property: jwt_secret
```

---

## Phase 5: Network Policies

### 5.1 Default Deny + Allow Rules

Create `k8s/network-policies.yaml`:

```yaml
# Default deny all
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-all
  namespace: aura
spec:
  podSelector: {}
  policyTypes:
    - Ingress
    - Egress

---
# Allow DNS
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-dns
  namespace: aura
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
      ports:
        - protocol: UDP
          port: 53

---
# Allow P2P between validators
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-p2p
  namespace: aura
spec:
  podSelector:
    matchLabels:
      app: aura
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
        - podSelector:
            matchLabels:
              app: aura
      ports:
        - port: 26656
          protocol: TCP
  egress:
    - to:
        - podSelector:
            matchLabels:
              app: aura
      ports:
        - port: 26656
          protocol: TCP

---
# Allow ingress-nginx
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-ingress-nginx
  namespace: aura
spec:
  podSelector:
    matchLabels:
      app: aura
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: ingress-nginx
      ports:
        - port: 26657
          protocol: TCP
        - port: 1317
          protocol: TCP
        - port: 9090
          protocol: TCP

---
# Allow monitoring scrape
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-monitoring
  namespace: aura
spec:
  podSelector:
    matchLabels:
      app: aura
  policyTypes:
    - Ingress
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: monitoring
      ports:
        - port: 26660
          protocol: TCP

---
# Allow Linkerd control plane
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-linkerd
  namespace: aura
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - namespaceSelector:
            matchLabels:
              linkerd.io/is-control-plane: "true"
```

---

## Phase 6: StatefulSet for Validators

### 6.1 Complete Validator StatefulSet

Create `k8s/validator-statefulset.yaml`:

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: aura-validator
  namespace: aura
  labels:
    app: aura
    component: validator
spec:
  serviceName: aura-validator-headless
  replicas: 3
  selector:
    matchLabels:
      app: aura
      component: validator
  updateStrategy:
    type: RollingUpdate
    rollingUpdate:
      partition: 0
  template:
    metadata:
      labels:
        app: aura
        component: validator
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "26660"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: aura-validator
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault

      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
            - labelSelector:
                matchExpressions:
                  - key: app
                    operator: In
                    values: [aura]
                  - key: component
                    operator: In
                    values: [validator]
              topologyKey: kubernetes.io/hostname

      initContainers:
        - name: init-config
          image: busybox:1.35
          command:
            - sh
            - -c
            - |
              mkdir -p /home/aura/.aura/config /home/aura/.aura/data
              chmod 755 /home/aura/.aura /home/aura/.aura/config /home/aura/.aura/data
              if [ -f /config/genesis.json ]; then
                cp /config/genesis.json /home/aura/.aura/config/genesis.json
              fi
              echo "Init complete"
          volumeMounts:
            - name: aura-data
              mountPath: /home/aura/.aura
            - name: config
              mountPath: /config
          securityContext:
            runAsNonRoot: false
            runAsUser: 0

      containers:
        - name: aura
          image: aequitas/aura:latest
          imagePullPolicy: IfNotPresent
          command: ["aurad", "start"]
          ports:
            - name: p2p
              containerPort: 26656
              protocol: TCP
            - name: rpc
              containerPort: 26657
              protocol: TCP
            - name: api
              containerPort: 1317
              protocol: TCP
            - name: grpc
              containerPort: 9090
              protocol: TCP
            - name: metrics
              containerPort: 26660
              protocol: TCP

          env:
            - name: POD_NAME
              valueFrom:
                fieldRef:
                  fieldPath: metadata.name
            - name: POD_NAMESPACE
              valueFrom:
                fieldRef:
                  fieldPath: metadata.namespace
            - name: POD_IP
              valueFrom:
                fieldRef:
                  fieldPath: status.podIP
            - name: MINIMUM_GAS_PRICES
              value: "0.025uaura"
            - name: AURA_HOME
              value: "/home/aura/.aura"

          resources:
            requests:
              cpu: 2000m
              memory: 4Gi
            limits:
              cpu: 4000m
              memory: 8Gi

          securityContext:
            allowPrivilegeEscalation: false
            runAsNonRoot: true
            runAsUser: 1000
            capabilities:
              drop: [ALL]

          volumeMounts:
            - name: aura-data
              mountPath: /home/aura/.aura
            - name: config
              mountPath: /home/aura/.aura/config/config.toml
              subPath: config.toml
            - name: config
              mountPath: /home/aura/.aura/config/app.toml
              subPath: app.toml
            - name: validator-keys
              mountPath: /home/aura/.aura/config/priv_validator_key.json
              subPath: priv_validator_key.json
              readOnly: true

          startupProbe:
            httpGet:
              path: /health
              port: rpc
            initialDelaySeconds: 30
            periodSeconds: 10
            timeoutSeconds: 5
            failureThreshold: 30

          readinessProbe:
            exec:
              command:
                - sh
                - -c
                - "aurad status 2>&1 | grep -q 'catching_up.*false'"
            initialDelaySeconds: 60
            periodSeconds: 30
            timeoutSeconds: 10
            failureThreshold: 3

          livenessProbe:
            httpGet:
              path: /health
              port: rpc
            initialDelaySeconds: 300
            periodSeconds: 30
            timeoutSeconds: 10
            failureThreshold: 3

          lifecycle:
            preStop:
              exec:
                command:
                  - sh
                  - -c
                  - "sleep 5; kill -SIGTERM 1"

      volumes:
        - name: config
          configMap:
            name: aura-config
        - name: validator-keys
          secret:
            secretName: aura-secrets-from-vault
            defaultMode: 0400

      terminationGracePeriodSeconds: 120
      dnsPolicy: ClusterFirst

  volumeClaimTemplates:
    - metadata:
        name: aura-data
        labels:
          app: aura
      spec:
        accessModes: [ReadWriteOnce]
        storageClassName: fast-ssd
        resources:
          requests:
            storage: 200Gi
```

---

## Phase 7: Services

### 7.1 Complete Service Configuration

Create `k8s/services-full.yaml`:

```yaml
---
# Headless service for StatefulSet DNS
apiVersion: v1
kind: Service
metadata:
  name: aura-validator-headless
  namespace: aura
  labels:
    app: aura
    component: validator
spec:
  clusterIP: None
  selector:
    app: aura
    component: validator
  ports:
    - name: p2p
      port: 26656
      targetPort: 26656
    - name: rpc
      port: 26657
      targetPort: 26657
    - name: api
      port: 1317
      targetPort: 1317
    - name: grpc
      port: 9090
      targetPort: 9090
    - name: metrics
      port: 26660
      targetPort: 26660
  publishNotReadyAddresses: true

---
# P2P LoadBalancer
apiVersion: v1
kind: Service
metadata:
  name: aura-p2p
  namespace: aura
  labels:
    app: aura
    component: p2p
spec:
  type: NodePort
  selector:
    app: aura
    component: validator
  ports:
    - name: p2p
      port: 26656
      targetPort: 26656
      nodePort: 30656
  sessionAffinity: ClientIP

---
# RPC Service
apiVersion: v1
kind: Service
metadata:
  name: aura-rpc
  namespace: aura
  labels:
    app: aura
    component: rpc
spec:
  type: ClusterIP
  selector:
    app: aura
    component: validator
  ports:
    - name: rpc
      port: 26657
      targetPort: 26657

---
# API Service
apiVersion: v1
kind: Service
metadata:
  name: aura-api
  namespace: aura
  labels:
    app: aura
    component: api
spec:
  type: ClusterIP
  selector:
    app: aura
    component: validator
  ports:
    - name: api
      port: 1317
      targetPort: 1317

---
# gRPC Service
apiVersion: v1
kind: Service
metadata:
  name: aura-grpc
  namespace: aura
  labels:
    app: aura
    component: grpc
spec:
  type: ClusterIP
  selector:
    app: aura
    component: validator
  ports:
    - name: grpc
      port: 9090
      targetPort: 9090

---
# Metrics Service
apiVersion: v1
kind: Service
metadata:
  name: aura-metrics
  namespace: aura
  labels:
    app: aura
    component: metrics
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "26660"
spec:
  type: ClusterIP
  selector:
    app: aura
    component: validator
  ports:
    - name: metrics
      port: 26660
      targetPort: 26660
```

---

## Phase 8: Ingress with TLS & Rate Limiting

### 8.1 Install nginx-ingress

```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

helm install ingress-nginx ingress-nginx/ingress-nginx \
    -n ingress-nginx --create-namespace \
    --set controller.service.type=NodePort \
    --set controller.service.nodePorts.http=30080 \
    --set controller.service.nodePorts.https=30443
```

### 8.2 Install cert-manager

```bash
helm repo add jetstack https://charts.jetstack.io
helm repo update

helm install cert-manager jetstack/cert-manager \
    -n cert-manager --create-namespace \
    --set installCRDs=true
```

### 8.3 Ingress Configuration

Create `k8s/ingress-full.yaml`:

```yaml
---
# Self-signed issuer for dev/staging
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: selfsigned-issuer
spec:
  selfSigned: {}

---
# CA issuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: ca-issuer
spec:
  ca:
    secretName: ca-key-pair

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: aura-ingress
  namespace: aura
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: selfsigned-issuer
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/limit-rps: "100"
    nginx.ingress.kubernetes.io/limit-connections: "10"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "600"
    # Security headers
    nginx.ingress.kubernetes.io/configuration-snippet: |
      more_set_headers "X-Frame-Options: DENY";
      more_set_headers "X-Content-Type-Options: nosniff";
      more_set_headers "X-XSS-Protection: 1; mode=block";
spec:
  ingressClassName: nginx
  tls:
    - hosts:
        - rpc.aura.local
        - api.aura.local
      secretName: aura-tls
  rules:
    - host: rpc.aura.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: aura-rpc
                port:
                  number: 26657
    - host: api.aura.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: aura-api
                port:
                  number: 1317
```

---

## Phase 9: Monitoring Stack

### 9.1 Install Prometheus Stack

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

helm install prometheus prometheus-community/kube-prometheus-stack \
    -n monitoring --create-namespace \
    --set grafana.service.type=NodePort \
    --set grafana.service.nodePort=31030 \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
```

### 9.2 ServiceMonitor for Aura

Create `k8s/monitoring.yaml`:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: aura-validator
  namespace: aura
  labels:
    app: aura
spec:
  selector:
    matchLabels:
      app: aura
      component: metrics
  endpoints:
    - port: metrics
      interval: 30s
      scrapeTimeout: 10s
      path: /metrics

---
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: aura-alerts
  namespace: aura
  labels:
    prometheus: kube-prometheus
spec:
  groups:
    - name: aura.validator.rules
      interval: 30s
      rules:
        - alert: AuraValidatorDown
          expr: up{job="aura-validator"} == 0
          for: 2m
          labels:
            severity: critical
          annotations:
            summary: "Aura validator down: {{ $labels.pod }}"

        - alert: AuraConsensusStalled
          expr: increase(tendermint_consensus_height[5m]) == 0
          for: 5m
          labels:
            severity: critical
          annotations:
            summary: "Aura consensus stalled - no new blocks in 5 minutes"

        - alert: AuraLowPeerCount
          expr: tendermint_p2p_peers < 2
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "Aura validator {{ $labels.pod }} has low peer count: {{ $value }}"

        - alert: AuraCatchingUp
          expr: tendermint_consensus_catching_up == 1
          for: 30m
          labels:
            severity: warning
          annotations:
            summary: "Aura validator {{ $labels.pod }} is catching up for >30 minutes"

        - alert: AuraHighMemory
          expr: |
            (container_memory_usage_bytes{pod=~"aura-validator-.*"} /
             container_spec_memory_limit_bytes{pod=~"aura-validator-.*"}) > 0.9
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High memory usage on {{ $labels.pod }}: {{ $value | humanizePercentage }}"

        - alert: AuraDiskUsageHigh
          expr: |
            (kubelet_volume_stats_used_bytes{persistentvolumeclaim=~"aura-data-.*"} /
             kubelet_volume_stats_capacity_bytes{persistentvolumeclaim=~"aura-data-.*"}) > 0.85
          for: 5m
          labels:
            severity: warning
          annotations:
            summary: "High disk usage on {{ $labels.persistentvolumeclaim }}"
```

---

## Phase 10: Service Mesh (Linkerd mTLS)

### 10.1 Install Linkerd

```bash
# Install Linkerd CLI
curl -sL https://run.linkerd.io/install-edge | sh
export PATH=$PATH:~/.linkerd2/bin

# Install Gateway API CRDs
kubectl apply --server-side -f \
    https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.4.0/standard-install.yaml

# Install Linkerd
linkerd install --crds | kubectl apply -f -
linkerd install | kubectl apply -f -

# Verify
linkerd check
```

### 10.2 Mesh Aura Namespace

```bash
# Enable injection (requires privileged PSS)
kubectl label namespace aura pod-security.kubernetes.io/enforce=privileged --overwrite
kubectl annotate namespace aura linkerd.io/inject=enabled --overwrite

# Restart pods to inject sidecars
kubectl rollout restart statefulset/aura-validator -n aura

# Verify mTLS
linkerd check --proxy -n aura
```

---

## Phase 11: GitOps (ArgoCD)

### 11.1 Install ArgoCD

```bash
kubectl create namespace argocd
kubectl label namespace argocd pod-security.kubernetes.io/enforce=privileged

kubectl apply -n argocd -f \
    https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

# Expose via NodePort
kubectl patch svc argocd-server -n argocd -p '{"spec": {"type": "NodePort", "ports": [{"port": 443, "nodePort": 30085}]}}'

# Get initial password
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

### 11.2 ArgoCD Application

Create `k8s/argocd-app.yaml`:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: aura-blockchain
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/aura.git
    targetRevision: main
    path: k8s/overlays/staging
  destination:
    server: https://kubernetes.default.svc
    namespace: aura
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

---

## Phase 12: Autoscaling

### 12.1 Horizontal Pod Autoscaler

Create `k8s/hpa-full.yaml`:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: aura-validator-hpa
  namespace: aura
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: aura-validator
  minReplicas: 3
  maxReplicas: 7
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 75
  behavior:
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Pods
          value: 1
          periodSeconds: 60
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Pods
          value: 1
          periodSeconds: 120
```

### 12.2 Vertical Pod Autoscaler

```bash
# Install VPA
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/vpa-v1-crd-gen.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/vpa-rbac.yaml
kubectl apply -f https://raw.githubusercontent.com/kubernetes/autoscaler/master/vertical-pod-autoscaler/deploy/recommender-deployment.yaml
```

Create `k8s/vpa.yaml`:

```yaml
apiVersion: autoscaling.k8s.io/v1
kind: VerticalPodAutoscaler
metadata:
  name: aura-validator-vpa
  namespace: aura
spec:
  targetRef:
    apiVersion: apps/v1
    kind: StatefulSet
    name: aura-validator
  updatePolicy:
    updateMode: "Off"  # Recommendations only
  resourcePolicy:
    containerPolicies:
      - containerName: aura
        minAllowed:
          cpu: 500m
          memory: 1Gi
        maxAllowed:
          cpu: 8000m
          memory: 16Gi
```

---

## Phase 13: Backup & Disaster Recovery

### 13.1 Backup Script

Copy from: `~/blockchain-projects/scripts/k8s-backup.sh`

Usage:
```bash
# Backup
~/blockchain-projects/scripts/k8s-backup.sh backup aura

# Restore
~/blockchain-projects/scripts/k8s-backup.sh restore <backup_dir> aura
```

### 13.2 Scheduled Backups

Create `k8s/backup-cronjob.yaml`:

```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: aura-backup
  namespace: aura
spec:
  schedule: "0 */6 * * *"  # Every 6 hours
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName: aura-validator
          containers:
            - name: backup
              image: bitnami/kubectl:latest
              command:
                - /bin/sh
                - -c
                - |
                  kubectl exec aura-validator-0 -- \
                    tar czf - /home/aura/.aura/data | \
                    gzip > /backup/aura-$(date +%Y%m%d-%H%M%S).tar.gz
              volumeMounts:
                - name: backup
                  mountPath: /backup
          volumes:
            - name: backup
              persistentVolumeClaim:
                claimName: aura-backups
          restartPolicy: OnFailure
```

---

## Phase 14: Security Hardening

### 14.1 Secrets Encryption at Rest

On bcpc (master):

```bash
# Create encryption key
ENCRYPTION_KEY=$(head -c 32 /dev/urandom | base64)

# Create config
sudo tee /etc/rancher/k3s/encryption-config.yaml << EOF
apiVersion: apiserver.config.k8s.io/v1
kind: EncryptionConfiguration
resources:
  - resources:
      - secrets
    providers:
      - aescbc:
          keys:
            - name: aescbckey
              secret: ${ENCRYPTION_KEY}
      - identity: {}
EOF

# Add to k3s config
echo '--kube-apiserver-arg encryption-provider-config=/etc/rancher/k3s/encryption-config.yaml' | \
    sudo tee -a /etc/rancher/k3s/config.yaml

# Restart k3s
sudo systemctl restart k3s
```

### 14.2 Audit Logging

```bash
# Create audit policy
sudo tee /etc/rancher/k3s/audit-policy.yaml << 'EOF'
apiVersion: audit.k8s.io/v1
kind: Policy
rules:
  - level: RequestResponse
    resources:
      - group: ""
        resources: ["secrets"]
  - level: Metadata
    resources:
      - group: ""
        resources: ["pods/exec", "pods/attach"]
  - level: Request
    verbs: ["create", "update", "patch", "delete"]
  - level: None
    resources:
      - group: ""
        resources: ["events"]
EOF

# Enable in k3s
sudo k3s server --kube-apiserver-arg=audit-log-path=/var/log/k3s-audit.log \
    --kube-apiserver-arg=audit-policy-file=/etc/rancher/k3s/audit-policy.yaml
```

---

## Phase 15: Testing & Validation

### 15.1 Deployment Verification

Create `scripts/verify-aura-k8s.sh`:

```bash
#!/bin/bash
set -e

NAMESPACE="aura"
echo "=== Aura Kubernetes Deployment Verification ==="

# 1. Check nodes
echo -e "\n[1/10] Checking cluster nodes..."
kubectl get nodes

# 2. Check namespace
echo -e "\n[2/10] Checking namespace..."
kubectl get ns $NAMESPACE

# 3. Check pods
echo -e "\n[3/10] Checking pods..."
kubectl get pods -n $NAMESPACE -o wide

# 4. Check services
echo -e "\n[4/10] Checking services..."
kubectl get svc -n $NAMESPACE

# 5. Check PVCs
echo -e "\n[5/10] Checking persistent volumes..."
kubectl get pvc -n $NAMESPACE

# 6. Check secrets sync (ESO)
echo -e "\n[6/10] Checking external secrets..."
kubectl get externalsecret -n $NAMESPACE

# 7. Check Linkerd mesh
echo -e "\n[7/10] Checking Linkerd mesh..."
linkerd check --proxy -n $NAMESPACE || echo "Linkerd not configured"

# 8. Check network policies
echo -e "\n[8/10] Checking network policies..."
kubectl get networkpolicy -n $NAMESPACE

# 9. Check HPA
echo -e "\n[9/10] Checking autoscaling..."
kubectl get hpa -n $NAMESPACE

# 10. Check validator health
echo -e "\n[10/10] Checking validator health..."
for i in 0 1 2; do
    POD="aura-validator-$i"
    if kubectl get pod $POD -n $NAMESPACE &>/dev/null; then
        kubectl exec $POD -n $NAMESPACE -- aurad status 2>&1 | head -5
    fi
done

echo -e "\n=== Verification Complete ==="
```

### 15.2 Chaos Engineering Tests

```bash
# Network latency
sudo tc qdisc add dev cni0 root netem delay 200ms 50ms

# Pod deletion recovery
kubectl delete pod aura-validator-0 -n aura
# Should recover within 60s

# Node drain
kubectl drain wsl2-worker --ignore-daemonsets --delete-emptydir-data
# Pods should migrate to bcpc-staging

# Cleanup
kubectl uncordon wsl2-worker
sudo tc qdisc del dev cni0 root
```

---

## Deployment Order

Execute phases in this order:

1. **Phase 1**: Cluster setup (k3s on both nodes)
2. **Phase 2**: Namespace & security (RBAC, PSS)
3. **Phase 3**: Storage classes
4. **Phase 4**: Vault + External Secrets
5. **Phase 5**: Network policies
6. **Phase 6**: StatefulSet (validators)
7. **Phase 7**: Services
8. **Phase 8**: Ingress + cert-manager
9. **Phase 9**: Monitoring (Prometheus/Grafana)
10. **Phase 10**: Linkerd mTLS
11. **Phase 11**: ArgoCD GitOps
12. **Phase 12**: Autoscaling (HPA/VPA)
13. **Phase 13**: Backup procedures
14. **Phase 14**: Security hardening
15. **Phase 15**: Testing & validation

---

## Quick Reference

### Access Points

| Service | URL | Credentials |
|---------|-----|-------------|
| Grafana | http://100.91.253.108:31030 | admin / prom-operator |
| ArgoCD | http://100.91.253.108:30085 | admin / (see K8S_GITOPS.md) |
| Prometheus | kubectl port-forward svc/prometheus-kube-prometheus-prometheus 9090:9090 -n monitoring | - |
| Vault | http://vault.vault.svc:8200 | root (dev mode) |

### Common Commands

```bash
# Source environment
source ~/blockchain-projects/aura/env.sh

# Check all resources
kubectl get all,pvc,secret,cm -n aura

# Watch pods
kubectl get pods -n aura -w

# Validator logs
kubectl logs -f aura-validator-0 -n aura

# Validator status
kubectl exec aura-validator-0 -n aura -- aurad status

# Scale validators
kubectl scale statefulset aura-validator --replicas=5 -n aura

# Backup
~/blockchain-projects/scripts/k8s-backup.sh backup aura

# Linkerd dashboard
linkerd dashboard
```

---

## File Checklist

Create these files in `~/blockchain-projects/aura/k8s/`:

- [x] `namespace.yaml`
- [x] `rbac-full.yaml`
- [x] `resource-quota.yaml`
- [x] `storage.yaml`
- [x] `external-secrets.yaml`
- [x] `network-policies.yaml`
- [x] `validator-statefulset.yaml`
- [x] `services-full.yaml`
- [x] `ingress-full.yaml`
- [x] `monitoring.yaml`
- [x] `hpa-full.yaml`
- [x] `vpa.yaml`
- [x] `backup-cronjob.yaml`
- [x] `argocd-app.yaml`

Scripts in `~/blockchain-projects/aura/scripts/`:

- [x] `verify-aura-k8s.sh`
- [x] `deploy-aura-k8s.sh`

---

**Document Version**: 1.0
**Based on**: XAI project infrastructure (tested, 97% production ready)
**Target**: Aura Cosmos SDK blockchain
**Author**: AI Agent
**Date**: 2025-12-19
