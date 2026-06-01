"""
Base class for LLM providers.
All provider implementations must inherit from this.
"""

from abc import ABC, abstractmethod
from dataclasses import dataclass


@dataclass
class AIRequest:
    """Represents a request to an AI model."""

    chat_id: str
    user_id: str
    model: str  # "chatgpt", "claude", "gemini"
    message: str
    api_key: str
    conversation_history: list[dict] | None = None


@dataclass
class AIResponse:
    """Represents a response from an AI model."""

    chat_id: str
    model: str
    content: str
    tokens_used: int = 0
    error: str | None = None


class BaseLLMProvider(ABC):
    """Abstract base class for LLM provider integrations."""

    @abstractmethod
    async def generate(self, request: AIRequest) -> AIResponse:
        """Generate a response from the LLM provider."""
        ...

    @abstractmethod
    def validate_api_key(self, api_key: str) -> bool:
        """Validate that an API key is properly formatted."""
        ...
