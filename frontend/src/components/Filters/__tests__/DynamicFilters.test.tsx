import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { QueryClient, QueryClientProvider } from 'react-query';
import DynamicFilters from '../DynamicFilters';

// Mock localStorage
const localStorageMock = {
  getItem: jest.fn(() => 'mock-token'),
  setItem: jest.fn(),
  removeItem: jest.fn(),
};
global.localStorage = localStorageMock as any;

// Mock API server
const server = setupServer(
  rest.get('http://localhost:8081/api/v1/filters/resource-types', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: [
          { key: 'ec2', label: 'EC2', count: 45 },
          { key: 'rds', label: 'RDS', count: 12 },
        ],
      })
    );
  }),
  rest.get('http://localhost:8081/api/v1/filters/tags', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: {
          tag_keys: [
            { key: 'Environment', count: 50 },
            { key: 'Team', count: 30 },
          ],
          tag_values: {
            Environment: ['prod', 'dev', 'staging'],
            Team: ['backend', 'frontend'],
          },
        },
      })
    );
  }),
  rest.get('http://localhost:8081/api/v1/filters/services', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: [
          { key: 'EC2', label: 'EC2', total_cost: 1000 },
          { key: 'RDS', label: 'RDS', total_cost: 500 },
        ],
      })
    );
  }),
  rest.get('http://localhost:8081/api/v1/filters/regions', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: [
          { key: 'us-east-1', label: 'us-east-1', count: 30 },
          { key: 'us-west-2', label: 'us-west-2', count: 15 },
        ],
      })
    );
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: false,
    },
  },
});

const renderWithProviders = (ui: React.ReactElement) => {
  return render(
    <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider>
  );
};

describe('DynamicFilters Component', () => {
  test('renders filter sections', async () => {
    const onFilterChange = jest.fn();
    renderWithProviders(<DynamicFilters onFilterChange={onFilterChange} />);

    await waitFor(() => {
      expect(screen.getByText(/filters/i)).toBeInTheDocument();
    });
  });

  test('loads and displays resource types', async () => {
    const onFilterChange = jest.fn();
    renderWithProviders(<DynamicFilters onFilterChange={onFilterChange} />);

    await waitFor(() => {
      expect(screen.getByText(/resource types/i)).toBeInTheDocument();
      expect(screen.getByText(/EC2/i)).toBeInTheDocument();
      expect(screen.getByText(/RDS/i)).toBeInTheDocument();
    });
  });

  test('toggles resource type filter', async () => {
    const user = userEvent.setup();
    const onFilterChange = jest.fn();
    renderWithProviders(<DynamicFilters onFilterChange={onFilterChange} />);

    await waitFor(() => {
      expect(screen.getByText(/EC2/i)).toBeInTheDocument();
    });

    const ec2Button = screen.getByText(/EC2/i);
    await user.click(ec2Button);

    // Wait for debounced callback
    await waitFor(
      () => {
        expect(onFilterChange).toHaveBeenCalled();
      },
      { timeout: 500 }
    );
  });

  test('clears all filters', async () => {
    const user = userEvent.setup();
    const onFilterChange = jest.fn();
    renderWithProviders(<DynamicFilters onFilterChange={onFilterChange} />);

    await waitFor(() => {
      expect(screen.getByText(/clear all/i)).toBeInTheDocument();
    });

    const clearButton = screen.getByText(/clear all/i);
    await user.click(clearButton);

    await waitFor(
      () => {
        expect(onFilterChange).toHaveBeenCalledWith({
          resourceTypes: [],
          tags: {},
          services: [],
          accounts: [],
          regions: [],
        });
      },
      { timeout: 500 }
    );
  });
});

