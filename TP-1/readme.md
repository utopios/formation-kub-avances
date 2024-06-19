## Adresse de l'image du webhook
mohamedsubmit.azurecr.io/webhook-appconfig:latest

## Command corection

```bash
# Déploiement du webhook
kubectl apply -f k8s.yaml

# Déploiement du crd
kubectl apply -f bases-crd.yaml

# Déploiement des ressources
kubectl apply -f v1.yaml
kubectl apply -f v2.yaml

kubectl get acfg.v1.myorg.com -o yaml 

```
