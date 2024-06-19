## Adresse de l'image du webhook
mohamedsubmit.azurecr.io/webhook-appconfig:latest

## Command corection

```bash
# Déploiement du webhook
kubectl apply -f k8s.yaml

# Déploiement du crd
kubectl apply -f bases-crd.yaml

# Déploiement des ressources
kbuectl apply -f v1.yaml
kbuectl apply -f v2.yaml
```
