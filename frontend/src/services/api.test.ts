import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fetchRecentReviews } from './api';
import type { Review } from '../types/review';

describe('API Service', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe('fetchRecentReviews', () => {
    it('should fetch reviews with default hours parameter', async () => {
      const mockReviews: Review[] = [
        {
          id: '123',
          app_id: '595068606',
          author: 'Test User',
          content: 'Great app!',
          rating: 5,
          submitted_at: '2025-09-29T10:00:00Z',
          fetched_at: '2025-09-29T11:00:00Z',
        },
      ];

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockReviews,
      });

      const result = await fetchRecentReviews('595068606');

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/reviews?app_id=595068606&hours=48'
      );
      expect(result).toEqual(mockReviews);
    });

    it('should fetch reviews with custom hours parameter', async () => {
      const mockReviews: Review[] = [];

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockReviews,
      });

      await fetchRecentReviews('595068606', 24);

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/reviews?app_id=595068606&hours=24'
      );
    });

    it('should throw error when response is not ok', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      });

      await expect(fetchRecentReviews('595068606')).rejects.toThrow(
        'Failed to fetch reviews'
      );
    });

    it('should throw error when fetch fails', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

      await expect(fetchRecentReviews('595068606')).rejects.toThrow(
        'Network error'
      );
    });

    it('should handle empty review array', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => [],
      });

      const result = await fetchRecentReviews('595068606');

      expect(result).toEqual([]);
    });
  });
});