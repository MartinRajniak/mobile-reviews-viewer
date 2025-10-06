import type { Review, AverageRating } from '../types/review';

const API_BASE_URL = 'http://localhost:8080/api';

export const fetchRecentReviews = async (
  appId: string,
  hours: number = 48
): Promise<Review[]> => {
  const response = await fetch(
    `${API_BASE_URL}/reviews?app_id=${appId}&hours=${hours}`
  );

  if (!response.ok) {
    throw new Error('Failed to fetch reviews');
  }

  return response.json();
};

export const fetchAverageRating = async (
  appId: string,
  hours: number = 48
): Promise<AverageRating> => {
  const response = await fetch(
    `${API_BASE_URL}/average-rating?app_id=${appId}&hours=${hours}`
  );

  if (!response.ok) {
    throw new Error('Failed to fetch average rating');
  }

  return response.json();
};