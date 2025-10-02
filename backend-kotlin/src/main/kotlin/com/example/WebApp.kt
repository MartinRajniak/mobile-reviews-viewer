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
                style {
                    unsafe {
                        raw("""
                            * {
                                margin: 0;
                                padding: 0;
                                box-sizing: border-box;
                            }

                            body {
                                font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
                                background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
                                min-height: 100vh;
                                padding: 2rem;
                                color: #333;
                            }

                            .app {
                                max-width: 1200px;
                                margin: 0 auto;
                            }

                            .app-header {
                                background: white;
                                padding: 2rem;
                                border-radius: 12px;
                                box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
                                margin-bottom: 2rem;
                            }

                            h1 {
                                color: #667eea;
                                margin-bottom: 1.5rem;
                                font-size: 2rem;
                            }

                            .app-selector {
                                display: flex;
                                align-items: center;
                                gap: 1rem;
                            }

                            .app-selector label {
                                font-weight: 600;
                                color: #555;
                            }

                            .app-selector select {
                                padding: 0.75rem 1rem;
                                border: 2px solid #e2e8f0;
                                border-radius: 8px;
                                font-size: 1rem;
                                background: white;
                                cursor: pointer;
                                transition: border-color 0.2s;
                            }

                            .app-selector select:hover {
                                border-color: #667eea;
                            }

                            .app-selector select:focus {
                                outline: none;
                                border-color: #667eea;
                                box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
                            }

                            .reviews-container {
                                background: white;
                                padding: 2rem;
                                border-radius: 12px;
                                box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
                            }

                            .loading {
                                text-align: center;
                                padding: 3rem;
                                color: #667eea;
                                font-size: 1.2rem;
                            }

                            .error {
                                background: #fee;
                                color: #c33;
                                padding: 1rem;
                                border-radius: 8px;
                                margin: 1rem 0;
                            }

                            .no-reviews {
                                text-align: center;
                                padding: 3rem;
                                color: #999;
                            }

                            .review-card {
                                border: 1px solid #e2e8f0;
                                border-radius: 8px;
                                padding: 1.5rem;
                                margin-bottom: 1rem;
                                transition: box-shadow 0.2s;
                            }

                            .review-card:hover {
                                box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
                            }

                            .review-header {
                                display: flex;
                                justify-content: space-between;
                                align-items: start;
                                margin-bottom: 1rem;
                            }

                            .review-author {
                                font-weight: 600;
                                color: #333;
                            }

                            .review-rating {
                                color: #fbbf24;
                                font-size: 1.2rem;
                            }

                            .review-content {
                                color: #555;
                                line-height: 1.6;
                                margin-bottom: 1rem;
                            }

                            .review-date {
                                font-size: 0.875rem;
                                color: #999;
                            }

                            .auto-refresh {
                                text-align: center;
                                margin-top: 1rem;
                                color: #999;
                                font-size: 0.875rem;
                            }
                        """)
                    }
                }
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

                script {
                    unsafe {
                        raw("""
                            let currentAppId = '${appIds.firstOrNull() ?: ""}';
                            let refreshInterval;

                            function formatDate(dateString) {
                                const date = new Date(dateString);
                                return date.toLocaleString('en-US', {
                                    year: 'numeric',
                                    month: 'short',
                                    day: 'numeric',
                                    hour: '2-digit',
                                    minute: '2-digit'
                                });
                            }

                            function renderStars(rating) {
                                return '★'.repeat(rating) + '☆'.repeat(5 - rating);
                            }

                            async function loadReviews(appId) {
                                currentAppId = appId || currentAppId;
                                const container = document.getElementById('reviews-list');

                                try {
                                    const response = await fetch(`/api/reviews?app_id=${'$'}{currentAppId}&hours=720`);

                                    if (!response.ok) {
                                        throw new Error(`HTTP error! status: ${'$'}{response.status}`);
                                    }

                                    const reviews = await response.json();

                                    if (reviews.length === 0) {
                                        container.innerHTML = '<div class="no-reviews">No reviews found for this app in the last 30 days.</div>';
                                        return;
                                    }

                                    container.innerHTML = reviews.map(review => `
                                        <div class="review-card">
                                            <div class="review-header">
                                                <div class="review-author">${'$'}{review.author}</div>
                                                <div class="review-rating">${'$'}{renderStars(review.rating)}</div>
                                            </div>
                                            <div class="review-content">${'$'}{review.content}</div>
                                            <div class="review-date">Submitted: ${'$'}{formatDate(review.submitted_at)}</div>
                                        </div>
                                    `).join('');

                                } catch (error) {
                                    console.error('Error loading reviews:', error);
                                    container.innerHTML = `<div class="error">Failed to load reviews: ${'$'}{error.message}</div>`;
                                }
                            }

                            function startAutoRefresh() {
                                if (refreshInterval) {
                                    clearInterval(refreshInterval);
                                }
                                refreshInterval = setInterval(() => {
                                    console.log('Auto-refreshing reviews...');
                                    loadReviews();
                                }, 300000); // 5 minutes
                            }

                            // Initial load
                            loadReviews();
                            startAutoRefresh();
                        """)
                    }
                }
            }
        }
    }
}

data class App(val id: String, val name: String)
