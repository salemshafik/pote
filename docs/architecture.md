# Pote — System Architecture

## Overview

Pote is a secure AI-powered chat platform built on a microservices architecture, deployed on Google Cloud Platform (GCP). The system supports real-time 1-on-1 and group messaging, with the ability to invite AI models (ChatGPT, Claude, Gemini) into conversations using `@` commands.

## Architecture Diagram

```
┌──────────────┐      HTTPS       ┌──────────────────┐
│   Frontend   │─────────────────→│   API Gateway /   │
│  (Next.js)   │                  │   Load Balancer   │
└──────┬───────┘                  └────────┬──────────┘
       │ WebSocket                         │ Routes to...
       │                     ┌─────────────┼─────────────────┐
       │               ┌─────▼──┐    ┌─────▼────┐    ┌──────▼───────┐
       │               │  Auth  │    │   Chat   │    │   Message    │
       │               │Service │    │ Service  │    │   Service    │
       │               └───┬────┘    └────┬─────┘    └──┬────┬──────┘
       │                   │              │             │    │
       │              REST │         REST │        Pub/Sub  Redis Pub/Sub
       │                   │              │             │    │
       │               ┌───▼────┐   ┌────▼─────┐  ┌───▼──┐ │
       │               │  User  │   │ Secrets  │  │  AI  │ │
       │               │Service │   │ Service  │  │Orch. │ │
       │               └────────┘   └──────────┘  └──┬───┘ │
       │                                             │      │
       │                  ┌──────────────────────────┼──────┤
       │             Pub/Sub                    Pub/Sub      │
       │                  │                      │           │
       │            ┌─────▼─────┐    ┌──────────▼─┐   ┌────▼────────┐
       │            │   Audit   │    │  Message   │   │ Notification│
       │            │  Service  │    │  Service   │   │   Service   │
       │            └───────────┘    │(write back)│   └─────────────┘
       │                             └────────────┘
       │
  ┌────▼──────────┐
  │   Realtime    │←── Redis Pub/Sub
  │   Service     │
  │  (WebSocket)  │
  └───────────────┘
```

## Services

| Service | Language | Port | Database | Description |
|---------|----------|------|----------|-------------|
| auth-service | Go | 8081 | pote_auth | Authentication, JWT, OAuth |
| user-service | Go | 8082 | pote_users | User profiles, contacts, invites |
| chat-service | Go | 8083 | pote_chats | Chat metadata, groups, RBAC |
| message-service | Go | 8084 | pote_messages | Message CRUD, AI command parsing |
| realtime-service | Go | 8085 | — (Redis) | WebSocket, typing, presence |
| ai-orchestrator | Python | 8086 | — | LLM integration (OpenAI, Anthropic, Google) |
| notification-service | Go | 8087 | pote_notifications | Email invites, push notifications |
| secrets-service | Go | 8088 | pote_secrets | BYOK API key encryption & storage |
| audit-service | Go | 8089 | pote_audit | Append-only security logging |

## Communication Patterns

- **Frontend → Services**: REST via API Gateway
- **Service ↔ Service (sync)**: REST (HTTP/JSON)
- **Service → Service (async)**: GCP Pub/Sub
- **Realtime delivery**: Redis Pub/Sub → WebSocket

## Data Flow: AI Command

1. User sends `@chatgpt What is Go?` in a chat
2. message-service saves the message to its database
3. message-service publishes to Redis → realtime-service pushes to WebSocket clients
4. message-service publishes to Pub/Sub topic `ai-requests`
5. ai-orchestrator subscribes, fetches user's API key from secrets-service
6. ai-orchestrator calls OpenAI API
7. ai-orchestrator publishes response to Pub/Sub topic `ai-responses`
8. message-service subscribes, saves AI response as a new message
9. message-service publishes to Redis → realtime-service pushes to WebSocket clients
