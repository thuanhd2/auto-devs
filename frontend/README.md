# Auto-Devs Frontend

Frontend application for Auto-Devs - an AI-powered developer task automation system. Built with React, TypeScript, and ShadcnUI.

## Features

- Light/dark mode toggle
- Responsive design
- Accessible UI components
- Built-in sidebar navigation
- Project management interface
- Settings management
- Modern React architecture with TypeScript

## Tech Stack

**UI Framework:** React 19 with TypeScript

**UI Components:** [ShadcnUI](https://ui.shadcn.com) (TailwindCSS + RadixUI)

**Build Tool:** [Vite](https://vitejs.dev/)

**Routing:** [TanStack Router](https://tanstack.com/router/latest)

**State Management:** [Zustand](https://github.com/pmndrs/zustand)

**Data Fetching:** [TanStack Query](https://tanstack.com/query/latest)

**HTTP Client:** [Axios](https://axios-http.com/)

**Styling:** [TailwindCSS](https://tailwindcss.com/)

**Type Checking:** [TypeScript](https://www.typescriptlang.org/)

**Linting/Formatting:** [ESLint](https://eslint.org/) & [Prettier](https://prettier.io/)

**Icons:** [Tabler Icons](https://tabler.io/icons) & [Lucide React](https://lucide.dev/)

## Getting Started

### Prerequisites

- Node.js 22.12.0 or higher
- npm or pnpm package manager

### Installation

1. Clone the repository and navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
# or
pnpm install
```

3. Create environment configuration:
```bash
cp .env.example .env.local
```

4. Update the environment variables in `.env.local`:
```bash
# API Configuration
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_WS_BASE_URL=ws://localhost:8080/ws

# Development
VITE_DEV_MODE=true
```

### Development

Start the development server:
```bash
npm run dev
```

The application will be available at `http://localhost:5173`

### Build

Build for production:
```bash
npm run build
```

Preview the production build:
```bash
npm run preview
```

### Code Quality

Run linting:
```bash
npm run lint
```

Format code:
```bash
npm run format
```

Check formatting:
```bash
npm run format:check
```

## Project Structure

```
src/
├── components/         # Reusable UI components
│   ├── ui/            # ShadcnUI components
│   └── layout/        # Layout components
├── config/            # Configuration files
├── context/           # React context providers
├── features/          # Feature-specific components
├── hooks/             # Custom React hooks
├── lib/               # Utility libraries
├── routes/            # Route components (TanStack Router)
├── stores/            # Zustand stores
└── utils/             # Utility functions
```

## API Integration

The frontend communicates with the Auto-Devs backend API:

- **Base URL:** Configured via `VITE_API_BASE_URL`
- **WebSocket:** Real-time updates via `VITE_WS_BASE_URL`
- **Endpoints:** Projects, tasks, and execution management

## Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint
- `npm run format` - Format code with Prettier
- `npm run format:check` - Check code formatting

## License

This project is part of the Auto-Devs system. See the main project repository for licensing information.