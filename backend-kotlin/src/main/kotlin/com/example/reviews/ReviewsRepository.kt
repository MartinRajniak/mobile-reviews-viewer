package com.example.reviews

import kotlinx.coroutines.*
import org.slf4j.Logger

class ReviewsRepository(
    private val logger: Logger,
    private val reviewsFetcher: ReviewsFetcher,
    private val reviewsStorage: ReviewsStorage,
    private val appIds: Set<String>
) {
    // SupervisorJob - so that failure in child coroutines doesn't cancel this scope
    suspend fun updateReviews() = supervisorScope {
        logger.info("Updating reviews...")

        appIds.map { appId ->
            launch {
                try {
                    logger.info("Fetch review for $appId")
                    val reviews = reviewsFetcher.fetchReviews(appId)
                    logger.info("Number of reviews fetched: ${reviews.size}")
                    reviewsStorage.saveReviews(reviews)
                } catch (e: CancellationException) {
                    logger.debug("Review fetch cancelled for $appId")
                    throw e
                } catch (e: Exception) {
                    logger.error("Failed to fetch review for $appId", e)
                }
            }
        }.joinAll()

        reviewsStorage.saveState()
    }

    fun getAllReviews(): List<Review> = reviewsStorage.getAllReviews()
}
