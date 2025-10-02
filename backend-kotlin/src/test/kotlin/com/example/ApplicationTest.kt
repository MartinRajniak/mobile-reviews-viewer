package com.example

import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.testing.*
import org.slf4j.LoggerFactory
import kotlin.test.*

class ApplicationTest {
    private val testLogger = LoggerFactory.getLogger(ApplicationTest::class.java)

    @Test
    fun testApiReviews() = testApplication {
        val config = Config(apps = setOf("test_app"))
        application {
            configureRouting(testLogger, testJson, FakeReviewsStorage(), config)
        }
        client.get("/api/reviews?app_id=test_app").apply {
            assertEquals(HttpStatusCode.OK, status)
            assertEquals("[]", bodyAsText())
        }
    }
}
