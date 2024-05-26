Можливо ось такий підхід спробувати

## Налаштування TLS для ArgoCD:

```code
cat <<EOF > argocd-tls-certs-cm.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-tls-certs-cm
  namespace: argocd
data:
  tls.crt: |
    <your-certificate>
  tls.key: |
    <your-private-key>
EOF
```

```code
kubectl apply -f argocd-tls-certs-cm.yaml
```

## Використати токен задля автентифікації:
Створити токен API у ArgoCD для автентифікації:

```code
argocd account generate-token --account <account-name>
```
## Зміна частини конфігураційного файла:
 Підключення до ArgoCD
```code
ARGOCD_OPTS="--grpc-web --server localhost:8080 --auth-token ${ARGOCD_TOKEN}"
argocd login ${ARGOCD_OPTS} || { echo "Failed to login to ArgoCD"; exit 1; }
```
