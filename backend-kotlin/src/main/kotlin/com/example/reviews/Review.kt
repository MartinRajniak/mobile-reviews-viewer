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
