export interface Review {
  id: string;
  app_id: string;
  author: string;
  content: string;
  rating: number;
  submitted_at: string;
  fetched_at: string;
}

export interface AverageRating {
  app_id: string;
  average_rating: number;
  review_count: number;
  hours: number;
}