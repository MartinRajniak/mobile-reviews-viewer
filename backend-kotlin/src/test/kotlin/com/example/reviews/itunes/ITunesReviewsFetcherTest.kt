package com.example.reviews.itunes

import com.example.testJson
import io.ktor.client.*
import io.ktor.client.engine.mock.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import io.ktor.utils.io.*
import kotlinx.coroutines.test.*
import kotlin.test.*

class ITunesReviewsFetcherTest {

    private fun createMockHttpClient(
        responseContent: String,
        statusCode: HttpStatusCode = HttpStatusCode.OK
    ): HttpClient {
        return HttpClient(MockEngine) {
            install(ContentNegotiation) {
                json(testJson)
            }
            engine {
                addHandler { request ->
                    respond(
                        content = ByteReadChannel(responseContent),
                        status = statusCode,
                        headers = headersOf(HttpHeaders.ContentType, "application/json")
                    )
                }
            }
        }
    }

    @Test
    fun testFetchReviewsSuccess() = runTest {
        val jsonResponse = """
            {
                "feed": {
                    "entry": [
                        {
                            "id": {"label": "review1"},
                            "author": {"name": {"label": "John Doe"}},
                            "content": {"label": "Great app!"},
                            "updated": {"label": "2023-10-01T12:00:00Z"},
                            "im:rating": {"label": "5"}
                        },
                        {
                            "id": {"label": "review2"},
                            "author": {"name": {"label": "Jane Smith"}},
                            "content": {"label": "Nice features"},
                            "updated": {"label": "2023-10-02T14:30:00Z"},
                            "im:rating": {"label": "4"}
                        }
                    ]
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(jsonResponse)
        val fetcher = ITunesReviewsFetcher(httpClient)

        val reviews = fetcher.fetchReviews("595068606")

        assertEquals(2, reviews.size)
        assertEquals("review1", reviews[0].id)
        assertEquals("595068606", reviews[0].appId)
        assertEquals("John Doe", reviews[0].author)
        assertEquals("Great app!", reviews[0].content)
        assertEquals(5, reviews[0].rating)

        assertEquals("review2", reviews[1].id)
        assertEquals("Jane Smith", reviews[1].author)
        assertEquals("Nice features", reviews[1].content)
        assertEquals(4, reviews[1].rating)
    }

    @Test
    fun testFetchReviewsEmptyList() = runTest {
        val jsonResponse = """
            {
                "feed": {
                    "entry": []
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(jsonResponse)
        val fetcher = ITunesReviewsFetcher(httpClient)

        val reviews = fetcher.fetchReviews("595068606")

        assertEquals(0, reviews.size)
    }

    @Test
    fun testFetchReviewsHttpError() = runTest {
        val httpClient = createMockHttpClient("", HttpStatusCode.InternalServerError)
        val fetcher = ITunesReviewsFetcher(httpClient)

        assertFailsWith<Exception> {
            fetcher.fetchReviews("595068606")
        }
    }

    @Test
    fun testFetchReviewsInvalidJson() = runTest {
        val httpClient = createMockHttpClient("invalid json")
        val fetcher = ITunesReviewsFetcher(httpClient)

        assertFailsWith<Exception> {
            fetcher.fetchReviews("595068606")
        }
    }

    @Test
    fun testFetchReviewsSetsCorrectAppId() = runTest {
        val jsonResponse = """
            {
                "feed": {
                    "entry": [
                        {
                            "id": {"label": "review1"},
                            "author": {"name": {"label": "Author"}},
                            "content": {"label": "Content"},
                            "updated": {"label": "2023-10-01T12:00:00Z"},
                            "im:rating": {"label": "3"}
                        }
                    ]
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(jsonResponse)
        val fetcher = ITunesReviewsFetcher(httpClient)

        val reviews = fetcher.fetchReviews("123456789")

        assertEquals(1, reviews.size)
        assertEquals("123456789", reviews[0].appId)
    }

    @Test
    fun testFetchReviewsSetsFetchedAt() = runTest {
        val jsonResponse = """
            {
                "feed": {
                    "entry": [
                        {
                            "id": {"label": "review1"},
                            "author": {"name": {"label": "Author"}},
                            "content": {"label": "Content"},
                            "updated": {"label": "2023-10-01T12:00:00Z"},
                            "im:rating": {"label": "5"}
                        }
                    ]
                }
            }
        """.trimIndent()

        val httpClient = createMockHttpClient(jsonResponse)
        val fetcher = ITunesReviewsFetcher(httpClient)

        val reviews = fetcher.fetchReviews("595068606")

        assertEquals(1, reviews.size)
        assertNotNull(reviews[0].fetchedAt)
        // fetchedAt should be close to now, but we can't test exact value
        // Just verify it exists
    }
}
