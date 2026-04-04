package io.celox.gosling

import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Build
import android.os.Bundle
import android.widget.Button
import android.widget.ScrollView
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity
import androidx.core.app.ActivityCompat
import com.google.android.material.materialswitch.MaterialSwitch
import com.google.android.material.textfield.TextInputEditText

class MainActivity : AppCompatActivity() {

    private lateinit var btnToggle: Button
    private lateinit var txtStatus: TextView
    private lateinit var txtLog: TextView
    private lateinit var editServer: TextInputEditText
    private lateinit var editPin: TextInputEditText
    private lateinit var editOutput: TextInputEditText
    private lateinit var switchExtract: MaterialSwitch
    private lateinit var scrollLog: ScrollView

    private var isRunning = false

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)

        btnToggle = findViewById(R.id.btnToggle)
        txtStatus = findViewById(R.id.txtStatus)
        txtLog = findViewById(R.id.txtLog)
        editServer = findViewById(R.id.editServer)
        editPin = findViewById(R.id.editPin)
        editOutput = findViewById(R.id.editOutput)
        switchExtract = findViewById(R.id.switchExtract)
        scrollLog = txtLog.parent as ScrollView

        // Load saved preferences
        val prefs = getSharedPreferences("gosling", MODE_PRIVATE)
        editServer.setText(prefs.getString("server", ""))
        editPin.setText(prefs.getString("pin", ""))
        editOutput.setText(prefs.getString("output", "go-sling"))
        switchExtract.isChecked = prefs.getBoolean("extract", true)

        // Request notification permission (Android 13+)
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            if (checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS) != PackageManager.PERMISSION_GRANTED) {
                ActivityCompat.requestPermissions(this, arrayOf(Manifest.permission.POST_NOTIFICATIONS), 1)
            }
        }

        btnToggle.setOnClickListener { toggle() }

        // Listen for log messages from the service
        ReceiverService.logCallback = { msg ->
            runOnUiThread {
                txtLog.append("$msg\n")
                scrollLog.post { scrollLog.fullScroll(ScrollView.FOCUS_DOWN) }
            }
        }

        ReceiverService.statusCallback = { status ->
            runOnUiThread {
                txtStatus.text = status
                txtStatus.setTextColor(
                    if (status.startsWith("Connected")) 0xFF22c55e.toInt()
                    else 0xFF9494b0.toInt()
                )
            }
        }
    }

    private fun toggle() {
        if (isRunning) {
            stopService(Intent(this, ReceiverService::class.java))
            isRunning = false
            btnToggle.text = "Start Receiving"
            txtStatus.text = "Disconnected"
            txtStatus.setTextColor(0xFF9494b0.toInt())
        } else {
            val server = editServer.text.toString().trim()
            if (server.isEmpty()) {
                editServer.error = "Required"
                return
            }

            // Save preferences
            getSharedPreferences("gosling", MODE_PRIVATE).edit().apply {
                putString("server", server)
                putString("pin", editPin.text.toString())
                putString("output", editOutput.text.toString())
                putBoolean("extract", switchExtract.isChecked)
                apply()
            }

            val intent = Intent(this, ReceiverService::class.java).apply {
                putExtra("server", server)
                putExtra("pin", editPin.text.toString())
                putExtra("output", editOutput.text.toString())
                putExtra("extract", switchExtract.isChecked)
            }

            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                startForegroundService(intent)
            } else {
                startService(intent)
            }

            isRunning = true
            btnToggle.text = "Stop"
            txtStatus.text = "Connecting..."
            txtStatus.setTextColor(0xFFf59e0b.toInt())
        }
    }
}
