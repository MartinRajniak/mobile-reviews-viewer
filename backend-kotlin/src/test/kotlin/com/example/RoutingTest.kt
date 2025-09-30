package com.example

import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.testing.*
import kotlin.test.*

class RoutingTest {
    @Test
    fun testRootEndpoint() = testApplication {
        application {
            configureRouting()
        }

        val response = client.get("/")

        assertEquals(HttpStatusCode.OK, response.status)
        assertEquals("Hello World!", response.bodyAsText())
        assertEquals(ContentType.Text.Plain.withCharset(Charsets.UTF_8), response.contentType())
    }

    @Test
    fun testRootEndpointMultipleRequests() = testApplication {
        application {
            configureRouting()
        }

        // Make multiple requests to ensure consistency
        repeat(5) {
            val response = client.get("/")
            assertEquals(HttpStatusCode.OK, response.status)
            assertEquals("Hello World!", response.bodyAsText())
        }
    }

    @Test
    fun testNonExistentEndpoint() = testApplication {
        application {
            configureRouting()
        }

        val response = client.get("/nonexistent")

        assertEquals(HttpStatusCode.NotFound, response.status)
    }

    @Test
    fun testRootEndpointWithDifferentMethods() = testApplication {
        application {
            configureRouting()
        }

        // GET should work
        client.get("/").apply {
            assertEquals(HttpStatusCode.OK, status)
        }

        // POST should return Method Not Allowed or Not Found
        client.post("/").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }

        // PUT should return Method Not Allowed or Not Found
        client.put("/").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }

        // DELETE should return Method Not Allowed or Not Found
        client.delete("/").apply {
            assertTrue(status == HttpStatusCode.MethodNotAllowed || status == HttpStatusCode.NotFound)
        }
    }

    @Test
    fun testRootEndpointResponseHeaders() = testApplication {
        application {
            configureRouting()
        }

        val response = client.get("/")

        assertNotNull(response.headers[HttpHeaders.ContentType])
        assertTrue(response.headers[HttpHeaders.ContentType]!!.contains("text/plain"))
    }
}
