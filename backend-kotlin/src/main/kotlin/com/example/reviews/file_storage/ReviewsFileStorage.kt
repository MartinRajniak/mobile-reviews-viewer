package com.example.reviews.file_storage

import com.example.reviews.Review
import com.example.reviews.ReviewsStorage
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import org.slf4j.Logger
import java.nio.charset.StandardCharsets
import java.nio.file.Files
import java.nio.file.Paths
import java.nio.file.StandardCopyOption
import java.nio.file.StandardOpenOption
import java.util.concurrent.ConcurrentHashMap
import kotlin.io.path.exists

class ReviewsFileStorage(
    private val logger: Logger,
    private val json: Json,
    private val storageFilePath: String
) : ReviewsStorage {
    // Can be accessed from multiple threads
    private val allReviews = ConcurrentHashMap<String, Review>()

    init {
        val path = Paths.get(storageFilePath)
        val parentPath = path.parent
        if (parentPath == null) {
            logger.info("Path is relative or just a file name. No directory creation needed.")
        } else {
            Files.createDirectories(parentPath)
            logger.info("Internal storage folder created successfully")
        }
    }

    override suspend fun loadState() {
        val path = Paths.get(storageFilePath)
        if (!path.exists()) {
            logger.info("Cannot load reviews since file does not exist yet.")
            return
        }
        try {
            val jsonString = withContext(Dispatchers.IO) {
                 Files.readString(path, StandardCharsets.UTF_8)
            }
            val reviews = json.decodeFromString<List<Review>>(jsonString)
            allReviews.clear()
            saveReviews(reviews)
            logger.info("Reviews successfully loaded from file storage.")
        } catch (e: Exception) {
            logger.error("Error while loading reviews from file storage.", e)
        }
    }

    override suspend fun saveState() {
        val jsonString = json.encodeToString<List<Review>>(getAllReviews())

        withContext(Dispatchers.IO) {
            // Atomic write: write to temp file, then rename
            // This prevents corruption if the process crashes mid-write

            val finalPath = Paths.get(storageFilePath)
            val tempPath = Paths.get("$storageFilePath.tmp")

            Files.writeString(
                tempPath, jsonString,
                StandardCharsets.UTF_8,
                StandardOpenOption.CREATE,
                StandardOpenOption.WRITE,
                StandardOpenOption.TRUNCATE_EXISTING
            )

            Files.move(
                tempPath,
                finalPath,
                StandardCopyOption.REPLACE_EXISTING,
                StandardCopyOption.ATOMIC_MOVE // may fail cross-filesystem
            )
        }

        logger.info("Reviews successfully persisted.")
    }

    override fun saveReviews(reviews: List<Review>) {
        reviews.forEach { allReviews[it.id] = it }
    }

    override fun getAllReviews(): List<Review> {
        return allReviews.values.toList()
    }
}