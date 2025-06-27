<div align="center">
  <img src="https://github.com/user-attachments/assets/a07c897e-c66f-47de-ac04-1649e4a3ea48" alt="Upcube logo" width="700px" />  
  <h3>Bare minimum kubernetes deployment management platform, using Pod Service Account, built for usage behind Cloudflare Zero Trust</h3>
</div>

### Production Deployment

`upkube` is a `~60 MiB` container built using `golang` with `templ` html templating, no js. When deployed using `UPKUBE_ENV=PROD` variables (recommended for production usage), it connects to the Kubernetes cluster using **Pod Service Account**. It does not have auth, it's built for usage behind **Cloudflare Zero Trust**. 

```yaml
...
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

For full Kubernetes deployment example: [k8s](https://github.com/KunalSin9h/upkube/tree/master/k8s) directory. 

#### Environment Variables

- `UPKUBE_HOST` - Set host for http service, default is `127.0.0.1`
- `UPKUBE_PORT` - Set port for http service, default is `8080`
- `UPKUBE_ENV` - Set application environment, `PROD` or `DEV`, default is `DEV`.

  - When using `PROD` environment (recommended for **production usage**), `upkube` connects with in-cluster configuration to the Kubernetes cluster, which uses the service account Kubernetes provides to pods. 

  - When using `DEV`, it connects from a master url or a kubeconfig filepath. default is `~/.kube/config`


#### Roadmap

- [ ] Support Activity Logs
- [ ] Request and Approve workflow

### Local Development

### Stack

- go
- [templ](https://templ.guide/)
- tailwindcss
- minikube

After cloning the repo...

Download dependencies and tools. 

```bash
go mod download
```

Download and make sure `minikube` is running, for local k8s testing. 

```bash
minikube start
```

Start live reloaded Application 

```bash
go tool air
```

This will start the application on specieif port (using env), or deafult is `http://localhost:8080`


