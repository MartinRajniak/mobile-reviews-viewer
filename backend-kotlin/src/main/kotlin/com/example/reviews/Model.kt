package com.example.reviews

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlin.time.Instant

@Serializable
data class Review(
    val id: String,
    @SerialName("app_id")
    val appId: String,
    val author: String,
    val content: String,
    val rating: Int,
    @SerialName("submitted_at")
    val submittedAt: Instant,
    @SerialName("fetched_at")
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
    fun getRecentReviews(appId: String?, since: Instant): List<Review>
}
