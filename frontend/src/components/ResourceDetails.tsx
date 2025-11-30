import React, { useState } from 'react';
import { 
  LayoutDashboard, 
  DollarSign, 
  Lightbulb 
} from 'lucide-react';
import { Resource } from '../types';
import { SlideOver } from './ui/SlideOver';
import { TabGroup } from './ui/TabGroup';
import { StatusBadge } from './StatusBadge/StatusBadge';
import { Card } from './ui/Card';
import { CostChart } from './CostChart';
import { UsageMetrics } from './UsageMetrics';
import { RecommendationList } from './RecommendationList';

interface ResourceDetailsProps {
  resource: Resource;
  isOpen: boolean;
  onClose: () => void;
}

const tabs = [
  { id: 'overview', label: 'Overview', icon: <LayoutDashboard className="w-4 h-4" /> },
  { id: 'cost', label: 'Cost & Usage', icon: <DollarSign className="w-4 h-4" /> },
  { id: 'recommendations', label: 'Recommendations', icon: <Lightbulb className="w-4 h-4" /> }
];

export const ResourceDetails = ({ resource, isOpen, onClose }: ResourceDetailsProps) => {
  const [activeTab, setActiveTab] = useState('overview');

  return (
    <SlideOver
      isOpen={isOpen}
      onClose={onClose}
      title={resource.resourceId}
      subtitle={`${resource.instanceType} • ${resource.region}`}
    >
      <div className="px-6 py-4 border-b border-neutral-200 dark:border-neutral-700">
        <TabGroup
          tabs={tabs}
          activeTab={activeTab}
          onChange={setActiveTab}
        />
      </div>
      {activeTab === 'overview' && (
        <div className="p-6 space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <Card title="Resource Details">
              <dl className="space-y-4">
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Status</dt>
                  <dd className="mt-1">
                    <StatusBadge status={resource.status} />
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Instance Type</dt>
                  <dd className="mt-1 text-sm text-neutral-900 dark:text-neutral-100">{resource.instanceType}</dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Region</dt>
                  <dd className="mt-1 text-sm text-neutral-900 dark:text-neutral-100">{resource.region}</dd>
                </div>
              </dl>
            </Card>
            <Card title="Tags">
              <dl className="space-y-4">
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Environment</dt>
                  <dd className="mt-1 text-sm text-neutral-900 dark:text-neutral-100">
                    {resource.tags?.environment || 'Untagged'}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Project</dt>
                  <dd className="mt-1 text-sm text-neutral-900 dark:text-neutral-100">
                    {resource.tags?.project || 'Untagged'}
                  </dd>
                </div>
                <div>
                  <dt className="text-sm font-medium text-neutral-500 dark:text-neutral-400">Owner</dt>
                  <dd className="mt-1 text-sm text-neutral-900 dark:text-neutral-100">
                    {resource.tags?.owner || 'Unassigned'}
                  </dd>
                </div>
              </dl>
            </Card>
          </div>
        </div>
      )}

      {activeTab === 'cost' && (
        <div className="p-6 space-y-6">
          <Card title="Cost Overview">
            <CostChart data={resource.costData} />
          </Card>
          <Card title="Usage Details">
            <UsageMetrics metrics={resource.usageMetrics} />
          </Card>
        </div>
      )}

      {activeTab === 'recommendations' && (
        <div className="p-6">
          <Card title="Cost Optimization Recommendations">
            <RecommendationList recommendations={resource.recommendations} />
          </Card>
        </div>
      )}
    </SlideOver>
  );
};