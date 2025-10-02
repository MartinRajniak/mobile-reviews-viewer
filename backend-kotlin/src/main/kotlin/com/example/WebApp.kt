package com.example

import com.example.reviews.Review
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.html.*
import io.ktor.server.routing.*
import kotlinx.html.*
import kotlinx.serialization.json.Json
import kotlinx.serialization.encodeToString
import org.slf4j.Logger
import kotlin.time.Clock
import kotlin.time.Duration.Companion.hours

fun Route.webApp(logger: Logger, config: Config, json: Json) {
    get("/") {
        val appIds = config.apps.toList()
        val apps = listOf(
            App("389801252", "Instagram"),
            App("447188370", "Twitter"),
            App("310633997", "WhatsApp Messenger")
        )

        call.respondHtml(HttpStatusCode.OK) {
            head {
                title { +"App Store Reviews Viewer" }
                meta(charset = "UTF-8")
                meta(name = "viewport", content = "width=device-width, initial-scale=1.0")
                link(rel = "stylesheet", href = "/static/css/app.css")
            }
            body {
                div(classes = "app") {
                    header(classes = "app-header") {
                        h1 { +"App Store Reviews Viewer" }
                        div(classes = "app-selector") {
                            label { +"Select App:" }
                            select {
                                id = "app-selector"
                                onChange = "loadReviews(this.value)"
                                apps.forEach { app ->
                                    option {
                                        value = app.id
                                        if (app.id == appIds.firstOrNull()) {
                                            selected = true
                                        }
                                        +app.name
                                    }
                                }
                            }
                        }
                    }
                    main(classes = "app-main") {
                        div(classes = "reviews-container") {
                            div {
                                id = "reviews-list"
                                div(classes = "loading") { +"Loading reviews..." }
                            }
                            div(classes = "auto-refresh") {
                                +"Auto-refreshing every 5 minutes"
                            }
                        }
                    }
                }

                script(src = "/static/js/app.js") {}
                script {
                    unsafe {
                        raw("initApp('${appIds.firstOrNull() ?: ""}');")
                    }
                }
            }
        }
    }
}

data class App(val id: String, val name: String)
