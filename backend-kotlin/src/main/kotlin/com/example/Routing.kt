package com.example

import com.example.reviews.Review
import com.example.reviews.ReviewsStorage
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.cors.routing.CORS
import io.ktor.server.response.*
import io.ktor.server.routing.*
import io.ktor.utils.io.*
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import org.slf4j.Logger
import kotlin.time.Clock
import kotlin.time.Duration.Companion.hours
import kotlin.time.Instant

@Serializable
data class HealthResponse(
    val status: String,
    val timestamp: Instant,
    @SerialName("total_reviews")
    val totalReviews: Int
)

@Serializable
data class AverageRatingResponse(
    @SerialName("app_id")
    val appId: String,
    @SerialName("average_rating")
    val averageRating: Double,
    @SerialName("review_count")
    val reviewCount: Int,
    val hours: Int
)

// TODO: Use ReviewsRepository instead
fun Application.configureRouting(logger: Logger, json: Json, reviewsStorage: ReviewsStorage) {
    install(CORS) {
        anyHost()
        allowMethod(HttpMethod.Get)
        allowMethod(HttpMethod.Post)
        allowMethod(HttpMethod.Options)
        allowHeader(HttpHeaders.ContentType)
    }
    routing {
        get("/api/reviews") {
            getRecentReviews(logger, reviewsStorage, json)
        }
        get("/api/health") {
            healthCheck(logger, reviewsStorage, json)
        }
        get("/api/average-rating") {
            getAverageRating(logger, reviewsStorage, json)
        }
    }
}

private suspend fun RoutingContext.getAverageRating(
    logger: Logger,
    reviewsStorage: ReviewsStorage,
    json: Json
) {
    try {
        val params = call.queryParameters
        if (!params.contains("app_id")) {
            logger.error("Missing app_id in request")
            call.respond(HttpStatusCode.BadRequest, "app_id query parameter is required")
            return
        }
        val appId = params["app_id"]!!

        val hours = try {
            params["hours"]?.toUInt()?.toInt() ?: 48
        } catch (e: NumberFormatException) {
            logger.error("Parameter hours is not a positive integer", e)
            call.respond(HttpStatusCode.BadRequest, "hours must be a positive integer")
            return
        }

        val since = Clock.System.now() - hours.hours
        val reviews = reviewsStorage.getRecentReviews(appId, since)

        val averageRating = if (reviews.isEmpty()) {
            0.0
        } else {
            reviews.map { it.rating }.average()
        }

        val response = AverageRatingResponse(
            appId, averageRating, reviews.size, hours
        )

        call.respondText(
            json.encodeToString<AverageRatingResponse>(response),
            ContentType.Application.Json
        )
    } catch (e: CancellationException) {
        logger.error("Application was terminated mid-request", e)
        call.respond(HttpStatusCode.InternalServerError, "Application was terminated mid-request")
        throw e
    } catch (e: Exception) {
        logger.error("Request could not be finished.", e)
        call.respond(HttpStatusCode.InternalServerError, e.toString())
    }
}

private suspend fun RoutingContext.getRecentReviews(
    logger: Logger,
    reviewsStorage: ReviewsStorage,
    json: Json
) {
    try {
        val params = call.queryParameters
        if (!params.contains("app_id")) {
            logger.error("Missing app_id in request")
            call.respond(HttpStatusCode.BadRequest, "app_id query parameter is required")
            return
        }
        val appId = params["app_id"]

        val hours = try {
            params["hours"]?.toUInt()?.toInt() ?: 48
        } catch (e: NumberFormatException) {
            logger.error("Parameter hours is not a positive integer", e)
            call.respond(HttpStatusCode.BadRequest, "hours must be a positive integer")
            return
        }

        val since = Clock.System.now() - hours.hours
        val reviews = reviewsStorage.getRecentReviews(appId, since)

        call.respondText(
            json.encodeToString<List<Review>>(reviews),
            ContentType.Application.Json
        )
    } catch (e: CancellationException) {
        logger.error("Application was terminated mid-request", e)
        call.respond(HttpStatusCode.InternalServerError, "Application was terminated mid-request")
        throw e
    } catch (e: Exception) {
        logger.error("Request could not be finished.", e)
        call.respond(HttpStatusCode.InternalServerError, e.toString())
    }
}

private suspend fun RoutingContext.healthCheck(
    logger: Logger,
    reviewsStorage: ReviewsStorage,
    json: Json
) {
    try {
        val allReviews = reviewsStorage.getAllReviews()
        val response = HealthResponse(
            status = "healthy",
            timestamp = Clock.System.now(),
            totalReviews = allReviews.size
        )
        call.respondText(
            json.encodeToString<HealthResponse>(response),
            ContentType.Application.Json
        )
    } catch (e: CancellationException) {
        logger.error("Application was terminated mid-request", e)
        call.respond(HttpStatusCode.InternalServerError, "Application was terminated mid-request")
        throw e
    } catch (e: Exception) {
        logger.error("Request could not be finished.", e)
        call.respond(HttpStatusCode.InternalServerError, e.toString())
    }
}
