# Kubernetes Deployment for seqctl

This directory contains Kubernetes manifests for deploying seqctl using Kustomize.

## Quick Start

### Basic deployment (default namespace: seqctl)

```bash
kubectl apply -k k8s/
```

### Development environment

```bash
kubectl apply -k k8s/overlays/development/
```

### Production environment

```bash
kubectl apply -k k8s/overlays/production/
```

## Directory Structure

```
k8s/
├── base resources
   ├── namespace.yaml     # Namespace definition
   ├── rbac.yaml          # ServiceAccount, ClusterRole, ClusterRoleBinding
   ├── configmap.yaml     # Base configuration
   ├── deployment.yaml    # Deployment specification
   ├── service.yaml       # Service definition
   ├── config.toml        # Configuration file
   └── kustomization.yaml # Kustomize base
```

## Customization

### Change the sequencer selector

Edit `config.toml`:

```toml
[k8s]
selector = "app=your-sequencer-label"
```

Then rebuild:

```bash
kubectl apply -k k8s/
```

### Change the image version

```bash
cd k8s/
kustomize edit set image golemnetwork/seqctl:v1.2.3
kubectl apply -k .
```

### Add Ingress

Create `k8s/ingress.yaml`:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: seqctl
  namespace: seqctl
spec:
  rules:
    - host: seqctl.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: seqctl
                port:
                  number: 80
```

Add to `kustomization.yaml`:

```yaml
resources:
  - ingress.yaml
```

## Access seqctl

### Port forwarding

```bash
kubectl port-forward -n seqctl svc/seqctl 8080:80
```

Then open http://localhost:8080

### Check status

```bash
# View all resources
kubectl get all -n seqctl

# Check logs
kubectl logs -n seqctl -l app=seqctl

# Describe pods
kubectl describe pod -n seqctl -l app=seqctl
```

## Cleanup

```bash
# Remove default deployment
kubectl delete -k k8s/

# Remove development
kubectl delete -k k8s/overlays/development/

# Remove production
kubectl delete -k k8s/overlays/production/
```

## Troubleshooting

### No sequencers found

Check your label selector matches your sequencer pods:

```bash
# List pods that would match
kubectl get pods -A -l "app=op-conductor"

# Update config.toml with correct selector
```

### Permission denied

Verify RBAC is working:

```bash
kubectl auth can-i list pods --as=system:serviceaccount:seqctl:seqctl
```

### Connection issues

Check pod logs:

```bash
kubectl logs -n seqctl deployment/seqctl
```
