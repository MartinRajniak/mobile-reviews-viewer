package com.example

import com.example.reviews.Review
import com.example.reviews.ReviewsFetcher
import com.example.reviews.ReviewsStorage
import kotlinx.serialization.json.Json

/**
 * Shared JSON configuration for tests
 */
val testJson = Json { ignoreUnknownKeys = true }

/**
 * Fake ReviewsStorage implementation for testing
 */
class FakeReviewsStorage(
    private val reviews: List<Review> = emptyList(),
    private val recentReviews: List<Review> = emptyList()
) : ReviewsStorage {
    val savedReviews = mutableListOf<Review>()

    override fun saveReviews(reviews: List<Review>) {
        savedReviews.addAll(reviews)
    }

    override fun getAllReviews(): List<Review> = reviews + savedReviews

    override suspend fun loadState() {}

    override suspend fun saveState() {}

    override fun getRecentReviews(appId: String?, since: kotlin.time.Instant) = recentReviews
}

/**
 * Fake ReviewsFetcher implementation for testing
 */
class FakeReviewsFetcher(
    private val shouldFail: Boolean = false,
    private val failForAppIds: Set<String> = emptySet()
) : ReviewsFetcher {
    var fetchCount = 0
    val fetchedAppIds = mutableListOf<String>()

    override suspend fun fetchReviews(appId: String): List<Review> {
        fetchCount++
        fetchedAppIds.add(appId)
        if (shouldFail || appId in failForAppIds) {
            throw Exception("Failed to fetch reviews for $appId")
        }
        return emptyList()
    }
}
