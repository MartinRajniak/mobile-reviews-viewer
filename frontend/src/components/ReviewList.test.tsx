import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { ReviewList } from './ReviewList';
import * as api from '../services/api';

vi.mock('../services/api');

describe('ReviewList', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  it('should call fetchRecentReviews with default hours prop', async () => {
    const mockFetch = vi.spyOn(api, 'fetchRecentReviews').mockResolvedValue([]);

    render(<ReviewList appId="123" hours={48} />);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('123', 48);
    });
  });

  it('should call fetchRecentReviews with custom hours prop', async () => {
    const mockFetch = vi.spyOn(api, 'fetchRecentReviews').mockResolvedValue([]);

    render(<ReviewList appId="123" hours={24} />);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('123', 24);
    });
  });

  it('should display dynamic time window text based on hours', async () => {
    vi.spyOn(api, 'fetchRecentReviews').mockResolvedValue([]);

    const { rerender } = render(<ReviewList appId="123" hours={24} />);

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Last 24 Hours/i })).toBeInTheDocument();
    });

    rerender(<ReviewList appId="123" hours={72} />);

    await waitFor(() => {
      expect(screen.getByRole('heading', { name: /Last 72 Hours/i })).toBeInTheDocument();
    });
  });

  it('should refetch reviews when hours prop changes', async () => {
    const mockFetch = vi.spyOn(api, 'fetchRecentReviews').mockResolvedValue([]);

    const { rerender } = render(<ReviewList appId="123" hours={48} />);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('123', 48);
    });

    rerender(<ReviewList appId="123" hours={24} />);

    await waitFor(() => {
      expect(mockFetch).toHaveBeenCalledWith('123', 24);
      expect(mockFetch).toHaveBeenCalledTimes(2);
    });
  });

  it('should display correct message when no reviews found', async () => {
    vi.spyOn(api, 'fetchRecentReviews').mockResolvedValue([]);

    render(<ReviewList appId="123" hours={12} />);

    await waitFor(() => {
      expect(screen.getByText(/No reviews found in the last 12 hours/i)).toBeInTheDocument();
    });
  });
});
