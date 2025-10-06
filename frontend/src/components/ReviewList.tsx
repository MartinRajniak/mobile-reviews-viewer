import { useState, useEffect } from 'react';
import type { Review, AverageRating } from '../types/review';
import { fetchRecentReviews, fetchAverageRating } from '../services/api';
import { ReviewCard } from './ReviewCard';

interface ReviewListProps {
  appId: string;
  hours?: number;
}

export const ReviewList: React.FC<ReviewListProps> = ({ appId, hours = 48 }) => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [averageRating, setAverageRating] = useState<AverageRating | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadData = async () => {
      setLoading(true);
      setError(null);

      try {
        const [reviewsData, averageData] = await Promise.all([
          fetchRecentReviews(appId, hours),
          fetchAverageRating(appId, hours)
        ]);
        setReviews(reviewsData);
        setAverageRating(averageData);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    loadData();

    // Auto-refresh every 5 minutes
    const interval = setInterval(loadData, 5 * 60 * 1000);
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
      {averageRating && averageRating.review_count > 0 && (
        <div className="average-rating">
          <strong>Average Rating:</strong> ⭐ {averageRating.average_rating.toFixed(1)} ({averageRating.review_count} reviews)
        </div>
      )}
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