package com.example.reviews

import kotlinx.serialization.Serializable
import kotlin.time.Instant

@Serializable
data class Review(
    val id: String,
    val appId: String,
    val author: String,
    val content: String,
    val rating: Int,
    val submittedAt: Instant,
    val fetchedAt: Instant
)

interface ReviewsFetcher {
    suspend fun fetchReviews(appId: String): List<Review>
}

interface ReviewsStorage {
    fun saveReviews(reviews: List<Review>)
    fun getAllReviews(): List<Review>
    suspend fun loadState()
    suspend fun saveState()
}