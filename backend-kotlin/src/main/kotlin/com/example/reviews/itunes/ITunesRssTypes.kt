package com.example.reviews.itunes

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable


@Serializable
data class RssFeed(
    val feed: Feed
)

@Serializable
data class Feed(
    val entry: List<Entry>
)

@Serializable
data class Entry(
    val id: Id,
    val author: Author,
    val content: Content,
    val updated: Updated,
    @SerialName("im:rating")
    val rating: Rating
)

@Serializable
data class Author(
    val name: Name
)

@Serializable
data class Name(
    val label: String
)

@Serializable
data class Content(
    val label: String
)

@Serializable
data class Updated(
    val label: String
)

@Serializable
data class Rating(
    val label: String
)

@Serializable
data class Id(
    val label: String
)
