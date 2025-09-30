package com.example

import io.ktor.server.application.*
import io.ktor.server.netty.*
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json

fun main(args: Array<String>) = EngineMain.main(args)

fun Application.module() {
    val config = loadConfig()
    configurePolling(config)
    configureRouting()
}

fun loadConfig(): Config {
    val classLoader = Config::class.java.classLoader

    val resource = classLoader.getResource("config.json")
        ?: throw IllegalStateException("Resource file not found: config.json")
    val jsonString = resource.readText()

    val json = Json { ignoreUnknownKeys = true }
    return json.decodeFromString<Config>(jsonString)
}

@Serializable
data class Config(
    val apps: Set<String>
)
