import React from 'react';
import { Card, CardHeader, CardContent, CardFooter } from './ui/Card';
import { StatusBadge } from './StatusBadge/StatusBadge';
import { Resource } from '../types';
import { DollarSign, Cpu, Database } from 'lucide-react';

interface ResourceCardProps {
  resource: Resource;
  onClick: () => void;
  selected?: boolean;
}

export const ResourceCard: React.FC<ResourceCardProps> = ({
  resource,
  onClick,
  selected,
}) => {
  return (
    <Card
      interactive
      selected={selected}
      onClick={onClick}
      className="flex flex-col h-full"
    >
      <CardHeader>
        <div className="flex justify-between items-start">
          <div>
            <h3 className="text-heading-md text-neutral-900 font-semibold">
              {resource.resourceId}
            </h3>
            <p className="text-body-sm text-neutral-500">
              {resource.instanceType}
            </p>
          </div>
          <StatusBadge 
            status={resource.status.toLowerCase() as any} 
            label={resource.status} 
            size="sm"
          />
        </div>
      </CardHeader>

      <CardContent className="flex-1">
        <div className="grid grid-cols-2 gap-4">
          <div className="flex items-center">
            <DollarSign className="w-5 h-5 text-neutral-400 mr-2" />
            <div>
              <p className="text-sm font-medium text-neutral-900">
                ${resource.monthlyCost}
              </p>
              <p className="text-xs text-neutral-500">Monthly Cost</p>
            </div>
          </div>
          <div className="flex items-center">
            <Cpu className="w-5 h-5 text-neutral-400 mr-2" />
            <div>
              <p className="text-sm font-medium text-neutral-900">
                {resource.cpuUtilization}%
              </p>
              <p className="text-xs text-neutral-500">CPU Usage</p>
            </div>
          </div>
        </div>

        <div className="mt-4">
          <div className="flex items-center text-sm text-neutral-600">
            <Database className="w-4 h-4 mr-1.5" />
            {resource.region}
          </div>
        </div>
      </CardContent>

      <CardFooter className="bg-neutral-50">
        <div className="flex justify-between items-center text-sm">
          <span className="text-neutral-500">
            Last Updated: {new Date().toLocaleTimeString()}
          </span>
          {resource.tags?.length > 0 && (
            <div className="flex gap-2">
              {resource.tags.slice(0, 2).map((tag, index) => (
                <span
                  key={index}
                  className="inline-flex items-center px-2 py-1 rounded-full bg-primary-100 text-primary-700 text-xs"
                >
                  {tag}
                </span>
              ))}
              {resource.tags.length > 2 && (
                <span className="text-neutral-500">
                  +{resource.tags.length - 2}
                </span>
              )}
            </div>
          )}
        </div>
      </CardFooter>
    </Card>
  );
};