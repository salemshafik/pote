-- ============================================================
-- Pote - Initialize per-service databases for local development
-- This script runs once when the PostgreSQL container is first created.
-- ============================================================

CREATE DATABASE pote_auth;
CREATE DATABASE pote_users;
CREATE DATABASE pote_chats;
CREATE DATABASE pote_messages;
CREATE DATABASE pote_secrets;
CREATE DATABASE pote_audit;
CREATE DATABASE pote_notifications;
