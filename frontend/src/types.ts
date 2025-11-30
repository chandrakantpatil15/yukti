export type ResourceStatus = 'running' | 'stopped' | 'terminated';

export interface ResourceTags {
  environment?: string;
  project?: string;
  owner?: string;
  [key: string]: string | undefined;
}

export interface Resource {
  resourceId: string;
  instanceType: string;
  region: string;
  status: ResourceStatus;
  tags?: ResourceTags;
  costData?: any; // TODO: Add proper type for cost data
  usageMetrics?: any; // TODO: Add proper type for usage metrics
  recommendations?: any[]; // TODO: Add proper type for recommendations
}