# 🤖 Secure AI-Powered Chat Platform

A scalable, and responsive chat application that seamlessly integrates human conversations with AI models. Built on a microservices architecture and deployed on Google Cloud Platform (GCP), this platform allows users to chat in real-time, react with emojis, import contacts, and invite AI assistants (like ChatGPT, Claude, and Gemini) directly into their conversations using `@` commands.

## ✨ Key Features

- **Real-Time Communication**: Instant messaging, typing indicators, read receipts, and user presence powered by Go and WebSockets.
- **AI Integration**: Invite AI models into 1-on-1 or group chats using commands (e.g., `@ChatGPT Can you summarize this?`).
- **Bring Your Own Key (BYOK)**: Users can securely add their own OpenAI, Anthropic, or Gemini API keys.
- **Enterprise-Grade Security**: Strict secrets management via GCP Secret Manager, in-transit/at-rest encryption, zero-exposure key policies, and full audit logging.
- **Rich Interactions**: Message emoji reactions and rich interactive UI.
- **Viral Growth & Networking**: Import contacts via Google People API, send email invites, or share via native Web Share API.
- **Granular Group Controls**: Role-based access control (RBAC) ensuring only authorized users can invite AI or add members to groups.

## 🛠 Tech Stack

- **Frontend**: Next.js (React), Tailwind CSS, Zustand, Socket.io-client / native WebSockets, **`pnpm`** (Package Manager).
- **Backend Services**: Go (Golang) with `chi`/`gin`, PostgreSQL, GCP Memorystore (Redis).
- **AI Orchestrator**: Python (FastAPI), Official LLM SDKs (OpenAI, Anthropic, Google GenAI).
- **Infrastructure**: Google Cloud Platform (Cloud Run, Cloud SQL, Pub/Sub, Secret Manager, API Gateway).
- **IaC & Deployment**: Terraform, GitHub Actions.

## 📂 Repository Structure

This project follows a monorepo structure. All microservices, frontend applications, shared packages, and infrastructure code live here.

├── frontend/                     # Next.js web application (Uses pnpm)
├── services/                     # Backend Microservices
│   ├── auth-service/             # Authentication, JWT handling, and OAuth
│   ├── user-service/             # User profiles, contacts, and invites
│   ├── chat-service/             # Chat metadata, groups, and RBAC
│   ├── message-service/          # Message CRUD and history
│   ├── realtime-service/         # Go WebSocket server + Redis Pub/Sub
│   ├── ai-orchestrator-service/  # Python FastAPI service for LLM integration
│   ├── notification-service/     # Push notifications and email invites
│   ├── secrets-service/          # KMS/GCP Secret Manager abstraction
│   └── audit-service/            # Append-only security and event logging
├── packages/                     # Shared internal libraries
│   ├── shared-types/             # Protobufs, OpenAPI specs, TS interfaces
│   ├── logger/                   # Standardized structured logging
│   ├── config/                   # Environment variable parsing
│   └── auth-utils/               # JWT validation middleware shared across Go services
├── infra/                        # Infrastructure as Code
│   ├── terraform/                
│   │   ├── modules/              # Reusable TF modules (Cloud Run, PubSub, etc.)
│   │   └── envs/
│   │       ├── dev/              # Development environment state
│   │       └── prod/             # Production environment state
├── docs/                         # Project Documentation
│   ├── architecture.md           # System design & data flow diagrams
│   ├── security.md               # Encryption, RBAC, and BYOK standards
│   ├── api-design.md             # REST/WebSocket API contracts
│   └── ai-integration.md         # Detailed lifecycle of the `@` AI commands
└── .github/
    └── workflows/                # CI/CD pipelines (Test, Build, Deploy to GCP)