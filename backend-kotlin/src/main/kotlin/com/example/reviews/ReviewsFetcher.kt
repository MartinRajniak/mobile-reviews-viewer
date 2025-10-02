package com.example.reviews

interface ReviewsFetcher {
    suspend fun fetchReviews(appId: String): List<Review>
}