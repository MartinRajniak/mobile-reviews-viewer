import type { Review } from '../types/review';

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