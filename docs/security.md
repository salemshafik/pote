# Pote — Security Standards

## Overview

Pote implements enterprise-grade security across all layers of the application.

## Authentication & Authorization

### JWT / OIDC
- Access tokens: Short-lived (15 minutes), signed with HMAC-SHA256
- Refresh tokens: Long-lived (7 days), stored server-side
- All service-to-service calls validate JWT at the middleware level
- Google OAuth 2.0 / OIDC for social login

### Role-Based Access Control (RBAC)
- **User roles**: `user`, `admin`
- **Chat member roles**: `member`, `admin`, `owner`
- Only chat owners/admins can invite AI models or add members to group chats
- Role checks enforced at the service layer

## Encryption

### In Transit
- All external traffic over HTTPS/TLS 1.3
- Internal service-to-service traffic over HTTPS within GCP VPC

### At Rest
- PostgreSQL databases encrypted at rest (Cloud SQL default)
- User API keys encrypted via GCP KMS before storage
- Secrets stored in GCP Secret Manager

## BYOK (Bring Your Own Key) Security

1. User submits API key via frontend → HTTPS to secrets-service
2. secrets-service encrypts key using GCP KMS (AES-256-GCM)
3. Encrypted key stored in pote_secrets database
4. Key is **never** returned to the frontend after initial save
5. Only ai-orchestrator can request decryption (service-to-service, no user-facing endpoint)
6. All key access is audit-logged

## API Security

- Rate limiting at API Gateway level
- Input validation on all endpoints
- CORS restricted to allowed origins in production
- Request ID tracking across all services

## Audit Logging

- All security-relevant events published to Pub/Sub `audit-events` topic
- Append-only storage in audit-service database
- Events include: login attempts, API key operations, AI invocations, role changes
- No DELETE operations on audit records

## Privacy Controls

- Users are notified when AI is invited into a conversation
- System messages indicate when AI joins/leaves a chat
- Users can remove AI from a chat at any time (`@model leave`)
- API keys can be deleted by the user at any time
