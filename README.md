# Runtir — AfriOps Document Intelligence Platform

[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://postgresql.org)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://docker.com)
[![Render](https://img.shields.io/badge/Render-Deploy%20Ready-46E3B7?style=flat)](https://render.com)

> A full-stack, AI-powered document intelligence and operations platform. The **Go backend (Gin + pgx)** handles secure document ingestion, multi-model RAG question-answering (OpenAI / Gemini), organisation membership, JWT authentication, and dashboard analytics. The **React + Vite frontend** provides a polished SPA for all operations. PostgreSQL migrations are managed with `goose`.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────────────────┐
│                       RUNTIR PLATFORM ARCHITECTURE                       │
└──────────────────────────────────────────────────────────────────────────┘

  ┌─────────────────────┐         ┌───────────────────────────────────────┐
  │  React + Vite SPA   │◄───────►│       Go Gin REST API (Port 8080)    │
  │  (TypeScript)       │  HTTP   │                                       │
  └─────────────────────┘         │  ┌──────────┐  ┌──────────────────┐  │
                                  │  │  Auth    │  │  Document        │  │
                                  │  │  Module  │  │  Ingestion       │  │
                                  │  │  JWT/HS  │  │  + Storage       │  │
                                  │  └──────────┘  └──────────────────┘  │
                                  │  ┌──────────┐  ┌──────────────────┐  │
                                  │  │  AI/RAG  │  │  Organisation    │  │
                                  │  │  OpenAI  │  │  & Membership    │  │
                                  │  │  Gemini  │  │  Module          │  │
                                  │  └──────────┘  └──────────────────┘  │
                                  │  ┌──────────┐  ┌──────────────────┐  │
                                  │  │Dashboard │  │  Background      │  │
                                  │  │Analytics │  │  Processor       │  │
                                  │  └──────────┘  └──────────────────┘  │
                                  └───────────────────────┬───────────────┘
                                                          │
                                              ┌───────────▼──────────┐
                                              │   PostgreSQL 16       │
                                              │   (pgx/v5 pool)       │
                                              └──────────────────────┘
```

---

## Features

- **JWT Authentication**: Secure cookie-based sessions with HS256 signing, configurable expiry, and CORS-aware middleware.
- **Document Management**: Upload, store, and manage documents with local or cloud storage abstraction.
- **Multi-Model AI/RAG**: Question-answering over ingested documents using OpenAI GPT-4o-mini or Google Gemini 1.5 Flash (switchable via `AI_PROVIDER` env var).
- **Background Processor**: Asynchronous document processing pipeline decoupled from request/response cycle.
- **Organisation & Membership**: Multi-tenant organisation support with member roles.
- **Dashboard Analytics**: Aggregated stats across documents, AI queries, and usage.
- **Database Migrations**: Schema managed with `goose` and `sqlc`-generated type-safe queries.
- **Graceful Shutdown**: SIGTERM/SIGINT handling with 15-second drain window.

---

## Project Structure

```
runtir/
├── backend/                    # Go Gin REST API
│   ├── cmd/
│   │   ├── main.go             # Entrypoint
│   │   └── api.go              # Server init, routes, graceful shutdown
│   ├── internal/
│   │   ├── config/             # Environment config loader
│   │   ├── db/                 # sqlc-generated queries + migrations
│   │   ├── middleware/         # JWT auth, CORS, request logger
│   │   ├── modules/
│   │   │   ├── ai/             # RAG & AI Q&A handlers
│   │   │   ├── auth/           # Login, register, token refresh
│   │   │   ├── dashboard/      # Analytics endpoints
│   │   │   ├── document/       # Upload, list, delete documents
│   │   │   └── organisation/   # Org & member management
│   │   ├── processor/          # Async document processing
│   │   ├── storage/            # Local file storage abstraction
│   │   └── utils/              # Response helpers
│   ├── uploads/                # Local document storage volume
│   ├── .env.example            # Environment variable reference
│   ├── Dockerfile              # Production container
│   ├── go.mod / go.sum
│   └── sqlc.yaml
│
└── frontend/                   # React 18 + Vite + TypeScript SPA
    ├── src/
    ├── public/
    ├── package.json
    ├── vite.config.ts
    └── tsconfig.json
```

---

## Environment Variables

Copy `backend/.env.example` to `backend/.env` and fill in your values:

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | Server listen port |
| `APP_ENV` | `development` | `development` or `production` |
| `DATABASE_URL` | — | PostgreSQL connection string |
| `JWT_SECRET` | — | **Required** — 32+ char secret key |
| `JWT_EXPIRY_HOURS` | `24` | Token lifetime in hours |
| `COOKIE_SECURE` | auto | `true` in production |
| `UPLOAD_DIR` | `./uploads` | Local storage path |
| `MAX_UPLOAD_SIZE_MB` | `25` | Max file size |
| `AI_PROVIDER` | `openai` | `openai` or `gemini` |
| `OPENAI_API_KEY` | — | Required if `AI_PROVIDER=openai` |
| `OPENAI_MODEL` | `gpt-4o-mini` | Chat model |
| `OPENAI_EMBEDDING_MODEL` | `text-embedding-3-small` | Embeddings model |
| `GEMINI_API_KEY` | — | Required if `AI_PROVIDER=gemini` |
| `GEMINI_MODEL` | `gemini-1.5-flash` | Gemini chat model |

---

## Local Development

### Prerequisites
- Go 1.24+
- Node.js 20+ and npm
- PostgreSQL 16 (local or managed)

### Backend

```bash
cd backend

# Copy env file and fill in your values
cp .env.example .env

# Install Go dependencies
go mod download

# Run database migrations
go run ./cmd goose up   # or: goose -dir internal/db/migrations postgres "$DATABASE_URL" up

# Start the API server (hot-reload with air)
go run cmd/main.go cmd/api.go
# Server starts at http://localhost:8080
```

### Frontend

```bash
cd frontend

# Install dependencies
npm install

# Start Vite dev server with hot reload
npm run dev
# Opens at http://localhost:5173 (proxies /api to :8080)
```

### Build Frontend for Production

```bash
cd frontend
npm run build
# Output in frontend/dist/ — served by backend or CDN
```

---

## API Reference

All endpoints are prefixed with `/api/v1`. JWT token is set as an HttpOnly cookie on login.

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `GET` | `/health` | — | Health check |
| `POST` | `/api/v1/auth/register` | — | Create account |
| `POST` | `/api/v1/auth/login` | — | Login, sets JWT cookie |
| `POST` | `/api/v1/auth/logout` | ✓ | Clear JWT cookie |
| `GET` | `/api/v1/auth/me` | ✓ | Current user profile |
| `POST` | `/api/v1/documents` | ✓ | Upload document |
| `GET` | `/api/v1/documents` | ✓ | List documents |
| `DELETE` | `/api/v1/documents/:id` | ✓ | Delete document |
| `POST` | `/api/v1/ai/ask` | ✓ | AI Q&A over documents |
| `GET` | `/api/v1/dashboard` | ✓ | Usage analytics |
| `POST` | `/api/v1/organisations` | ✓ | Create organisation |
| `GET` | `/api/v1/organisations` | ✓ | List organisations |

---

## Docker & Docker Compose

```bash
# From the runtir root directory
docker-compose up --build
```

> ⚠️ A `docker-compose.yml` can be added at the `runtir/` root combining the backend service and a PostgreSQL container.

---

## Render Deployment

This project includes `render.yaml` at the repo root (the `runtir/` folder) that deploys:
- **`runtir-backend-api`** — Go backend web service with auto-provisioned env vars
- **`runtir-frontend`** — React SPA as a static site
- **`runtir-postgres`** — Managed PostgreSQL database

### Deploy in 3 steps:
1. Push the `runtir/` folder to GitHub as a standalone repo.
2. Go to [dashboard.render.com](https://dashboard.render.com) → **New + → Blueprint**
3. Select your repo → **Apply**

After deploy, go to `runtir-backend-api` → **Environment** and add:
- `OPENAI_API_KEY` (or `GEMINI_API_KEY` if using Gemini)
