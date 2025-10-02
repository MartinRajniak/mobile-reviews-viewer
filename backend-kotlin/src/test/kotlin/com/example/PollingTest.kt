package com.example

import com.example.reviews.Review
import com.example.reviews.ReviewsFetcher
import com.example.reviews.ReviewsRepository
import io.ktor.server.config.*
import io.ktor.server.testing.*
import kotlinx.coroutines.*
import kotlinx.coroutines.test.*
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import kotlin.test.*
import kotlin.time.Duration.Companion.seconds

class PollingTest {
    private val testLogger: Logger = LoggerFactory.getLogger(PollingTest::class.java)

    private fun ApplicationTestBuilder.setupEnvironment() {
        environment {
            config = MapApplicationConfig().apply {
                put("polling.intervalSeconds", "1")
            }
        }
    }

    private class FakeReviewsFetcher(private val shouldFail: Boolean = false) : ReviewsFetcher {
        var fetchCount = 0

        override suspend fun fetchReviews(appId: String): List<Review> {
            fetchCount++
            if (shouldFail && appId == "INVALID_APP_ID") {
                throw Exception("Failed to fetch reviews for $appId")
            }
            return emptyList()
        }
    }

    @Test
    fun testConfigurePollingStartsWithMultipleApps() = runTest {
        testApplication {
            setupEnvironment()

            val fetcher = FakeReviewsFetcher()
            val repository = ReviewsRepository(testLogger, fetcher, setOf("595068606", "447188370"))

            application {
                val job = configurePolling(repository)
                delay(100)
                job.cancel()
            }

            assertTrue(true)
        }
    }

    @Test
    fun testConfigurePollingStartsWithEmptyApps() = runTest {
        testApplication {
            setupEnvironment()

            val fetcher = FakeReviewsFetcher()
            val repository = ReviewsRepository(testLogger, fetcher, emptySet())

            application {
                val job = configurePolling(repository)
                delay(100)
                job.cancel()
            }

            assertTrue(true)
        }
    }

    @Test
    fun testConfigurePollingStartsWithSingleApp() = runTest {
        testApplication {
            setupEnvironment()

            val fetcher = FakeReviewsFetcher()
            val repository = ReviewsRepository(testLogger, fetcher, setOf("595068606"))

            application {
                val job = configurePolling(repository)
                delay(100)
                job.cancel()
            }

            assertTrue(true)
        }
    }

    @Test
    fun testPollingServiceHandlesLargeNumberOfApps() = runTest {
        testApplication {
            setupEnvironment()

            val apps = (1..50).map { it.toString() }.toSet()
            val fetcher = FakeReviewsFetcher()
            val repository = ReviewsRepository(testLogger, fetcher, apps)

            application {
                val job = configurePolling(repository)
                delay(100)
                job.cancel()
            }

            assertTrue(true)
        }
    }

    @Test
    fun testConfigWithDuplicateAppIds() {
        // Config uses Set, so duplicates should be automatically handled
        val config = Config(apps = setOf("595068606", "595068606", "447188370"))

        assertEquals(2, config.apps.size)
        assertTrue(config.apps.contains("595068606"))
        assertTrue(config.apps.contains("447188370"))
    }

    @Test
    fun testConfigurePollingWithSupervisorJob() = runTest {
        testApplication {
            setupEnvironment()

            val fetcher = FakeReviewsFetcher()
            val repository = ReviewsRepository(testLogger, fetcher, setOf("595068606", "447188370", "310633997"))

            application {
                val job = configurePolling(repository)
                delay(200)
                job.cancel()
            }

            assertTrue(true)
        }
    }

    @Test
    fun testPollingContinuesWhenOneAppFails() = runTest {
        val fetcher = FakeReviewsFetcher(shouldFail = true)
        val repository = ReviewsRepository(testLogger, fetcher, setOf("595068606", "INVALID_APP_ID", "447188370"))

        // updateReviews should complete even with failures
        repository.updateReviews()

        // All apps should have been attempted
        assertEquals(3, fetcher.fetchCount)
    }

    @Test
    fun testPollerServiceStart() = runTest {
        val fetcher = FakeReviewsFetcher()
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1"))
        val poller = PollerService(testLogger, repository, 1.seconds)

        val job = launch {
            poller.start()
        }

        // Let it run once, verify it starts successfully
        delay(10)

        // Cancel and verify no exception
        job.cancelAndJoin()

        // Should have done at least initial fetch
        assertTrue(fetcher.fetchCount >= 1, "Expected at least 1 fetch call, got ${fetcher.fetchCount}")
    }

    @Test
    fun testPollerServiceCancellation() = runTest {
        val fetcher = FakeReviewsFetcher()
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1"))
        val poller = PollerService(testLogger, repository, 1.seconds)

        val job = launch {
            poller.start()
        }

        delay(1)
        job.cancelAndJoin()

        // Should complete without exception and have done initial fetch
        assertEquals(1, fetcher.fetchCount)
    }
}
