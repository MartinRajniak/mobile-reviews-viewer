let currentAppId = '';
let refreshInterval;

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('en-US', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function renderStars(rating) {
    return '★'.repeat(rating) + '☆'.repeat(5 - rating);
}

async function loadReviews(appId) {
    currentAppId = appId || currentAppId;
    const container = document.getElementById('reviews-list');

    try {
        const response = await fetch(`/api/reviews?app_id=${currentAppId}&hours=48`);

        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }

        const reviews = await response.json();

        if (reviews.length === 0) {
            container.innerHTML = '<div class="no-reviews">No reviews found for this app in the last 2 days.</div>';
            return;
        }

        container.innerHTML = reviews.map(review => `
            <div class="review-card">
                <div class="review-header">
                    <div class="review-author">${review.author}</div>
                    <div class="review-rating">${renderStars(review.rating)}</div>
                </div>
                <div class="review-content">${review.content}</div>
                <div class="review-date">Submitted: ${formatDate(review.submitted_at)}</div>
            </div>
        `).join('');

    } catch (error) {
        console.error('Error loading reviews:', error);
        container.innerHTML = `<div class="error">Failed to load reviews: ${error.message}</div>`;
    }
}

function startAutoRefresh() {
    if (refreshInterval) {
        clearInterval(refreshInterval);
    }
    refreshInterval = setInterval(() => {
        console.log('Auto-refreshing reviews...');
        loadReviews();
    }, 300000); // 5 minutes
}

function initApp(initialAppId) {
    currentAppId = initialAppId;
    loadReviews();
    startAutoRefresh();
}
