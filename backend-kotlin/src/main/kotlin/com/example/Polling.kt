package com.example

import com.example.reviews.ReviewsRepository
import io.ktor.server.application.*
import io.ktor.util.logging.*
import kotlinx.coroutines.*
import kotlin.time.Duration
import kotlin.time.Duration.Companion.seconds

fun Application.configurePolling(reviewsRepository: ReviewsRepository): Job {
    val pollIntervalInSeconds = environment.config.property("polling.intervalSeconds").getString().toLong()

    val poller = PollerService(
        log,
        reviewsRepository,
        pollIntervalInSeconds.seconds
    )

    // use CoroutineScope from Application so that this coroutine is cancelled when application is terminated
    return launch {
        poller.start()
    }
}

class PollerService(
    val log: Logger,
    val reviewsRepository: ReviewsRepository,
    val pollInterval: Duration,
) {
    suspend fun start() {
        log.info("Starting poller service...")
        try {
            while (currentCoroutineContext().isActive) {
                reviewsRepository.updateReviews()
                delay(pollInterval)
            }
        } catch (e: CancellationException) {
            log.info("Polling service was cancelled.")
            throw e
        } catch (e: Exception) {
            log.error("Error in polling service", e)
        } finally {
            log.info("Polling service stopped.")
        }
    }
}
