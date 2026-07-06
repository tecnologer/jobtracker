# Job Tracker

Track job applications — Go REST API + Vue 3 SPA.

<img width="1197" height="628" alt="image" src="https://github.com/user-attachments/assets/e216a7f7-6ab2-4c50-b913-0123d9d73194" />


## Run manually

**Requirements:** Go 1.22+, Node 20+

```bash
# Terminal 1 — backend
go run .

# Terminal 2 — frontend (dev, with hot reload)
cd web && npm install && npm run dev
```

Open http://localhost:5173. Vite proxies `/api` to `:8080`.

The SQLite database is created at `jobs.db` on first run. Override with:

```bash
DB_PATH=/path/to/jobs.db go run .
```

### Production build

```bash
cd web && npm run build   # outputs web/dist/
go build -o jobtracker .
./jobtracker              # serves everything on :8080
```

---

## Docker Compose (dev, hot reload)

```bash
docker compose up
```

Go backend uses `air` for hot reload; Vite frontend at http://localhost:5173. DB is `jobs_test.db` (set via `DB_PATH` in `docker-compose.yml`).

---

## Docker

```bash
docker build -t jobtracker:latest .
docker run -p 8080:8080 -v jobtracker-data:/data -e DB_PATH=/data/jobs.db jobtracker:latest
```

Open http://localhost:8080.

---

## Kubernetes

The manifest at `k8s/jobtracker.yaml` creates a PVC (1 Gi), a Deployment, and a ClusterIP Service.

> SQLite is single-writer — replicas is fixed at 1.

### Deploy

```bash
# Build and load the image into your cluster (minikube example)
minikube start
docker build -t jobtracker:latest .
minikube image load jobtracker:latest   # skip if using a registry

# Apply
kubectl apply -f k8s/jobtracker.yaml

# Verify
kubectl get pods -l app=jobtracker
kubectl logs -l app=jobtracker
```

### Access

The Service is `ClusterIP`. Expose it locally with:

```bash
kubectl port-forward svc/jobtracker 8080:80
```

Then open http://localhost:8080.

For a real ingress, add an Ingress resource or change the Service type to `LoadBalancer`.

### Teardown

```bash
kubectl delete -f k8s/jobtracker.yaml
```

The PVC is deleted with it. To keep your data, remove the PVC from the delete command:

```bash
kubectl delete deployment,service jobtracker
```
