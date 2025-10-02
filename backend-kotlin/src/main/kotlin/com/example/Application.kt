package com.example

import com.example.reviews.ReviewsRepository
import com.example.reviews.itunes.ITunesReviewsFetcher
import io.ktor.client.*
import io.ktor.client.engine.cio.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.serialization.kotlinx.json.*
import io.ktor.server.application.*
import io.ktor.server.netty.*
import kotlinx.coroutines.cancelAndJoin
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

fun main(args: Array<String>) = EngineMain.main(args)

@Suppress("unused")
fun Application.module() {
    val config = loadConfig()
    val httpClient = HttpClient(CIO) {
        install(ContentNegotiation) {
            json(Json { ignoreUnknownKeys = true })
        }
    }
    val reviewsFetcher = ITunesReviewsFetcher(httpClient)
    val reviewsRepository = ReviewsRepository(
        log,
        reviewsFetcher,
        config.apps
    )

    val pollingJob = configurePolling(reviewsRepository)
    configureRouting()

    monitor.subscribe(ApplicationStopping) {
        runBlocking {
            // Cancel if not already cancelled
            pollingJob.cancelAndJoin()
            // Close HttpClient only after polling is done (coroutines finished)
            httpClient.close()
        }
    }
}

fun loadConfig(): Config {
    val classLoader = Config::class.java.classLoader

    val resource = classLoader.getResource("config.json")
        ?: throw IllegalStateException("Resource file not found: config.json")
    val jsonString = resource.readText()

    val json = Json { ignoreUnknownKeys = true }
    return json.decodeFromString<Config>(jsonString)
}

@Serializable
data class Config(
    val apps: Set<String>
)
