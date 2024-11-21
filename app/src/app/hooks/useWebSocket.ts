import { useState, useEffect } from 'react';
import { getWebSocketInstance, WEBSOCKET_URL } from '../services/webSocketManager';

export function useWebSocket() {
    const [status, setStatus] = useState<'Connecting...' | 'Connected' | 'Disconnected' | 'Error'>(
        'Connecting...'
    );
    const [messageData, setMessageData] = useState<string | null>(null);

    useEffect(() => {
        const ws = getWebSocketInstance();

        const handleOpen = () => setStatus('Connected');
        const handleClose = () => setStatus('Disconnected');
        const handleError = () => setStatus('Error');
        const handleMessage = (event: MessageEvent) => setMessageData(event.data);

        ws.addEventListener('open', handleOpen);
        ws.addEventListener('close', handleClose);
        ws.addEventListener('error', handleError);
        ws.addEventListener('message', handleMessage);

        return () => {
            ws.removeEventListener('open', handleOpen);
            ws.removeEventListener('close', handleClose);
            ws.removeEventListener('error', handleError);
            ws.removeEventListener('message', handleMessage);
        };
    }, [WEBSOCKET_URL]);

    const sendMessage = (value: any) => {
        const ws = getWebSocketInstance();
        if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify(value));
        } else {
            // console.warn('WebSocket is not open. Unable to send message.');
        }
    };

    return { status, messageData, sendMessage };
}
