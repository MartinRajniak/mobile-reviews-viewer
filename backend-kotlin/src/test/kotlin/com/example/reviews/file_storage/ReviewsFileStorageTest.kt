package com.example.reviews.file_storage

import com.example.reviews.Review
import kotlinx.coroutines.test.runTest
import kotlinx.serialization.json.Json
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import java.nio.file.Files
import java.nio.file.Paths
import kotlin.test.*
import kotlin.time.Instant

class ReviewsFileStorageTest {
    private val testLogger: Logger = LoggerFactory.getLogger(ReviewsFileStorageTest::class.java)
    private val json = Json { ignoreUnknownKeys = true }
    private val testFilePath = "build/test-data/reviews-test.json"

    @BeforeTest
    fun setup() {
        // Clean up test file before each test
        val path = Paths.get(testFilePath)
        val tempPath = Paths.get("$testFilePath.tmp")
        Files.deleteIfExists(path)
        Files.deleteIfExists(tempPath)
    }

    @AfterTest
    fun cleanup() {
        // Clean up test file after each test
        val path = Paths.get(testFilePath)
        val tempPath = Paths.get("$testFilePath.tmp")
        Files.deleteIfExists(path)
        Files.deleteIfExists(tempPath)
    }

    @Test
    fun testSaveAndLoadReviews() = runTest {
        val storage = ReviewsFileStorage(testLogger, json, testFilePath)

        val now = Instant.parse("2023-10-01T12:00:00Z")
        val reviews = listOf(
            Review("1", "app1", "Author1", "Content1", 5, now, now),
            Review("2", "app2", "Author2", "Content2", 4, now, now)
        )

        storage.saveReviews(reviews)
        storage.saveState()

        // Create new storage instance to test loading
        val storage2 = ReviewsFileStorage(testLogger, json, testFilePath)
        storage2.loadState()

        val loadedReviews = storage2.getAllReviews()
        assertEquals(2, loadedReviews.size)
        assertEquals("1", loadedReviews.find { it.id == "1" }?.id)
        assertEquals("2", loadedReviews.find { it.id == "2" }?.id)
    }

    @Test
    fun testSaveReviewsDeduplicatesById() {
        val storage = ReviewsFileStorage(testLogger, json, testFilePath)

        val now = Instant.parse("2023-10-01T12:00:00Z")
        val review1 = Review("1", "app1", "Author1", "Old Content", 3, now, now)
        val review2 = Review("1", "app1", "Author1", "New Content", 5, now, now)

        storage.saveReviews(listOf(review1))
        storage.saveReviews(listOf(review2))

        val allReviews = storage.getAllReviews()
        assertEquals(1, allReviews.size)
        assertEquals("New Content", allReviews[0].content)
    }

    @Test
    fun testGetAllReviewsEmpty() {
        val storage = ReviewsFileStorage(testLogger, json, testFilePath)

        val reviews = storage.getAllReviews()
        assertEquals(0, reviews.size)
    }

    @Test
    fun testAtomicWrite() = runTest {
        val storage = ReviewsFileStorage(testLogger, json, testFilePath)

        val now = Instant.parse("2023-10-01T12:00:00Z")
        val reviews = listOf(
            Review("1", "app1", "Author1", "Content1", 5, now, now)
        )

        storage.saveReviews(reviews)
        storage.saveState()

        // Verify temp file was cleaned up
        val tempPath = Paths.get("$testFilePath.tmp")
        assertFalse(Files.exists(tempPath), "Temp file should be removed after atomic move")

        // Verify final file exists
        val finalPath = Paths.get(testFilePath)
        assertTrue(Files.exists(finalPath), "Final file should exist")
    }

    @Test
    fun testMultipleSaveReviewsCalls() {
        val storage = ReviewsFileStorage(testLogger, json, testFilePath)

        val now = Instant.parse("2023-10-01T12:00:00Z")

        storage.saveReviews(listOf(Review("1", "app1", "Author1", "Content1", 5, now, now)))
        storage.saveReviews(listOf(Review("2", "app2", "Author2", "Content2", 4, now, now)))
        storage.saveReviews(listOf(Review("3", "app3", "Author3", "Content3", 3, now, now)))

        val allReviews = storage.getAllReviews()
        assertEquals(3, allReviews.size)
    }

    @Test
    fun testDirectoryCreation() = runTest {
        val nestedPath = "build/test-data/nested/deep/reviews.json"

        try {
            val storage = ReviewsFileStorage(testLogger, json, nestedPath)

            val now = Instant.parse("2023-10-01T12:00:00Z")
            storage.saveReviews(listOf(Review("1", "app1", "Author", "Content", 5, now, now)))
            storage.saveState()

            assertTrue(Files.exists(Paths.get(nestedPath)))
        } finally {
            // Cleanup
            Files.deleteIfExists(Paths.get(nestedPath))
            Files.deleteIfExists(Paths.get("$nestedPath.tmp"))
            var parent = Paths.get(nestedPath).parent
            while (parent != null && parent.toString().startsWith("build/test-data/nested")) {
                Files.deleteIfExists(parent)
                parent = parent.parent
            }
        }
    }
}
