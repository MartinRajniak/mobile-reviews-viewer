import io.ktor.server.application.Application
import io.ktor.server.application.log
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlinx.coroutines.time.delay
import java.time.Duration

fun Application.configurePolling() {
    // TODO: change to minutes for production
    val pollInterval = Duration.ofSeconds(5)

    launch(Dispatchers.IO) {
        log.info("Starting poller service...")
        try {
            while (isActive) {
                try {
                    log.info("Fetching reviews...")
                    // TODO: fetch and store reviews
                    delay(pollInterval)
                } catch (e: Exception) {
                    log.error("Error in polling service", e)
                }
            }
        } finally {
            log.info("Polling service stopped.")
        }
    }
}
