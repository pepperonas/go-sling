package io.celox.gosling

import android.app.*
import android.content.Intent
import android.os.Build
import android.os.Environment
import android.os.IBinder
import androidx.core.app.NotificationCompat
import org.java_websocket.client.WebSocketClient
import org.java_websocket.handshake.ServerHandshake
import org.json.JSONObject
import java.io.*
import java.net.HttpURLConnection
import java.net.URI
import java.net.URL
import java.text.SimpleDateFormat
import java.util.*
import java.util.zip.ZipInputStream

class ReceiverService : Service() {

    companion object {
        var logCallback: ((String) -> Unit)? = null
        var statusCallback: ((String) -> Unit)? = null
        private const val CHANNEL_ID = "gosling_receiver"
        private const val NOTIFICATION_ID = 1
    }

    private var wsClient: WebSocketClient? = null
    private var server = ""
    private var pin = ""
    private var outputFolder = "go-sling"
    private var autoExtract = true
    private var sessionCookie: String? = null
    private var running = false
    private var peerName = ""

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()
        createNotificationChannel()
        peerName = generateName()
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        server = intent?.getStringExtra("server") ?: return START_NOT_STICKY
        pin = intent.getStringExtra("pin") ?: ""
        outputFolder = intent.getStringExtra("output") ?: "go-sling"
        autoExtract = intent.getBooleanExtra("extract", true)

        val notification = buildNotification("Connecting...")
        startForeground(NOTIFICATION_ID, notification)

        running = true
        Thread { connectLoop() }.start()

