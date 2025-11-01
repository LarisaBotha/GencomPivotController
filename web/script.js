const API_URL = "https://pivot-api-of3d.onrender.com";

// Elements
const startBtn = document.getElementById('startBtn');
const stopBtn = document.getElementById('stopBtn');
const pivotIdInput = document.getElementById('pivotIdInput');
const sseMessageSpan = document.getElementById('sseMessageSpan');
const connectSseBtn = document.getElementById('connectSseBtn');

let evtSource;
connectSseBtn.addEventListener("click", () => {
    const pivotId = pivotIdInput.value.trim();
    if (!pivotId) return alert("Enter a Pivot IMEI");

    if (evtSource) {
        evtSource.close();
    }

    evtSource = new EventSource(`${API_URL}/api/sse?id=${pivotId}`);

    evtSource.onmessage = (event) => {
        const p = document.createElement('p');
        p.textContent = event.data;
        sseMessageSpan.appendChild(p);
    };

    evtSource.onerror = () => {
        sseMessageSpan.textContent = "Connection lost.";
        evtSource.close();
    };
});

async function sendCommand(command) {
  const id = pivotIdInput.value;
  if (!id) {
    alert("Please enter a Pivot ID!");
    return;
  }

  await fetch(`${API_URL}/api/${command}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ id })
  });
}

startBtn.addEventListener('click', () => sendCommand('start'));
stopBtn.addEventListener('click', () => sendCommand('stop'));
