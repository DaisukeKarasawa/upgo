---
name: react-development
description: Guide React/TypeScript development for the frontend. Use when working with React components, Tailwind CSS, Vite configuration, or frontend state management.
allowed-tools: Read, Write, Edit, Bash, Grep, Glob
---

# React Development Skill

This skill provides guidance for React/TypeScript development in the Upgo frontend.

## Project Structure

```
web/
├── src/
│   ├── components/    # Reusable UI components
│   ├── pages/         # Page components
│   ├── hooks/         # Custom React hooks
│   ├── services/      # API client functions
│   ├── types/         # TypeScript type definitions
│   └── utils/         # Utility functions
├── index.html
├── vite.config.ts
├── tailwind.config.js
└── tsconfig.json
```

## Component Patterns

### Functional Component
```tsx
interface Props {
  title: string;
  onAction: () => void;
}

export const MyComponent: React.FC<Props> = ({ title, onAction }) => {
  return (
    <div className="p-4">
      <h1 className="text-xl font-bold">{title}</h1>
      <button onClick={onAction} className="btn-primary">
        Action
      </button>
    </div>
  );
};
```

### Custom Hook
```tsx
export const usePRData = (id: number) => {
  const [data, setData] = useState<PR | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<Error | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const result = await api.getPR(id);
        setData(result);
      } catch (err) {
        setError(err as Error);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id]);

  return { data, loading, error };
};
```

## Tailwind CSS Patterns

### Common Classes
```tsx
// Layout
<div className="flex items-center justify-between p-4">

// Typography
<h1 className="text-2xl font-bold text-gray-900">

// Buttons
<button className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700">

// Cards
<div className="bg-white rounded-lg shadow p-6">

// Responsive
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
```

## API Integration

### API Client
```tsx
// services/api.ts
const API_BASE = '/api/v1';

export const api = {
  async getPRs(): Promise<PR[]> {
    const res = await fetch(`${API_BASE}/prs`);
    if (!res.ok) throw new Error('Failed to fetch PRs');
    return res.json();
  },

  async syncPR(id: number): Promise<void> {
    const res = await fetch(`${API_BASE}/prs/${id}/sync`, {
      method: 'POST',
    });
    if (!res.ok) throw new Error('Failed to sync PR');
  },
};
```

## TypeScript Types

### Type Definitions
```tsx
// types/pr.ts
export interface PR {
  id: number;
  number: number;
  title: string;
  state: 'open' | 'closed' | 'merged';
  createdAt: string;
  updatedAt: string;
}

export interface PRComment {
  id: number;
  prId: number;
  body: string;
  author: string;
  createdAt: string;
}
```

## Vite Configuration

```ts
// vite.config.ts
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': {
        target: 'http://localhost:8081',
        changeOrigin: true,
      },
    },
  },
});
```

## Testing

### Component Test
```tsx
import { render, screen } from '@testing-library/react';
import { MyComponent } from './MyComponent';

describe('MyComponent', () => {
  it('renders title correctly', () => {
    render(<MyComponent title="Test" onAction={() => {}} />);
    expect(screen.getByText('Test')).toBeInTheDocument();
  });
});
```

## Commands

- Dev server: `npm run dev`
- Build: `npm run build`
- Preview: `npm run preview`
- Lint: `npm run lint`
- Test: `npm test`
