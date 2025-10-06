import { describe, it, expect, beforeEach, vi } from 'vitest';
import { fetchRecentReviews, fetchAverageRating } from './api';
import type { Review, AverageRating } from '../types/review';

describe('API Service', () => {
  beforeEach(() => {
    vi.resetAllMocks();
  });

  describe('fetchRecentReviews', () => {
    it('should fetch reviews with default hours parameter', async () => {
      const mockReviews: Review[] = [
        {
          id: '123',
          app_id: '389801252',
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

      const result = await fetchRecentReviews('389801252');

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/reviews?app_id=389801252&hours=48'
      );
      expect(result).toEqual(mockReviews);
    });

    it('should fetch reviews with custom hours parameter', async () => {
      const mockReviews: Review[] = [];

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockReviews,
      });

      await fetchRecentReviews('389801252', 24);

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/reviews?app_id=389801252&hours=24'
      );
    });

    it('should throw error when response is not ok', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      });

      await expect(fetchRecentReviews('389801252')).rejects.toThrow(
        'Failed to fetch reviews'
      );
    });

    it('should throw error when fetch fails', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

      await expect(fetchRecentReviews('389801252')).rejects.toThrow(
        'Network error'
      );
    });

    it('should handle empty review array', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => [],
      });

      const result = await fetchRecentReviews('389801252');

      expect(result).toEqual([]);
    });
  });

  describe('fetchAverageRating', () => {
    it('should fetch average rating with default hours parameter', async () => {
      const mockAverageRating: AverageRating = {
        app_id: '389801252',
        average_rating: 4.5,
        review_count: 100,
        hours: 48,
      };

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockAverageRating,
      });

      const result = await fetchAverageRating('389801252');

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/average-rating?app_id=389801252&hours=48'
      );
      expect(result).toEqual(mockAverageRating);
    });

    it('should fetch average rating with custom hours parameter', async () => {
      const mockAverageRating: AverageRating = {
        app_id: '389801252',
        average_rating: 4.2,
        review_count: 50,
        hours: 24,
      };

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockAverageRating,
      });

      await fetchAverageRating('389801252', 24);

      expect(globalThis.fetch).toHaveBeenCalledWith(
        'http://localhost:8080/api/average-rating?app_id=389801252&hours=24'
      );
    });

    it('should throw error when response is not ok', async () => {
      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 500,
      });

      await expect(fetchAverageRating('389801252')).rejects.toThrow(
        'Failed to fetch average rating'
      );
    });

    it('should throw error when fetch fails', async () => {
      globalThis.fetch = vi.fn().mockRejectedValue(new Error('Network error'));

      await expect(fetchAverageRating('389801252')).rejects.toThrow(
        'Network error'
      );
    });

    it('should handle zero average rating', async () => {
      const mockAverageRating: AverageRating = {
        app_id: '389801252',
        average_rating: 0,
        review_count: 0,
        hours: 48,
      };

      globalThis.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: async () => mockAverageRating,
      });

      const result = await fetchAverageRating('389801252');

      expect(result).toEqual(mockAverageRating);
      expect(result.average_rating).toBe(0);
      expect(result.review_count).toBe(0);
    });
  });
});