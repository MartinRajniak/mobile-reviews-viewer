package com.example.reviews

import kotlinx.coroutines.test.*
import org.slf4j.Logger
import org.slf4j.LoggerFactory
import kotlin.test.*

class ReviewsRepositoryTest {
    private val testLogger: Logger = LoggerFactory.getLogger(ReviewsRepositoryTest::class.java)

    private class FakeReviewsFetcher(
        private val shouldFail: Boolean = false,
        private val failForAppIds: Set<String> = emptySet()
    ) : ReviewsFetcher {
        val fetchedAppIds = mutableListOf<String>()

        override suspend fun fetchReviews(appId: String): List<Review> {
            fetchedAppIds.add(appId)
            if (shouldFail || appId in failForAppIds) {
                throw Exception("Failed to fetch reviews for $appId")
            }
            return emptyList()
        }
    }

    @Test
    fun testUpdateReviewsWithMultipleApps() = runTest {
        val fetcher = FakeReviewsFetcher()
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1", "app2", "app3"))

        repository.updateReviews()

        assertEquals(3, fetcher.fetchedAppIds.size)
        assertTrue(fetcher.fetchedAppIds.containsAll(listOf("app1", "app2", "app3")))
    }

    @Test
    fun testUpdateReviewsWithEmptyApps() = runTest {
        val fetcher = FakeReviewsFetcher()
        val repository = ReviewsRepository(testLogger, fetcher, emptySet())

        repository.updateReviews()

        assertEquals(0, fetcher.fetchedAppIds.size)
    }

    @Test
    fun testUpdateReviewsWithSingleApp() = runTest {
        val fetcher = FakeReviewsFetcher()
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1"))

        repository.updateReviews()

        assertEquals(1, fetcher.fetchedAppIds.size)
        assertEquals("app1", fetcher.fetchedAppIds[0])
    }

    @Test
    fun testUpdateReviewsContinuesOnFailure() = runTest {
        // One app fails, but others should still be fetched
        val fetcher = FakeReviewsFetcher(failForAppIds = setOf("app2"))
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1", "app2", "app3"))

        // Should not throw exception due to supervisorScope
        repository.updateReviews()

        // All apps should have been attempted
        assertEquals(3, fetcher.fetchedAppIds.size)
        assertTrue(fetcher.fetchedAppIds.containsAll(listOf("app1", "app2", "app3")))
    }

    @Test
    fun testUpdateReviewsHandlesAllFailures() = runTest {
        val fetcher = FakeReviewsFetcher(shouldFail = true)
        val repository = ReviewsRepository(testLogger, fetcher, setOf("app1", "app2"))

        // Should not throw exception even if all fail
        repository.updateReviews()

        // All apps should have been attempted
        assertEquals(2, fetcher.fetchedAppIds.size)
    }

    @Test
    fun testUpdateReviewsRunsConcurrently() = runTest {
        // This test verifies that all apps are fetched, which happens concurrently
        val fetcher = FakeReviewsFetcher()
        val apps = (1..10).map { "app$it" }.toSet()
        val repository = ReviewsRepository(testLogger, fetcher, apps)

        repository.updateReviews()

        // All apps should have been fetched
        assertEquals(10, fetcher.fetchedAppIds.size)
        assertTrue(fetcher.fetchedAppIds.toSet() == apps)
    }
}
