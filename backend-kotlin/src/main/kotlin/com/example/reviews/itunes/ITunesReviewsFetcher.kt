package com.example.reviews.itunes

import com.example.reviews.Review
import com.example.reviews.ReviewsFetcher
import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.request.get
import io.ktor.http.*
import io.ktor.http.headers
import kotlinx.serialization.json.Json
import kotlin.time.Clock
import kotlin.time.Instant

class ITunesReviewsFetcher(
    private val httpClient: HttpClient
) : ReviewsFetcher {

    private val json = Json { ignoreUnknownKeys = true }

    override suspend fun fetchReviews(appId: String): List<Review> {
        val feedUrl = "https://itunes.apple.com/us/rss/customerreviews/id=$appId/sortBy=mostRecent/page=1/json"
        val response = httpClient.get(feedUrl) {
            headers {
                append("User-Agent", "AppReviewPoller/1.0")
            }
        }

        if (response.status != HttpStatusCode.OK) throw Exception("Unexpected status code: ${response.status}")

        val rssFeed = json.decodeFromString<RssFeed>(response.body())

        val now = Clock.System.now()
        val reviews = rssFeed.feed.entry.map {
            Review(
                it.id.label,
                appId,
                it.author.name.label,
                it.content.label,
                it.rating.label.toInt(),
                Instant.parse(it.updated.label),
                now
            )
        }

        return reviews
    }
}
