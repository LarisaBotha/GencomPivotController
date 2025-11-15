const API_URL = "https://pivot-api-of3d.onrender.com"; // "http://127.0.0.1:8080"; //

// ======================
// Elements
// ======================

// Pivot Controller
const startBtn = document.getElementById('startBtn');
const stopBtn = document.getElementById('stopBtn');
const pivotIdInput = document.getElementById('pivotIdInput');

// Status Card
const statusBtn = document.getElementById("statusBtn");
const statusImeiInput = document.getElementById("statusImeiInput");
const statusSpan = document.getElementById("statusSpan");

// Register Pivot Card
const registerBtn = document.getElementById("registerBtn");
const registerImeiInput = document.getElementById("registerImeiInput");
const registerNameInput = document.getElementById("registerNameInput");
const registerStatus = document.getElementById("registerStatus");

// Command Card
const sendCommandBtn = document.getElementById("sendCommandBtn");
const commandImeiInput = document.getElementById("commandImeiInput");
const commandInput = document.getElementById("commandInput");
const commandStatus = document.getElementById("commandStatus");

// Update Pivot Card
const updateBtn = document.getElementById("updateBtn");
const updateImeiInput = document.getElementById("updateImeiInput");
const updatePositionInput = document.getElementById("updatePositionInput");
const updateSpeedInput = document.getElementById("updateSpeedInput");
const updateDirectionInput = document.getElementById("updateDirectionInput");
const updateWetInput = document.getElementById("updateWetInput");
const updateStatusInput = document.getElementById("updateStatusInput");
const updateStatus = document.getElementById("updateStatus");

// SSE Card
// const sseMessageSpan = document.getElementById('sseMessageSpan');
// const connectSseBtn = document.getElementById('connectSseBtn');
// let evtSource;

// ======================
// SSE Connection
// ======================
// connectSseBtn.addEventListener("click", () => {
//     const pivotId = pivotIdInput.value.trim();
//     if (!pivotId) return alert("Enter a Pivot IMEI");

//     if (evtSource) evtSource.close();

//     evtSource = new EventSource(`${API_URL}/api/sse?id=${pivotId}`);
//     evtSource.onmessage = (event) => {
//         const p = document.createElement('p');
//         p.textContent = event.data;
//         sseMessageSpan.appendChild(p);
//     };
//     evtSource.onerror = () => {
//         sseMessageSpan.textContent = "Connection lost.";
//         evtSource.close();
//     };
// });

// ======================
// Status Card
// ======================
statusBtn.addEventListener("click", async () => {
    const imei = statusImeiInput.value.trim();
    if (!imei) {
        statusSpan.textContent = "Please enter an IMEI.";
        return;
    }

    try {
        const res = await fetch(`${API_URL}/api/status?imei=${encodeURIComponent(imei)}`);
        if (!res.ok) throw new Error(`Server returned ${res.status}`);
        const data = await res.json();

        statusSpan.innerHTML = `Position: ${data.PositionDeg}°<br> Speed: ${data.SpeedPct}%<br> Direction: ${data.Direction}<br> Wet: ${data.Wet}<br> Status: ${data.Status}`;
    } catch (err) {
        statusSpan.textContent = `Error: ${err.message}`;
    }
});

// ======================
// Register Pivot Card
// ======================
registerBtn.addEventListener("click", async () => {
    const imei = registerImeiInput.value.trim();

    if (!imei) {
        registerStatus.textContent = "Please enter an IMEI.";
        return;
    }

    try {
        const res = await fetch(`${API_URL}/api/register`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ imei })
        });

        let data;
        try {
            data = await res.json(); // parse JSON if returned
        } catch {
            data = {};
        }

        if (res.status == 202 ) {
            registerStatus.textContent = "Pivot already registered!";
            return
        } else if (!res.ok) {
            // show error message from server JSON or fallback
            registerStatus.textContent = data.error || `Server returned ${res.status}`;
            return;
        }

        // success message
        registerStatus.textContent = "Pivot registered!";
    } catch (err) {
        registerStatus.textContent = `Error: ${err.message}`;
    }
});

// ======================
// Send Command Card
// ======================
sendCommandBtn.addEventListener("click", async () => {
    const imei = commandImeiInput.value.trim();
    const cmd = commandInput.value.trim();

    if (!imei || !cmd) {
        commandStatus.textContent = "Please enter both IMEI and command.";
        return;
    }

    try {
        const res = await fetch(`${API_URL}/api/command`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ imei, command: cmd })
        });

        if (!res.ok) throw new Error(`Server returned ${res.status}`);
        commandStatus.textContent = "Command sent successfully!";
    } catch (err) {
        commandStatus.textContent = `Error: ${err.message}`;
    }
});

// ======================
// Update Pivot Card
// ======================
updateBtn.addEventListener("click", async () => {
    const imei = updateImeiInput.value.trim();
    if (!imei) {
        updateStatus.textContent = "Please enter an IMEI.";
        return;
    }

    const payload = {
        imei: imei,
        position: parseFloat(updatePositionInput.value) || 0,
        speed: parseFloat(updateSpeedInput.value) || 0,
        direction: updateDirectionInput.value,
        wet: updateWetInput.value === "true",
        status: updateStatusInput.value || ""
    };

    try {
        const res = await fetch(`${API_URL}/api/update`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload)
        });

        if (!res.ok) throw new Error(`Server returned ${res.status}`);
        const data = await res.json();
        updateStatus.textContent = `Update successful: ${JSON.stringify(data)}`;
    } catch (err) {
        updateStatus.textContent = `Error: ${err.message}`;
    }
});
