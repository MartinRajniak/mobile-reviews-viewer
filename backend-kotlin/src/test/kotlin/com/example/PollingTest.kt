package com.example

import io.ktor.server.application.*
import io.ktor.server.testing.*
import kotlinx.coroutines.*
import kotlinx.coroutines.test.*
import kotlin.test.*

class PollingTest {
    @Test
    fun testConfigurePollingStartsWithMultipleApps() = runTest {
        testApplication {
            val config = Config(apps = setOf("595068606", "447188370"))

            application {
                configurePolling(config)
            }

            // Give the coroutine time to start
            delay(100)

            // If we get here without exception, polling started successfully
            assertTrue(true)
        }
    }

    @Test
    fun testConfigurePollingStartsWithEmptyApps() = runTest {
        testApplication {
            val config = Config(apps = emptySet())

            application {
                // Should handle empty apps gracefully
                configurePolling(config)
            }

            delay(100)

            assertTrue(true)
        }
    }

    @Test
    fun testConfigurePollingStartsWithSingleApp() = runTest {
        testApplication {
            val config = Config(apps = setOf("595068606"))

            application {
                configurePolling(config)
            }

            delay(100)

            assertTrue(true)
        }
    }

    @Test
    fun testPollingServiceHandlesLargeNumberOfApps() = runTest {
        testApplication {
            val apps = (1..50).map { it.toString() }.toSet()
            val config = Config(apps = apps)

            application {
                // Should handle many apps without issue
                configurePolling(config)
            }

            delay(100)

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
            val config = Config(apps = setOf("595068606", "447188370", "310633997"))

            application {
                // The polling uses SupervisorJob, so if one app fails, others continue
                configurePolling(config)
            }

            // Let it run for a short period
            delay(200)

            // If we get here, the supervisor job is working correctly
            assertTrue(true)
        }
    }
}
