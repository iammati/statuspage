'use client';

import React, { useEffect } from 'react';
import { TrendingUp } from 'lucide-react';
import { Area, AreaChart, XAxis } from 'recharts';

import {
    Card,
    CardContent,
    CardDescription,
    CardFooter,
    CardHeader,
    CardTitle,
} from '@/components/ui/card';
import {
    ChartConfig,
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
} from '@/components/ui/chart';
import { ws } from '../layout';

function MonitoredServices() {
    console.log(ws);
    const chartData = [
        { label: '10:00', dnsResolutionTime: 55, httpTime: 40, tcpConnectionTime: 22 },
        { label: '10:05', dnsResolutionTime: 74, httpTime: 50, tcpConnectionTime: 14 },
        { label: '10:10', dnsResolutionTime: 58, httpTime: 30, tcpConnectionTime: 18 },
        { label: '10:15', dnsResolutionTime: 51, httpTime: 35, tcpConnectionTime: 16 },
    ];

    const chartConfig = {
        tcpConnectionTime: {
            label: 'TCP Connection (ms)',
            color: '#e74c3c',
        },
        httpTime: {
            label: 'HTTP Time (ms)',
            color: '#3498db',
        },
        dnsResolutionTime: {
            label: 'DNS Resolution Time (ms)',
            color: "#2563eb",
        },
    } satisfies ChartConfig;

    return (
        <div className="flex flex-col items-center justify-center">
            <Card>
                <CardHeader>
                    <CardTitle>Performance</CardTitle>
                    <CardDescription>
                        See real-time performance metrics data.
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <ChartContainer config={chartConfig} className="min-h-[300px]">
                        <AreaChart
                            data={chartData}
                            margin={{ left: 12, right: 12 }}
                        >
                            <XAxis
                                dataKey="label"
                                tickLine={false}
                                axisLine={false}
                                allowDecimals={true}
                                tickMargin={8}
                                tickFormatter={(value) => value}
                                interval={'preserveStartEnd'}
                            />
                            <ChartTooltip cursor={false} content={<ChartTooltipContent />} />
                            <defs>
                                {Object.entries(chartConfig).map(([key, { color }]) => (
                                    <linearGradient id={`fill${key}`} x1="0" y1="0" x2="0" y2="1" key={key}>
                                        <stop offset="5%" stopColor={color} stopOpacity={0.8} />
                                        <stop offset="95%" stopColor={color} stopOpacity={0.1} />
                                    </linearGradient>
                                ))}
                            </defs>
                            {Object.keys(chartConfig).map((key) => (
                                <Area
                                    key={key}
                                    dataKey={key}
                                    type="natural"
                                    fill={`url(#fill${key})`}
                                    fillOpacity={0.4}
                                    stroke={chartConfig[key].color}
                                />
                            ))}
                        </AreaChart>
                    </ChartContainer>
                </CardContent>
                <CardFooter>
                    <div className="flex w-full items-start gap-2 text-sm">
                    <div className="grid gap-2">
                        <div className="flex items-center gap-2 font-medium leading-none">
                            Trending up by 5.2% this month <TrendingUp className="h-4 w-4" />
                        </div>
                        <div className="flex items-center gap-2 leading-none text-muted-foreground">
                            {chartData[0].label} - {chartData[chartData.length - 1].label}
                        </div>
                    </div>
                    </div>
                </CardFooter>
            </Card>
        </div>
    );
}

export default MonitoredServices;
