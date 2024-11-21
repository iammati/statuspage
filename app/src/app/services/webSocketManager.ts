let socket: WebSocket | null = null;

export const WEBSOCKET_URL = 'ws://localhost:8080/ws';

export function getWebSocketInstance(): WebSocket {
    if (!socket || socket.readyState === WebSocket.CLOSED) {
        socket = new WebSocket(WEBSOCKET_URL);
    }
    return socket;
}
