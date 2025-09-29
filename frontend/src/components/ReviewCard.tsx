import type { Review } from '../types/review';

interface ReviewCardProps {
  review: Review;
}

export const ReviewCard: React.FC<ReviewCardProps> = ({ review }) => {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  const renderStars = (rating: number) => {
    return '★'.repeat(rating) + '☆'.repeat(5 - rating);
  };

  return (
    <div className="review-card">
      <div className="review-header">
        <span className="author">{review.author}</span>
        <span className="rating" title={`${review.rating}/5`}>
          {renderStars(review.rating)}
        </span>
      </div>
      <p className="review-content">{review.content}</p>
      <div className="review-footer">
        <span className="date">{formatDate(review.submitted_at)}</span>
      </div>
    </div>
  );
};