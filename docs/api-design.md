# Pote — API Design

## Overview

All REST APIs follow standard conventions with JSON request/response bodies. Authentication is via Bearer JWT tokens in the Authorization header.

## Base URLs (Local Development)

| Service | Base URL |
|---------|----------|
| auth-service | `http://localhost:8081` |
| user-service | `http://localhost:8082` |
| chat-service | `http://localhost:8083` |
| message-service | `http://localhost:8084` |
| realtime-service | `ws://localhost:8085` |
| ai-orchestrator | `http://localhost:8086` |
| notification-service | `http://localhost:8087` |
| secrets-service | `http://localhost:8088` |
| audit-service | `http://localhost:8089` |

## Auth Service Endpoints

```
POST   /api/v1/auth/register        # Email/password registration
POST   /api/v1/auth/login            # Email/password login
POST   /api/v1/auth/refresh          # Refresh access token
POST   /api/v1/auth/logout           # Invalidate refresh token
GET    /api/v1/auth/google           # Initiate Google OAuth flow
GET    /api/v1/auth/google/callback  # Google OAuth callback
```

## User Service Endpoints

```
GET    /api/v1/users/me              # Get current user profile
PUT    /api/v1/users/me              # Update current user profile
GET    /api/v1/users/:id             # Get user by ID
GET    /api/v1/users/search?q=       # Search users by email or name
GET    /api/v1/contacts              # List contacts
POST   /api/v1/contacts              # Add contact
DELETE /api/v1/contacts/:id          # Remove contact
POST   /api/v1/invites               # Send email invite
```

## Chat Service Endpoints

```
POST   /api/v1/chats                 # Create chat (direct or group)
GET    /api/v1/chats                 # List user's chats
GET    /api/v1/chats/:id             # Get chat details
PUT    /api/v1/chats/:id             # Update chat (name, description)
DELETE /api/v1/chats/:id             # Delete chat (owner only)
POST   /api/v1/chats/:id/members     # Add member to group
DELETE /api/v1/chats/:id/members/:uid # Remove member from group
PUT    /api/v1/chats/:id/members/:uid/role # Update member role
```

## Message Service Endpoints

```
GET    /api/v1/chats/:id/messages    # Get message history (cursor-based)
POST   /api/v1/chats/:id/messages    # Send message
PUT    /api/v1/messages/:id          # Edit message
DELETE /api/v1/messages/:id          # Delete message
```

## Secrets Service Endpoints (Internal + User-facing)

```
POST   /api/v1/keys                  # Save API key (encrypted)
GET    /api/v1/keys                  # List saved keys (metadata only, no key values)
DELETE /api/v1/keys/:provider        # Delete API key
GET    /internal/v1/keys/:userId/:provider  # Retrieve decrypted key (internal only)
```

## WebSocket Protocol

Connect: `ws://localhost:8085/ws?token=<jwt>`

See `packages/shared-types/ts/realtime.ts` for the full event type definitions.

## Standard Response Format

```json
{
  "data": { ... },
  "error": null
}
```

```json
{
  "data": null,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid or expired token"
  }
}
```

## Pagination (Cursor-based)

```
GET /api/v1/chats/:id/messages?cursor=<messageId>&limit=50
```

Response includes `nextCursor` and `hasMore` fields.
