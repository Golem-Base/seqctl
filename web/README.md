# Seqctl Web Interface

A React-based web interface for managing Optimism conductor sequencer clusters.

## Technology Stack

- **Runtime**: Bun
- **Framework**: React 19 with TypeScript
- **UI Components**: shadcn/ui
- **Styling**: Tailwind CSS v4
- **State Management**: Zustand
- **Build Tool**: Bun's native bundler

## Development

### Prerequisites

- Bun installed globally
- Go backend running on port 8080

### Install Dependencies

```bash
bun install
```

### Development Mode

Start the development server with hot module replacement:

```bash
bun dev
```

The app will be available at http://localhost:3000 and will proxy API requests
to http://localhost:8080.

### Production Build

Build the app for production:

```bash
bun run build
```

This creates optimized files in the `dist/` directory that will be embedded into
the Go binary.

## Project Structure

```
src/
├── api/              # API client and TypeScript types
├── components/       # React components
│   ├── ui/          # shadcn/ui components
│   ├── features/    # Feature components (NetworkList, NetworkDetail)
│   ├── layout/      # Layout components (Header, Layout)
│   └── modals/      # Modal dialogs for sequencer operations
├── hooks/           # Custom React hooks
├── stores/          # Zustand state management
└── lib/             # Utilities
```

## Key Features

- Real-time network and sequencer monitoring
- Sequencer operations:
  - Pause/Resume conductor
  - Transfer/Resign leadership
  - Override leader status
  - Force sequencer active
  - Manage cluster membership
- Auto-refresh with configurable intervals
- Toast notifications for actions
- Responsive design

## API Integration

The app communicates with the Go backend through REST APIs:

- `/api/v1/networks` - Network management
- `/api/v1/sequencers` - Sequencer operations
- `/ws` - WebSocket for real-time updates (planned)

During development, the Bun server (src/index.tsx) proxies API requests to avoid
CORS issues.

## State Management

Uses Zustand for state management with two main stores:

- `networkStore` - Manages network/sequencer data and fetching
- `actionStore` - Handles modal states and complex operations

## Build Integration

The production build is embedded into the Go binary using Go's embed package.
The build process:

1. `bun run build` creates optimized files in `dist/`
2. `just build-web` copies files to `pkg/server/dist/`
3. Go embeds the files at compile time
4. The Go server serves the SPA for all non-API routes
