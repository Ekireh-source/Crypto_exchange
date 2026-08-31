'use client';

import { useMemo } from 'react';
import { LineChart, Line, ResponsiveContainer, YAxis } from 'recharts';

interface PortfolioChartProps {
  currentBalance: number;
}

export default function PortfolioChart({ currentBalance }: PortfolioChartProps) {
  // Generate a realistic-looking mock chart data series that ends at the current balance
  const data = useMemo(() => {
    const points = [];
    let prev = currentBalance * 0.85; // Start roughly 15% lower 7 days ago
    
    // Add 30 data points (simulating roughly 4 points a day for 7 days)
    for (let i = 0; i < 30; i++) {
      if (i === 29) {
        points.push({ value: currentBalance });
      } else {
        // Random walk towards the current balance
        const volatility = currentBalance * 0.02; // 2% volatility
        const change = (Math.random() - 0.4) * volatility; 
        const trend = (currentBalance - prev) / (30 - i);
        prev = prev + change + trend;
        points.push({ value: prev });
      }
    }
    return points;
  }, [currentBalance]);

  if (currentBalance === 0) return null;

  return (
    <div className="h-[120px] w-full -mt-4 mb-4">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <YAxis domain={['dataMin', 'dataMax']} hide />
          <Line 
            type="monotone" 
            dataKey="value" 
            stroke="#10b981" // emerald-500
            strokeWidth={3} 
            dot={false}
            animationDuration={2000}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
