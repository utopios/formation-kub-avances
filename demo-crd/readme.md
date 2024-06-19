### Initialisation d'un projet operateur avec kubebuilder 
```bash
kubebuilder init --domain mydomain.com --repo github.com/myname/myoperator
```
### Création de l'api crds
```bash
create api --group domain.com --version v1 --kind demo
```

### Run du manifest
```bash
make manifest
```