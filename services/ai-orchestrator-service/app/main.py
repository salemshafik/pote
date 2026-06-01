"""
Pote AI Orchestrator Service

This FastAPI service handles LLM integration for Pote. It:
1. Subscribes to the 'ai-requests' Pub/Sub topic
2. Retrieves the user's decrypted API key from the secrets-service
3. Calls the appropriate LLM provider (OpenAI, Anthropic, Gemini)
4. Publishes the AI response to the 'ai-responses' Pub/Sub topic

Supported AI models:
- @chatgpt  → OpenAI (GPT-4o / GPT-4o-mini)
- @claude   → Anthropic (Claude 3.5 Sonnet)
- @gemini   → Google (Gemini 1.5 Pro)
"""

import logging
from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

from app.config import settings

logger = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Application startup and shutdown lifecycle."""
    logger.info("Starting AI Orchestrator Service on port %s", settings.port)
    # TODO: Initialize Pub/Sub subscriber for ai-requests topic
    # TODO: Initialize HTTP client for secrets-service
    yield
    # TODO: Cleanup Pub/Sub connections
    logger.info("Shutting down AI Orchestrator Service")


app = FastAPI(
    title="Pote AI Orchestrator",
    description="LLM integration service for Pote chat platform",
    version="0.1.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],  # Restrict in production
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)


@app.get("/health")
async def health_check():
    """Health check endpoint for Cloud Run."""
    return {"status": "healthy", "service": "ai-orchestrator"}
