> **TL;DR** what it is? it is an instant go backend server ready to be use with different backend server project with p infra (logger, jwt, authen, ratelimit, setup boilerplate) provided so you "don't repeat yourself" and save time and effort

# Go Reusable — Instant Go Backend Boilerplate

Go Reusable is a minimal, ergonomic Go backend starter that provides common infrastructure so you can avoid repeating the same setup across projects. It bundles logging, JWT-based auth middleware, rate limiting, Docker tooling, and Kubernetes kustomize overlays so you can build, run, and deploy quickly.

### Why this repo ###
- Reusable: pick the components you need and plug them into your app.
- Ergonomic: batteries-included defaults while keeping customization simple.
- Practical: includes `Dockerfile`, `docker-compose.yaml`, and kustomize overlays for straightforward local development and production deployments.

### Focus ###
This repo focuses on general-purpose middleware and infra that apply across most web backends, with an emphasis on three goals:
- Builds: a cross-arch build via the provided `Dockerfile` and a Makefile helper.
- Works: an easy local dev flow using `docker-compose` and `.env` support.
- Deploys: example kustomize overlays under `deployment/` so you can adapt manifests for environments.

### Quick links ###
- Dockerfile: [Dockerfile](Dockerfile#L1)
- Compose: [docker-compose.yaml](docker-compose.yaml#L1)
- Main entry: [main.go](main.go#L1)

### Want to see this thing in action? ###

- Visit [PopclickV2 Backend Server](https://github.com/EarthYim/popclickV2/tree/main/backend) to see how I intergrate this frontend, and deploy in production environment.

### Getting started ###

1. Copy `.env.template` to `.env` and fill the keys (JWT keys, redis, etc.).
2. For local development using Docker Compose:

```bash
cp .env.template .env
docker-compose build
docker-compose up
```

3. The server listens on port `8080` by default. Healthcheck endpoints are available at `/` and `/health`.

## Components (detailed) ##

- **Logger**
	- Implemented in [logger/logger.go](logger/logger.go#L1).
	- Uses `log/slog` with a JSON handler and exposes a helper `logger.Logger(ctx)` to get the request-scoped logger.
	- `middleware.LoggerMiddleware` attaches a request logger into the request context so other handlers and middleware can log with request fields.

- **JWT middleware**
	- Implementation in [middleware/jwt.go](middleware/jwt.go#L1) and auth helpers in `auth/`.
	- `JwtMiddleware` enforces a Bearer token in `Authorization` header and validates tokens using the repo's ECDSA public key (base64-encoded via env).
	- `JwtOptionalMiddleware` allows unauthenticated requests while still extracting claims when present.
	- `auth.AdminLoginHandler` in [auth/admin_authen.go](auth/admin_authen.go#L1) shows how tokens are created and signed using the base64 private key.

- **Rate limiter middleware**
	- Implemented in [middleware/rate_limiter.go](middleware/rate_limiter.go#L1).
	- Uses Redis sorted sets to track requests within a rolling window per key.
	- Middleware entries `AppRateLimiterMiddleware` and `AuthRateLimiterMiddleware` integrate with application context to enforce per-client or per-auth limits defined in config/env.

**Configuration**
- Configuration is driven by environment variables parsed via `config.C()` in [config/config.go](config/config.go#L1). See `.env.template` for available settings (log level, JWT keys, rate limiter values, redis).

**Redis**
- `redis.NewConnection` (in `redis/redis.go`) builds a `redis.UniversalClient` with optional TLS support driven from env.

**Build & Deploy**
- Makefile already provided everything you needed
- For local development

```bash
make dev
# run docker-compose build 
# and
# docker-compose up
```

- Build for production using the Makefile helper:

```bash
make build-prod
```

- Kustomize overlays are in `deployment/server/k8s/overlays/` to help with environment-specific customization and deployment.

**Project layout**

- `main.go` — application entrypoint.
- `middleware/` — JWT, logger, rate limiter and other middleware.
- `auth/` — example auth handlers and key parsing helpers.
- `logger/` — global logger helpers.
- `redis/` — redis connection helper.
- `deployment/` — k8s manifests and overlays.

**Next steps**
- Fill `.env` with real keys and redis address.
- Integrate business routes and handlers into `main.go` or a separate router package.
- Tune rate limiter and CORS as required for your app.
