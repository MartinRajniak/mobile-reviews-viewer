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

// TODO: Use ReviewsRepository instead
fun Application.configureRouting(logger: Logger, json: Json, reviewsStorage: ReviewsStorage, config: Config) {
    install(CORS) {
        anyHost()
        allowMethod(HttpMethod.Get)
        allowMethod(HttpMethod.Post)
        allowMethod(HttpMethod.Options)
        allowHeader(HttpHeaders.ContentType)
    }
    routing {
        // Web app route
        webApp(logger, config, json)

        // API routes
        get("/api/reviews") {
            getRecentReviews(logger, reviewsStorage, json)
        }
        get("/api/health") {
            healthCheck(reviewsStorage, json, logger)
        }
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
    reviewsStorage: ReviewsStorage,
    json: Json,
    logger: Logger
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
