package com.example

import com.example.reviews.ReviewsRepository
import com.example.reviews.file_storage.ReviewsFileStorage
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
    val json = Json { ignoreUnknownKeys = true }
    val config = loadConfig(json)
    val httpClient = HttpClient(CIO) {
        install(ContentNegotiation) {
            json(json)
        }
    }
    val reviewsFetcher = ITunesReviewsFetcher(httpClient)
    val storageFilePath = "data/reviews.json"
    val reviewsStorage = ReviewsFileStorage(log, json, storageFilePath)
    runBlocking {
        reviewsStorage.loadState()
    }
    val reviewsRepository = ReviewsRepository(
        log,
        reviewsFetcher,
        reviewsStorage,
        config.apps
    )
    val reviews = reviewsRepository.getAllReviews()
    log.info("Initial number of reviews is: ${reviews.size}")

    val pollingJob = configurePolling(reviewsRepository)
    configureRouting(
        log,
        json,
        reviewsStorage,
        config
    )

    monitor.subscribe(ApplicationStopping) {
        runBlocking {
            // Cancel if not already cancelled
            pollingJob.cancelAndJoin()
            // Close HttpClient only after polling is done (coroutines finished)
            httpClient.close()
            // Persist reviews
            reviewsStorage.saveState()
        }
    }
}

fun loadConfig(json: Json): Config {
    val classLoader = Config::class.java.classLoader

    val resource = classLoader.getResource("config.json")
        ?: throw IllegalStateException("Resource file not found: config.json")
    val jsonString = resource.readText()

    return json.decodeFromString<Config>(jsonString)
}

@Serializable
data class Config(
    val apps: Set<String>
)
