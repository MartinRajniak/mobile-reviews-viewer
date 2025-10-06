import { useState, useEffect } from 'react';
import type { Review } from '../types/review';
import { fetchRecentReviews } from '../services/api';
import { ReviewCard } from './ReviewCard';

interface ReviewListProps {
  appId: string;
  hours?: number;
}

export const ReviewList: React.FC<ReviewListProps> = ({ appId, hours = 48 }) => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadReviews = async () => {
      setLoading(true);
      setError(null);

      try {
        const data = await fetchRecentReviews(appId, hours);
        setReviews(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    loadReviews();

    // Auto-refresh every 5 minutes
    const interval = setInterval(loadReviews, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [appId, hours]);

  if (loading) {
    return <div className="loading">Loading reviews...</div>;
  }

  if (error) {
    return <div className="error">Error: {error}</div>;
  }

  return (
    <div className="review-list">
      <h2>Recent Reviews (Last {hours} Hours)</h2>
      {reviews.length === 0 ? (
        <p className="no-reviews">No reviews found in the last {hours} hours.</p>
      ) : (
        <>
          <p className="review-count">{reviews.length} reviews found</p>
          <div className="reviews">
            {reviews.map((review) => (
              <ReviewCard key={review.id} review={review} />
            ))}
          </div>
        </>
      )}
    </div>
  );
};