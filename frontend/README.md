# Pote Frontend

Next.js web application for the Pote AI-powered chat platform.

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **WebSocket**: Native WebSocket API
- **Package Manager**: pnpm

## Getting Started

```bash
# Install dependencies
pnpm install

# Start development server
pnpm dev

# Build for production
pnpm build

# Start production server
pnpm start
```

## Project Structure (Planned)

```
frontend/
├── src/
│   ├── app/                    # Next.js App Router pages
│   │   ├── (auth)/             # Auth pages (login, register)
│   │   ├── (chat)/             # Chat pages (main UI)
│   │   ├── settings/           # User settings (API keys, profile)
│   │   └── layout.tsx
│   ├── components/             # Reusable UI components
│   │   ├── ui/                 # Base UI components (Button, Input, etc.)
│   │   ├── chat/               # Chat-specific components
│   │   ├── auth/               # Auth-specific components
│   │   └── layout/             # Layout components (Sidebar, Header)
│   ├── hooks/                  # Custom React hooks
│   ├── lib/                    # Utility functions and API clients
│   ├── stores/                 # Zustand state stores
│   ├── types/                  # TypeScript type definitions
│   └── styles/                 # Global styles
├── public/                     # Static assets
├── next.config.ts
├── tailwind.config.ts
├── tsconfig.json
└── package.json
```

## Setup

The frontend will be scaffolded with `pnpm create next-app` when we reach Phase 3 (Frontend Implementation).