        return START_STICKY
    }

    override fun onDestroy() {
        running = false
        wsClient?.close()
        super.onDestroy()
    }

    private fun connectLoop() {
        // Authenticate first
        if (!authenticate()) {
            log("Authentication failed")
            statusCallback?.invoke("Auth failed")
            stopSelf()
            return
        }

        while (running) {
            try {
                connectWebSocket()
            } catch (e: Exception) {
                log("Connection error: ${e.message}")
            }
            if (running) {
                log("Reconnecting in 5s...")
                Thread.sleep(5000)
            }
        }
    }

    private fun authenticate(): Boolean {
        try {
            // Check if auth is required
            val statusUrl = URL("http://$server/api/auth/status")
            val conn = statusUrl.openConnection() as HttpURLConnection
            conn.connectTimeout = 5000
            val resp = conn.inputStream.bufferedReader().readText()
            val json = JSONObject(resp)
            if (!json.getBoolean("required")) {
                log("No authentication required")
                return true
            }
        } catch (e: Exception) {
            log("Cannot reach server: ${e.message}")
            return false
        }

        if (pin.isEmpty()) {
            log("Server requires PIN but none provided")
            return false
        }

        try {
            val authUrl = URL("http://$server/api/auth")
            val conn = authUrl.openConnection() as HttpURLConnection
            conn.requestMethod = "POST"
            conn.setRequestProperty("Content-Type", "application/json")
            conn.doOutput = true
            conn.outputStream.write("""{"pin":"$pin","remember":true}""".toByteArray())

            if (conn.responseCode == 200) {
                val cookies = conn.headerFields["Set-Cookie"]
                cookies?.forEach { cookie ->
                    if (cookie.startsWith("gosling_session=")) {
                        sessionCookie = cookie.split(";")[0]
                    }
                }
                log("Authenticated")
                return true
            } else {
                log("Auth failed: ${conn.responseCode}")
                return false
            }
        } catch (e: Exception) {
            log("Auth error: ${e.message}")
            return false
        }
    }

    private fun connectWebSocket() {
        val uri = URI("ws://$server/ws")
        val headers = mutableMapOf<String, String>()
        sessionCookie?.let { headers["Cookie"] = it }

        val latch = java.util.concurrent.CountDownLatch(1)

        wsClient = object : WebSocketClient(uri, headers) {
            override fun onOpen(handshake: ServerHandshake?) {
                log("Connected as '$peerName'")
                statusCallback?.invoke("Connected")
                updateNotification("Connected — waiting for files")

                val joinMsg = JSONObject().apply {
                    put("type", "join")
                    put("payload", JSONObject().apply {
                        put("name", peerName)
                        put("os", "Android Phone")
                        put("browser", "Android App")
                        put("headless", true)
                    })
                }
                send(joinMsg.toString())
            }

            override fun onMessage(message: String?) {
                message ?: return
                try {
                    val msg = JSONObject(message)
                    handleMessage(msg)
                } catch (_: Exception) {}
            }

            override fun onClose(code: Int, reason: String?, remote: Boolean) {
                log("Disconnected")
                statusCallback?.invoke("Disconnected")
                updateNotification("Disconnected")
                latch.countDown()
            }

            override fun onError(ex: Exception?) {
                log("WS error: ${ex?.message}")
                latch.countDown()
            }
        }

        wsClient?.connect()
        latch.await() // Block until disconnected
    }

    private fun handleMessage(msg: JSONObject) {
        when (msg.optString("type")) {
            "welcome" -> {
                val id = msg.optJSONObject("payload")?.optString("id") ?: "?"
                log("Registered as peer $id")
            }
            "peer-list" -> {
                val peers = msg.optJSONArray("peers")
                val count = peers?.length() ?: 0
                log("Peers online: $count")
            }
            "file-ready" -> {
                val payload = msg.optJSONObject("payload") ?: return
                val fileId = payload.optString("id")
                val fileName = payload.optString("name", "unknown")
                val fileSize = payload.optLong("size", 0)
                log("Incoming: $fileName (${formatSize(fileSize)})")
                Thread { downloadFile(fileId, fileName) }.start()
            }
        }
    }

    private fun downloadFile(fileId: String, fileName: String) {
        try {
            val url = URL("http://$server/api/download/$fileId")
            val conn = url.openConnection() as HttpURLConnection
            conn.connectTimeout = 10000
            conn.readTimeout = 300000
            sessionCookie?.let { conn.setRequestProperty("Cookie", it) }

            val downloadsDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS)
            val outputDir = File(downloadsDir, outputFolder)
            outputDir.mkdirs()

            val file = File(outputDir, fileName)
            var size = 0L

            conn.inputStream.use { input ->
                FileOutputStream(file).use { output ->
                    val buffer = ByteArray(65536)
                    var read: Int
                    while (input.read(buffer).also { read = it } != -1) {
                        output.write(buffer, 0, read)
                        size += read
                    }
                }
            }

            log("Saved: $fileName (${formatSize(size)})")
            showFileNotification(fileName, size)

            // Auto-extract ZIP
            if (autoExtract && fileName.lowercase().endsWith(".zip")) {
                extractZip(file, outputDir)
            }
        } catch (e: Exception) {
            log("Download failed: ${e.message}")
        }
    }

    private fun extractZip(zipFile: File, outputDir: File) {
        try {
            val extractDir = File(outputDir, zipFile.nameWithoutExtension)
            extractDir.mkdirs()

            ZipInputStream(FileInputStream(zipFile)).use { zis ->
                var entry = zis.nextEntry
                while (entry != null) {
                    val outFile = File(extractDir, entry.name)
                    // Prevent zip slip
                    if (!outFile.canonicalPath.startsWith(extractDir.canonicalPath)) {
                        log("Skipping unsafe zip entry: ${entry.name}")
                        entry = zis.nextEntry
                        continue
                    }
                    if (entry.isDirectory) {
                        outFile.mkdirs()
                    } else {
                        outFile.parentFile?.mkdirs()
                        FileOutputStream(outFile).use { fos ->
                            zis.copyTo(fos)
                        }
                    }
                    entry = zis.nextEntry
                }
            }

            zipFile.delete()
            log("Extracted: ${zipFile.name} → ${extractDir.name}/")
        } catch (e: Exception) {
            log("Extract failed: ${e.message}")
        }
    }

    private fun generateName(): String {
        val adjectives = listOf("Turbo", "Flash", "Blitz", "Rapid", "Bolt")
        val nouns = listOf("Catcher", "Receiver", "Vault", "Pocket", "Inbox")
        val adj = adjectives.random()
        val noun = nouns.random()
        val suffix = (1..3).map { "abcdefghijklmnopqrstuvwxyz0123456789".random() }.joinToString("")
        return "$adj-$noun-$suffix"
    }

    private fun formatSize(bytes: Long): String {
        val units = arrayOf("B", "KB", "MB", "GB")
        var size = bytes.toDouble()
        var i = 0
        while (size >= 1024 && i < units.size - 1) { size /= 1024; i++ }
        return "%.1f %s".format(size, units[i])
    }

    private fun log(msg: String) {
        val ts = SimpleDateFormat("HH:mm:ss", Locale.getDefault()).format(Date())
        val line = "[$ts] $msg"
        android.util.Log.d("GoSling", line)
        logCallback?.invoke(line)
    }

    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(CHANNEL_ID, "go-sling Receiver", NotificationManager.IMPORTANCE_LOW)
            channel.description = "Background file receiver"
            getSystemService(NotificationManager::class.java).createNotificationChannel(channel)
        }
    }

    private fun buildNotification(text: String): Notification {
        val intent = Intent(this, MainActivity::class.java)
        val pending = PendingIntent.getActivity(this, 0, intent, PendingIntent.FLAG_IMMUTABLE)

        return NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("go-sling")
            .setContentText(text)
            .setSmallIcon(android.R.drawable.stat_sys_download)
            .setContentIntent(pending)
            .setOngoing(true)
            .build()
    }

    private fun updateNotification(text: String) {
        val notification = buildNotification(text)
        getSystemService(NotificationManager::class.java).notify(NOTIFICATION_ID, notification)
    }

    private fun showFileNotification(fileName: String, size: Long) {
        val notification = NotificationCompat.Builder(this, CHANNEL_ID)
            .setContentTitle("File received")
            .setContentText("$fileName (${formatSize(size)})")
            .setSmallIcon(android.R.drawable.stat_sys_download_done)
            .setAutoCancel(true)
            .build()

        getSystemService(NotificationManager::class.java)
            .notify(System.currentTimeMillis().toInt(), notification)
    }
}
