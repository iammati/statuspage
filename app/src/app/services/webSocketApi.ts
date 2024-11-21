import { useEffect } from "react";
import { useWebSocket } from "../hooks/useWebSocket";

export default function Api(
    endpoint: string,
    callback: Function,
    data?: Record<string, any>,
) {
    const { messageData, sendMessage } = useWebSocket();
    const payload = { ...data, api: endpoint };

    sendMessage(payload);

    return useEffect(() => {
        if (messageData) {
            const json = JSON.parse(messageData);
            const { api } = json;

            if (api !== endpoint) {
                return;
            }

            try {
                return callback(api as string, json as any);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        }
    }, [messageData]);
}
