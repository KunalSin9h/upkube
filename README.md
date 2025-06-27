<div align="center">
  <img src="https://github.com/user-attachments/assets/a07c897e-c66f-47de-ac04-1649e4a3ea48" alt="Upcube logo" width="700px" />  
  <h3>Bare minimum kubernetes deployment management platform, using Pod Service Account, build for usage behind Cloudflare Zeroauth</h3>
</div>

## Production Deployment

`Upkube` is a `~60 MiB` container build using `golang` with `templ` html templating, no js. When deployed using `UPKUBE_ENV=PROD` variables (recommended for production usage), it connect to kubernetes cluster using **Pod Service Account**. It is does not have auth, its build for usage behind **Cloudflare Zerotrust**. 

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: upkube
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: upkube
  template:
    metadata:
      labels:
        app: upkube
    spec:
      serviceAccountName: upkube-sa
      containers:
      - name: upkube
        image: ghcr.io/kunalsin9h/upkube:latest
        ports:
        - containerPort: 8080
        env:
        - name: UPKUBE_ENV
          value: "PROD"
        - name: UPKUBE_HOST
          value: "0.0.0.0"
        - name: UPKUBE_PORT
          value: "8080"
```

Fore full kubernetes deployment exampel: 



### Environment Variables

- `UPKUBE_HOST` - Set host for http service, default is `127.0.0.1`
- `UPKUBE_PORT` - Set port for http serivde, default is `8080`
- `UPKUBE_ENV` - Set application environment, `PROD` or `DEV`, default is `DEV`.

  - When using `PROD` environment (recommended for **production usage**), `upkube` connect with **In Cluster** configuration to kubernetes cluster, which uses the **service account** kubernetes gives to pods. 

  - When using `DEV`, it connects from a master url or a kubeconfig filepath. default is `~/.kube/config`



## Local Development

## Stack

- GO
- [TEMPL](https://templ.guide/)
- TAILWINDCSS
- MINIKUBE

Download dependencies and tools. 

```bash
go mod download
```

Download and make sure `minikube` is running. 

```bash
minikube start
```

Start live reloaded Application 

```bash
go tool air
```

After update the template, generate go code using: 

```bash
go tool templ generate
```
