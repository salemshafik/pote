"""
Configuration for the AI Orchestrator Service.
Loaded from environment variables using pydantic-settings.
"""

from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    """AI Orchestrator configuration."""

    # Service
    env: str = "development"
    port: int = 8086
    log_level: str = "info"

    # Secrets Service URL (for retrieving user API keys)
    secrets_service_url: str = "http://localhost:8088"

    # Default AI model versions
    openai_default_model: str = "gpt-4o-mini"
    anthropic_default_model: str = "claude-sonnet-4-20250514"
    gemini_default_model: str = "gemini-1.5-pro"

    # GCP Pub/Sub
    gcp_project_id: str = ""
    pubsub_ai_requests_topic: str = "ai-requests"
    pubsub_ai_responses_topic: str = "ai-responses"
    pubsub_ai_requests_subscription: str = "ai-requests-sub"

    class Config:
        env_prefix = ""
        case_sensitive = False


settings = Settings()
