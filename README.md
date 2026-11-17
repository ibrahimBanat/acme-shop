# Acme Shop

The order and refund API behind acme.shop.

```
go test ./...
```

## Web

React + TypeScript frontend in `web/` (Vite, Vitest).

```
cd web
npm install
npm run dev        # local dev server
npm test           # unit tests
npm run typecheck
npm run build
```

The login page (`/login`) posts to the existing auth service at
`$VITE_AUTH_API_URL/login` (set it in `.env` at the repo root, see `.env.example`)
with `{ identifier, password, rememberMe }`, where `identifier` is an email or
username. The service is expected to set an HttpOnly session cookie; on success
the user is redirected to `/dashboard`.
