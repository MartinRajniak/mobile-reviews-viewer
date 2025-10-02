package com.example

import kotlin.test.*

class ConfigTest {
    @Test
    fun testConfigDeserialization() {
        val jsonString = """
            {
                "apps": ["389801252", "447188370", "310633997"]
            }
        """.trimIndent()

        val config = testJson.decodeFromString<Config>(jsonString)

        assertEquals(3, config.apps.size)
        assertTrue(config.apps.contains("389801252"))
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

        val config = testJson.decodeFromString<Config>(jsonString)

        assertEquals(0, config.apps.size)
    }

    @Test
    fun testConfigIgnoresUnknownKeys() {
        val jsonString = """
            {
                "apps": ["389801252"],
                "unknownField": "value"
            }
        """.trimIndent()

        val config = testJson.decodeFromString<Config>(jsonString)

        assertEquals(1, config.apps.size)
        assertTrue(config.apps.contains("389801252"))
    }

    @Test
    fun testConfigDeduplicatesApps() {
        val jsonString = """
            {
                "apps": ["389801252", "389801252", "447188370"]
            }
        """.trimIndent()

        val config = testJson.decodeFromString<Config>(jsonString)

        // Set should deduplicate
        assertEquals(2, config.apps.size)
        assertTrue(config.apps.contains("389801252"))
        assertTrue(config.apps.contains("447188370"))
    }

    @Test
    fun testLoadConfigFromResources() {
        val config = loadConfig(testJson)

        assertNotNull(config)
        assertTrue(config.apps.isNotEmpty())
        assertEquals(3, config.apps.size)
    }

    @Test
    fun testConfigDataClass() {
        val config = Config(apps = setOf("389801252"))

        // Verify that apps is properly initialized
        assertEquals(1, config.apps.size)
        assertTrue(config.apps.contains("389801252"))

        // Test data class copy functionality
        val config2 = config.copy(apps = setOf("447188370"))
        assertEquals(1, config2.apps.size)
        assertTrue(config2.apps.contains("447188370"))
        assertFalse(config2.apps.contains("389801252"))
    }
}
