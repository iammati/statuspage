// import type { Metadata } from "next";
import { Montserrat } from "next/font/google";
import "./globals.css";

const montserrat = Montserrat({ subsets: ["latin"] });

// export const metadata: Metadata = {
//     title: "Statuspage by InfraOps.dev",
//     description: "A status page for your self-hosted and public services.",
// };

export const ws = new WebSocket("ws://api:8080/ws");
// @ts-ignore
ws.onopen = () => console.log("Connected hehe") && ws.send('ping');
ws.onmessage = (msg) => console.log("Received:", msg.data);
ws.onclose = () => console.log("Disconnected");
ws.onerror = e => console.error("Error", e);

export default function RootLayout({ children }: Readonly<{children: React.ReactNode}>) {


    return (
        <html lang="en">
            <body className={montserrat.className}>
                {children}
            </body>
        </html>
    );
}
