# Pote — AI Integration Guide

## Overview

Pote allows users to invite AI models into any chat (1-on-1 or group) using `@` commands. AI models act as chat participants and respond to messages.

## Supported AI Models

| Command | Provider | Model |
|---------|----------|-------|
| `@chatgpt` | OpenAI | GPT-4o-mini (default) |
| `@claude` | Anthropic | Claude 3.5 Sonnet (default) |
| `@gemini` | Google | Gemini 1.5 Pro (default) |

## Command Lifecycle

### Inviting AI into a Chat

1. User types: `@chatgpt What is Kubernetes?`
2. **message-service** parses the message and detects the `@chatgpt` command
3. If this is the first `@chatgpt` message in the chat:
   - A system message is posted: *"ChatGPT has joined the conversation"*
   - ChatGPT is added to the chat's participant list (as an AI participant)
4. The message is saved and published to Pub/Sub `ai-requests` topic
5. **ai-orchestrator** picks up the request:
   - Retrieves the user's OpenAI API key from **secrets-service**
   - If no key is found → returns an error message asking user to add their API key
   - Calls OpenAI API with the message (and optional conversation context)
6. AI response is published to Pub/Sub `ai-responses` topic
7. **message-service** saves the AI response and pushes it via Redis to WebSocket clients

### Removing AI from a Chat

1. User types: `@chatgpt leave`
2. **message-service** detects the `leave` command
3. ChatGPT is removed from the chat's participant list
4. A system message is posted: *"ChatGPT has left the conversation"*
5. Future `@chatgpt` messages will trigger a fresh "join" event

### Command Parsing Rules

- Commands are **case-insensitive**: `@ChatGPT` = `@chatgpt` = `@CHATGPT`
- The `@` must be at the start of the message or after a space
- Only one AI model per message (first `@` mention wins)
- `leave` must follow the mention directly: `@chatgpt leave`
- Any other text after the mention is treated as the prompt

### Examples

```
@chatgpt What is Go?                    → ChatGPT responds
@claude Explain microservices           → Claude responds
@gemini Summarize this conversation     → Gemini responds
@chatgpt leave                          → ChatGPT leaves the chat
Hey @gemini what do you think?          → Gemini responds (mid-sentence mention)
```

## BYOK (Bring Your Own Key)

- Users must provide their own API keys for each AI provider
- Keys are managed in **Settings → API Keys** in the frontend
- Keys are encrypted and stored via the **secrets-service**
- If a user invokes an AI model without a saved key, they receive:
  > *"You need to add your OpenAI API key in Settings to use @chatgpt"*

## Privacy & Consent

- When AI is first invited into a chat, all participants see a system message
- In group chats, only users with `admin` or `owner` roles can invite AI
- Any member can type `@model leave` to remove the AI
- AI models only see messages that are explicitly sent to them (not chat history by default)

## AI as Chat Participants

When an AI is active in a chat, it appears in the member list with a special indicator:
- 🤖 ChatGPT
- 🤖 Claude
- 🤖 Gemini

AI participants:
- Do NOT count toward group member limits
- Have no role (they cannot be promoted to admin/owner)
- Can be in multiple chats simultaneously
- Only respond when explicitly mentioned with `@`
