#!/bin/bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: scripts/deploy-aura-k8s.sh [--context <name>] [--skip-helm]

Deploys the Aura Kubernetes manifests in the recommended order.

Options:
  --context <name>   Use a specific kube-context (applies to kubectl and helm)
  --skip-helm        Skip Helm-based dependencies (ingress, cert-manager,
                     Vault, External Secrets Operator, kube-prometheus-stack)
EOF
}

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
K8S_DIR="${ROOT_DIR}/k8s"
KUBECTL_CONTEXT=""
SKIP_HELM=0

while [[ $# -gt 0 ]]; do
  case "$1" in
    --context)
      KUBECTL_CONTEXT="${2:-}"
      shift
      ;;
    --skip-helm)
      SKIP_HELM=1
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "Unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
  shift
done

log() {
  echo "[$(date -Iseconds)] $*"
}

require_cmd() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

require_cmd kubectl
(( SKIP_HELM == 0 )) && require_cmd helm

KUBECTL=(kubectl)
HELM=(helm)
if [[ -n "$KUBECTL_CONTEXT" ]]; then
  KUBECTL+=(--context "$KUBECTL_CONTEXT")
  HELM+=(--kube-context "$KUBECTL_CONTEXT")
fi

apply() {
  local target="$1"
  shift
  "${KUBECTL[@]}" apply "$@" -f "$target"
}

log "Applying namespace"
apply "${K8S_DIR}/namespace.yaml"

if (( SKIP_HELM == 0 )); then
  log "Adding/refreshing Helm repositories"
  "${HELM[@]}" repo add hashicorp https://helm.releases.hashicorp.com >/dev/null
  "${HELM[@]}" repo add external-secrets https://charts.external-secrets.io >/dev/null
  "${HELM[@]}" repo add ingress-nginx https://kubernetes.github.io/ingress-nginx >/dev/null
  "${HELM[@]}" repo add jetstack https://charts.jetstack.io >/dev/null
  "${HELM[@]}" repo add prometheus-community https://prometheus-community.github.io/helm-charts >/dev/null
  "${HELM[@]}" repo update >/dev/null

  log "Installing Vault (dev mode)"
  "${HELM[@]}" upgrade --install vault hashicorp/vault -n vault --create-namespace \
    --set server.dev.enabled=true \
    --set server.dev.devRootToken=root \
    --set injector.enabled=false

  log "Installing External Secrets Operator"
  "${HELM[@]}" upgrade --install external-secrets external-secrets/external-secrets \
    -n external-secrets --create-namespace \
    --set installCRDs=true

  log "Installing ingress-nginx"
  "${HELM[@]}" upgrade --install ingress-nginx ingress-nginx/ingress-nginx \
    -n ingress-nginx --create-namespace \
    --set controller.service.type=NodePort \
    --set controller.service.nodePorts.http=30080 \
    --set controller.service.nodePorts.https=30443

  log "Installing cert-manager"
  "${HELM[@]}" upgrade --install cert-manager jetstack/cert-manager \
    -n cert-manager --create-namespace \
    --set installCRDs=true

  log "Installing kube-prometheus-stack"
  "${HELM[@]}" upgrade --install prometheus prometheus-community/kube-prometheus-stack \
    -n monitoring --create-namespace \
    --set grafana.service.type=NodePort \
    --set grafana.service.nodePort=10030 \
    --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false
fi

log "Applying RBAC"
apply "${K8S_DIR}/rbac-full.yaml"

log "Applying resource quotas and limits"
apply "${K8S_DIR}/resource-quota.yaml"

log "Applying storage classes"
apply "${K8S_DIR}/storage.yaml"

log "Applying validator config map"
apply "${K8S_DIR}/base/configmap.yaml" -n aura

log "Applying genesis config map"
apply "${K8S_DIR}/genesis-configmap.yaml"

log "Applying external secrets wiring"
apply "${K8S_DIR}/external-secrets.yaml"

log "Applying network policies"
apply "${K8S_DIR}/network-policies.yaml"

log "Applying services"
apply "${K8S_DIR}/services-full.yaml"

log "Applying validator StatefulSet"
apply "${K8S_DIR}/validator-statefulset.yaml"

log "Applying autoscaling (HPA/VPA)"
apply "${K8S_DIR}/hpa-full.yaml"
apply "${K8S_DIR}/vpa.yaml"

log "Applying ingress and TLS issuers"
apply "${K8S_DIR}/ingress-full.yaml"

log "Applying monitoring resources"
apply "${K8S_DIR}/monitoring.yaml"

log "Applying backup cronjob"
apply "${K8S_DIR}/backup-cronjob.yaml"

log "Applying ArgoCD application definition"
apply "${K8S_DIR}/argocd-app.yaml"

log "Deployment manifest application complete"
