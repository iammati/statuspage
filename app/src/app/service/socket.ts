'use client';

let ws: WebSocket | undefined = undefined;

export function getWebSocket() {
    if (ws && ws.readyState === WebSocket.OPEN) {
        return ws;
    }

    ws = new WebSocket("ws://api:8080/ws");

    ws.onopen = () => {
        console.log("WebSocket connection established");
    };

    ws.onclose = () => {
        console.log("WebSocket connection closed");
        ws = undefined; // Allow reconnection if closed
    };

    ws.onerror = (error) => {
        console.error("WebSocket error:", error);
        ws = undefined; // Reset on error to allow reconnection
    };

    console.log(ws, ws?.readyState);

    return ws;
}
