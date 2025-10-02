package com.example

import com.example.reviews.Review
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.testing.*
import org.slf4j.LoggerFactory
import kotlin.test.*
import kotlin.time.Instant

class RoutingTest {
    private val testLogger = LoggerFactory.getLogger(RoutingTest::class.java)

    @Test
    fun testApiReviewsEndpoint() = testApplication {
        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage())
        }

        val response = client.get("/api/reviews?app_id=test_app")

        assertEquals(HttpStatusCode.OK, response.status)
        assertEquals("[]", response.bodyAsText())
        assertEquals(ContentType.Application.Json, response.contentType())
    }

    @Test
    fun testApiHealthEndpoint() = testApplication {
        val now = Instant.parse("2023-10-01T12:00:00Z")
        val mockReviews = listOf(
            Review("1", "app1", "Author", "Content", 5, now, now),
            Review("2", "app2", "Author", "Content", 4, now, now)
        )

        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage(mockReviews))
        }

        val response = client.get("/api/health")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.bodyAsText()
        assertTrue(body.contains("\"status\":\"healthy\""))
        assertTrue(body.contains("\"total_reviews\":2"))
    }

    @Test
    fun testApiHealthEndpointEmpty() = testApplication {
        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage())
        }

        val response = client.get("/api/health")

        assertEquals(HttpStatusCode.OK, response.status)
        val body = response.bodyAsText()
        assertTrue(body.contains("\"status\":\"healthy\""))
        assertTrue(body.contains("\"total_reviews\":0"))
    }

    @Test
    fun testNonExistentEndpoint() = testApplication {
        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage())
        }

        val response = client.get("/nonexistent")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }

    @Test
    fun testApiReviewsWithDifferentMethods() = testApplication {
        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage())
        }

        // GET should work
        client.get("/api/reviews?app_id=test_app").apply {
            assertEquals(HttpStatusCode.OK, status)
        }

        // POST should return Method Not Allowed or Not Found
        client.post("/api/reviews?app_id=test_app").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }

        // PUT should return Method Not Allowed or Not Found
        client.put("/api/reviews?app_id=test_app").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }

        // DELETE should return Method Not Allowed or Not Found
        client.delete("/api/reviews?app_id=test_app").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }
    }
}
