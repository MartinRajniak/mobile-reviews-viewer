package com.example

import kotlinx.serialization.json.Json
import kotlin.test.*

class ConfigTest {
    @Test
    fun testConfigDeserialization() {
        val jsonString = """
            {
                "apps": ["595068606", "447188370", "310633997"]
            }
        """.trimIndent()

        val json = Json { ignoreUnknownKeys = true }
        val config = json.decodeFromString<Config>(jsonString)

        assertEquals(3, config.apps.size)
        assertTrue(config.apps.contains("595068606"))
        assertTrue(config.apps.contains("447188370"))
        assertTrue(config.apps.contains("310633997"))
    }

    @Test
    fun testConfigWithEmptyApps() {
        val jsonString = """
            {
                "apps": []
            }
        """.trimIndent()

        val json = Json { ignoreUnknownKeys = true }
        val config = json.decodeFromString<Config>(jsonString)

        assertEquals(0, config.apps.size)
    }

    @Test
    fun testConfigIgnoresUnknownKeys() {
        val jsonString = """
            {
                "apps": ["595068606"],
                "unknownField": "value"
            }
        """.trimIndent()

        val json = Json { ignoreUnknownKeys = true }
        val config = json.decodeFromString<Config>(jsonString)

        assertEquals(1, config.apps.size)
        assertTrue(config.apps.contains("595068606"))
    }

    @Test
    fun testConfigDeduplicatesApps() {
        val jsonString = """
            {
                "apps": ["595068606", "595068606", "447188370"]
            }
        """.trimIndent()

        val json = Json { ignoreUnknownKeys = true }
        val config = json.decodeFromString<Config>(jsonString)

        // Set should deduplicate
        assertEquals(2, config.apps.size)
        assertTrue(config.apps.contains("595068606"))
        assertTrue(config.apps.contains("447188370"))
    }

    @Test
    fun testLoadConfigFromResources() {
        val config = loadConfig()

        assertNotNull(config)
        assertTrue(config.apps.isNotEmpty())
        assertEquals(3, config.apps.size)
    }

    @Test
    fun testConfigDataClass() {
        val config = Config(apps = setOf("595068606"))

        // Verify that apps is properly initialized
        assertEquals(1, config.apps.size)
        assertTrue(config.apps.contains("595068606"))

        // Test data class copy functionality
        val config2 = config.copy(apps = setOf("447188370"))
        assertEquals(1, config2.apps.size)
        assertTrue(config2.apps.contains("447188370"))
        assertFalse(config2.apps.contains("595068606"))
    }
}
