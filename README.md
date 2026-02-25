# PolicyFlow

> **Work in progress** — MVP is functional. Production hardening, SSO integration, and advanced reporting are on the roadmap.

PolicyFlow is a **self-hosted, open-source policy management system** for companies that need a simple, auditable way to distribute policies and track employee acknowledgements — without handing your compliance data to a SaaS vendor.

---

## Features

| Feature                                                      | Status     |
| ------------------------------------------------------------ | ---------- |
| Magic-link authentication (no passwords)                     | ✅         |
| Policy authoring with versioning                             | ✅         |
| Policy state machine (Draft → Review → Published → Archived) | ✅         |
| Employee acknowledgement with cryptographic signature hash   | ✅         |
| Admin dashboard (users, policies, acknowledgement rates)     | ✅         |
| Single self-contained binary (Go + embedded Next.js)         | ✅         |
| Email notifications via SMTP                                 | ✅         |
| SQLite database (zero-ops, single file backup)               | ✅         |
| Docker / docker-compose deployment                           | ✅         |
| SSO / LDAP integration                                       | 🔜 Roadmap |
| Employee directory import (CSV / SCIM)                       | 🔜 Roadmap |
| PDF export of acknowledgement audit log                      | 🔜 Roadmap |
| Department-scoped policies                                   | 🔜 Roadmap |

