# Job Tracker

Track job applications — Go REST API + Vue 3 SPA.

<img width="1197" height="628" alt="image" src="https://github.com/user-attachments/assets/e216a7f7-6ab2-4c50-b913-0123d9d73194" />


## Authentication

The entire app — all `/api/*` routes and the static SPA — is gated behind HTTP
Basic Auth. Credentials are read from two required environment variables at
startup; the server refuses to start (`log.Fatal`) if either is unset:

| Variable        | Description                                      |
| --------------- | ------------------------------------------------ |
| `AUTH_EMAIL`    | Basic Auth username (the login email).           |
| `AUTH_PASSWORD` | Basic Auth password.                             |

The browser's native login dialog handles the prompt — there is no custom login
UI. The only unauthenticated route is `GET /healthz`, which returns `200 ok` for
container healthchecks.

```bash
AUTH_EMAIL=you@example.com AUTH_PASSWORD=secret go run .
```

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

## Desktop build

A Wails v2 desktop target (`cmd/desktop`) packages the same backend and
frontend as a native app: no basic auth, no network listener, SQLite lives in
the OS user data dir (override with `DB_PATH`).

```bash
cd web && npm run build   # embedded via web/embed.go, must run first
go build -tags desktop,production,webkit2_41 -o jobtracker-desktop ./cmd/desktop
./jobtracker-desktop
```

Requires the `webkit2gtk-4.1` runtime on Linux (the `webkit2_41` build tag
matches it; omit the tag only if `webkit2gtk-4.0` is installed instead).
Windows and macOS builds drop `webkit2_41` (see
`.github/workflows/release.yml`, which builds and publishes all three OSes on
tag push).

## Deploy to Railway

Railway builds from the repo-root `Dockerfile` (config in `railway.toml`), which
builds the Vue frontend and Go binary and serves both on `:8080`.

- **Volume**: mount a persistent volume (e.g. at `/data`) for the SQLite file.
- **`DB_PATH`**: set to a path inside that volume, e.g. `/data/jobs.db`, so data
  survives redeploys.
- **`AUTH_EMAIL` / `AUTH_PASSWORD`**: set the basic-auth credentials (see
  `.env.example`).
- **Healthcheck**: `/healthz` (no auth required).
- **Replicas**: keep at **1** — SQLite is single-writer and must never scale out.
