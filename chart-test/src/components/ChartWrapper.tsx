import { type ReactNode } from 'react';

interface ChartWrapperProps {
  title: string;
  description: string;
  children: ReactNode;
}

export function ChartWrapper({ title, description, children }: ChartWrapperProps) {
  return (
    <div className="card">
      <div className="card-header">
        <h3 className="card-title">{title}</h3>
        <p className="card-description">{description}</p>
      </div>
      <div className="card-content">
        {children}
      </div>
    </div>
  );
}