---

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                   Single Binary                       │
│                  ./build/policyflow                   │
│                                                       │
│  ┌─────────────────┐    ┌─────────────────────────┐  │
│  │  Go HTTP Server │    │  Embedded Next.js (SPA) │  │
│  │  Echo v4        │    │  Static files via       │  │
│  │  /api/*         │    │  go:embed               │  │
│  └────────┬────────┘    └─────────────────────────┘  │
│           │                                           │
│  ┌────────▼────────┐                                  │
│  │  SQLite (WAL)   │                                  │
│  │  policyflow.db  │                                  │
│  └─────────────────┘                                  │
└──────────────────────────────────────────────────────┘
```

## Tech Stack

- **Backend**: Go 1.25, Echo v4, SQLite (`modernc.org/sqlite` — pure Go, no CGO)
- **Auth**: JWT magic-link (no passwords)
- **Frontend**: Next.js 15, Tailwind CSS v4, Lucide icons, static export
- **Docs**: Next.js 16 + Fumadocs

---

## Quickstart

### Option 1 — From source (Makefile)

```bash
# Prerequisites: Go 1.25+, Node.js 22+, pnpm 10+
git clone https://github.com/yourorg/policyflow
cd policyflow

# Build everything (frontend → embed → Go binary)
make build

# Run
./build/policyflow
# → http://localhost:8080
```

### Option 2 — Docker

```bash
cd policyflow

# Copy and edit environment variables
cp docker/.env.example docker/.env
# Edit docker/.env — at minimum, set JWT_SECRET

# Build and start
make docker-build
make docker-up
# → http://localhost:8080
```

### Option 3 — Development mode

Run the Go backend and Next.js dev server simultaneously for hot reload:

```bash
# Terminal 1 — Next.js dev server (port 3001)
make dev-frontend

# Terminal 2 — Go backend with dev proxy (port 8080)
make dev-backend
```

---

## Environment Variables

| Variable        | Default                 | Required                  |
| --------------- | ----------------------- | ------------------------- |
| `JWT_SECRET`    | `dev-secret-change-me`  | **Yes** (production)      |
| `DB_PATH`       | `policyflow.db`         | No                        |
| `PORT`          | `8080`                  | No                        |
| `BASE_URL`      | `http://localhost:8080` | Yes (production)          |
| `SMTP_HOST`     | _(empty)_               | No — emails log to stdout |
| `SMTP_PORT`     | `587`                   | No                        |
| `SMTP_USER`     | _(empty)_               | No                        |
| `SMTP_PASSWORD` | _(empty)_               | No                        |
| `SMTP_FROM`     | _(empty)_               | No                        |

---

## Seed Data

On first startup, PolicyFlow creates:

- **Admin user**: `admin@policyflow.local`
- **Staff user**: `staff@policyflow.local`
- **Policy**: "Employee Code of Conduct" (Published, v1.0.0)

In development (no SMTP configured), magic-link emails are printed to the server log:

```
📧 EMAIL (no SMTP configured)
To: admin@policyflow.local
Subject: PolicyFlow — Your login link
Body:
  Click here: http://localhost:8080/api/magic-login?token=eyJ...
```

Copy the URL from the log to log in.

---

## API Reference (Summary)

| Method | Path                            | Auth   | Description                        |
| ------ | ------------------------------- | ------ | ---------------------------------- |
| `POST` | `/api/magic-link`               | Public | Request a login link by email      |
| `GET`  | `/api/magic-login?token=`       | Public | Validate magic link → redirect     |
| `GET`  | `/api/me`                       | Any    | Get current user                   |
| `GET`  | `/api/policies`                 | Any    | List all policies                  |
| `GET`  | `/api/policies/:id`             | Any    | Policy detail + current version    |
| `GET`  | `/api/policies/:id/versions`    | Any    | Full version history               |
| `POST` | `/api/policies/:id/acknowledge` | Any    | Acknowledge current version        |
| `GET`  | `/api/users`                    | Admin  | List all users                     |
| `POST` | `/api/users`                    | Admin  | Create user + send welcome email   |
| `GET`  | `/api/admin/stats`              | Admin  | Dashboard statistics               |
| `POST` | `/api/policies`                 | Admin  | Create new policy                  |
| `PUT`  | `/api/policies/:id`             | Admin  | Update title / status / department |
| `POST` | `/api/policies/:id/versions`    | Admin  | Publish a new version              |

Full API documentation: [apps/docs](apps/docs) or `/docs` when running the docs app.

---

## Workflow

```
Admin                           Staff
  │                               │
  ├─ POST /api/users ─────────────► Receives welcome email
  │                               │
  │                               ├─ Clicks magic link
  │                               ├─ GET /api/magic-login?token=...
  │                               └─ Receives session JWT → logged in
  │
  ├─ POST /api/policies           (creates Draft)
  ├─ POST /api/policies/:id/versions
  ├─ PUT  /api/policies/:id  status=Published
  │
  │                               ├─ GET /api/policies (sees Published policy)
  │                               └─ POST /api/policies/:id/acknowledge
  │
  └─ GET /api/admin/stats (sees acknowledgement count)
```

---

## Project Structure

```
PolicyFlow/
├── Makefile                   ← root build orchestration
├── README.md
├── LICENSE                    ← Apache 2.0
├── build/
│   └── policyflow             ← compiled binary (gitignored)
├── docker/
│   ├── Dockerfile.backend     ← multi-stage build
│   ├── docker-compose.yml
│   └── .env.example
├── apps/
│   ├── app/                   ← Go backend + embedded Next.js frontend
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── Makefile           ← app-level dev commands
│   │   ├── internal/
│   │   │   ├── database/      ← SQLite schema + queries
│   │   │   ├── handlers/      ← HTTP handlers
│   │   │   ├── middleware/    ← JWT auth
│   │   │   ├── email/         ← SMTP mailer
│   │   │   └── seed/          ← initial data
│   │   └── web/               ← Next.js 15 source
│   │       ├── app/           ← App Router pages
│   │       ├── lib/           ← API client, auth helpers
│   │       └── components/    ← shared UI
│   └── docs/                  ← Fumadocs documentation site
│       └── content/docs/      ← MDX documentation pages
└── shared/                    ← (future: shared types, migrations)
```

---

## Roadmap

- [ ] Employee directory import (CSV / SCIM / LDAP)
- [ ] SSO (SAML 2.0, OIDC)
- [ ] Department-scoped policy visibility
- [ ] PDF audit log export
- [ ] Email reminders for unacknowledged policies
- [ ] Rich text editor (Tiptap/ProseMirror)
- [ ] Webhook notifications (Slack, Teams)
- [ ] Multi-tenant support

---

## License

Apache License 2.0 — see [LICENSE](LICENSE).
