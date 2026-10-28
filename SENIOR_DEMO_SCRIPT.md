# Runtir — AfriOps Document Intelligence & Enterprise RAG Platform Demo Script

## 1. The 30-Second Elevator Pitch
> *"Enterprise document search and AI extraction often fail because documents are ingested synchronously, leaking sensitive multi-tenant data across organizations, or choking under large PDFs. **Runtir (AfriOps)** is a full-stack document intelligence and Retrieval-Augmented Generation (RAG) platform. The backend is written in high-throughput Go (Gin, pgx/v5, sqlc, Goose migrations) with background asynchronous document chunking and vector embedding pipelines. The frontend is a modern React 18 / TypeScript application with JWT-based multi-tenant organization workspaces, granular membership permissions, and real-time document Q&A."*

---

## 2. Senior Architecture Deep Dive

```
[React 18 / TypeScript / Vite Frontend (Port 5173)]
                         |
                 (REST API + JWT Auth)
                         v
   [Go / Gin REST API Server (Port 8080)]
   • JWT Middleware & RBAC Organization Scopes
   • sqlc-generated type-safe PostgreSQL queries
   • Goose versioned schema migrations
                         |
          +--------------+---------------+
          |                              |
          v                              v
[Storage Abstraction]          [Asynchronous Processor]
(Local Disk / S3 Blob)         • PDF / TXT / Markdown chunking
                               • Overlapping window tokenization
                               • OpenAI / Gemini / Dummy Embeddings
                                         |
                                         v
                         [PostgreSQL + pgvector / Full-Text RAG]
```

---

## 3. How to Run & Demo

### A. Run the Frontend (Quickest UI Walkthrough)
```powershell
Set-Location "C:\Users\HP\Documents\Exquisapps\ID's\lekeOmotayo\runtir\frontend"
npm run dev
```
Open **`http://localhost:5173`** in your browser.

### B. Run the Go Backend (When PostgreSQL is running)
```powershell
Set-Location "C:\Users\HP\Documents\Exquisapps\ID's\lekeOmotayo\runtir\backend"
go run cmd/main.go cmd/api.go
```
Backend URL: **`http://localhost:8080`**

### Demo Flow:
1. **Multi-Tenant Workspaces**:
   - Register or Sign In.
   - Show Tenant Switcher: switching organizations isolates documents and queries so Tenant A can never view or search Tenant B's documents.
2. **Document Ingestion Pipeline**:
   - Upload a sample corporate policy or technical manual (PDF or TXT).
   - Point out the asynchronous background status transition (`pending` → `processing` → `indexed`).
   - Explain how sliding-window chunking preserves semantic context across paragraph boundaries.
3. **Retrieval-Augmented Generation (RAG)**:
   - Ask a question: *"What is the policy for reimbursement under travel expenses?"*.
   - Show the grounded answer with citation chunk references.
   - Point out that without paid API keys, the built-in provider gracefully returns clearly labelled mock responses so demos never fail during interviews.

---

## 4. Senior Defense Questions & Answers

### Q: "Why sqlc over a traditional ORM like GORM in Go?"
> **Answer:** *"ORMs like GORM introduce reflection overhead, runtime type coercion, and obscure the generated SQL queries. `sqlc` compiles raw, hand-optimized SQL into compile-time type-safe Go structs and functions. Any schema or query mismatch is caught during build time, completely eliminating SQL injection and runtime type mismatches while providing maximum query execution speed with `pgx/v5`."*

### Q: "How does Runtir handle document processing without blocking user requests?"
> **Answer:** *"Document upload handlers only validate the file header and persist the raw bytes to storage, returning a `202 Accepted` immediately. Ingestion (text extraction, chunking, and embedding generation) runs in a background goroutine worker queue, updating the document state in the database once indexing finishes."*
