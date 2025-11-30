export interface Resource {
  resourceId: string;
  instanceType: string;
  status: 'running' | 'stopped' | 'terminated';
  region: string;
  environment: string;
  cpuUtilization: number;
  monthlyCost: number;
  tags: string[];
  associatedResources: AssociatedResource[];
  securityCompliance: SecurityCompliance;
  billingBreakdown: BillingBreakdown;
}

export interface AssociatedResource {
  resourceType: string;
  resourceId: string;
  monthlyCost: number;
  status: string;
  tags: Record<string, string>;
}

export interface SecurityCompliance {
  score: number;
  issues: ComplianceIssue[];
  lastAudit: string;
}

export interface ComplianceIssue {
  severity: 'low' | 'medium' | 'high' | 'critical';
  category: string;
  description: string;
  recommendation: string;
}

export interface BillingBreakdown {
  compute: number;
  storage: number;
  network: number;
  associatedServices: number;
  total: number;
}

export interface CostSummary {
  totalMonthlyCost: number;
  potentialSavings: number;
  totalResources: number;
  optimizationOpportunities: number;
  averageUtilization: number;
}

export interface Recommendation {
  type: string;
  resourceId: string;
  from: string;
  to: string;
  savings: number;
  currentCost: number;
  optimizedCost: number;
  confidence: number;
}