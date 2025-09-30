package com.example

import io.ktor.server.application.*
import kotlinx.coroutines.*
import kotlinx.coroutines.time.delay
import java.time.Duration

fun Application.configurePolling(config: Config) {
    val pollInterval = environment.config.property("polling.intervalSeconds").getString().toLong()

    // If one application fetch fails, continue with others
    launch(Dispatchers.Default + SupervisorJob()) {
        log.info("Starting poller service...")
        try {
            while (isActive) {
                try {
                    log.info("Fetching reviews...")
                    loadReviewsConcurrently(config)
                    delay(pollInterval)
                } catch (e: CancellationException) {
                    log.info("Polling service was cancelled.", e)
                    throw e
                } catch (e: Exception) {
                    log.error("Error in polling service", e)
                }
            }
        } finally {
            log.info("Polling service stopped.")
        }
    }
}

private suspend fun Application.loadReviewsConcurrently(config: Config) = coroutineScope {
    val jobs = mutableSetOf<Job>()
    for (appId in config.apps) {
        val job = launch(Dispatchers.IO) {
            log.info("Fetch review for $appId")
            // TODO: fetch and store reviews
        }
        jobs.add(job)
    }
    jobs.joinAll()
}
