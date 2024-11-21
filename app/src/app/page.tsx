'use client';

import MonitoredServices from './components/MonitoredServices';
import { useWebSocket } from './hooks/useWebSocket';
import { Badge } from '@/components/ui/badge';

export default function Home() {
    const { status } = useWebSocket();

    return (
        <main className="flex min-h-screen flex-col items-center justify-between p-24">
            <section id="header">
                <Badge>
                    Status: {status}
                </Badge>
            </section>
            <MonitoredServices />
        </main>
    );
}
