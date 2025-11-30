import React, { useState } from 'react';
import { useQuery } from 'react-query';
import { 
  Search,
  Filter,
  Server,
  Database,
  HardDrive,
  Cloud,
  AlertCircle,
} from 'lucide-react';
import { api } from '../services/api';
import ResourcePanel from '../components/ResourceDetails/ResourcePanel';
import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Alert } from '../components/ui/alert';

interface Resource {
  id: string;
  name: string;
  type: string;
  region: string;
  status: string;
  estimated_savings: number;
  tags: Record<string, string>;
}

const getResourceIcon = (type: string) => {
  switch (type.split('::')[1]?.toLowerCase()) {
    case 'ec2':
      return <Server className="w-5 h-5" />;
    case 'rds':
      return <Database className="w-5 h-5" />;
    case 'ebs':
      return <HardDrive className="w-5 h-5" />;
    default:
      return <Cloud className="w-5 h-5" />;
  }
};

const ResourcesPage: React.FC = () => {
  const [selectedResource, setSelectedResource] = useState<Resource | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filters, setFilters] = useState<{
    region?: string;
    type?: string;
    tags?: Record<string, string>;
  }>({});

  const { data: resources, isLoading, error } = useQuery(
    ['resources', filters],
    () => api.getResources(filters),
    {
      refetchInterval: 30000, // Refetch every 30 seconds
    }
  );

  const filteredResources = React.useMemo(() => {
    if (!resources) return [];
    return resources.filter((resource: Resource) =>
      resource.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      resource.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      Object.values(resource.tags || {}).some(tag =>
        tag.toLowerCase().includes(searchTerm.toLowerCase())
      )
    );
  }, [resources, searchTerm]);

  if (error) {
    return (
      <Alert variant="destructive">
        Error loading resources: {(error as Error).message}
      </Alert>
    );
  }
      region: 'us-east-1',
      estimated_savings: 12.50,
    },
  ];

  return (
    <div className="container mx-auto p-6">
      {/* Header */}
      <div className="flex justify-between items-center mb-6">
        <h1 className="text-2xl font-bold">Cloud Resources</h1>
        <div className="flex items-center gap-4">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 w-4 h-4" />
            <Input
              type="text"
              placeholder="Search resources..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10 w-64"
            />
          </div>
          <Button variant="outline" onClick={() => {/* Open filters modal */}}>
            <Filter className="w-4 h-4 mr-2" />
            Filters
          </Button>
        </div>
      </div>

      {/* Resources Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {isLoading ? (
          // Loading skeletons
          Array.from({ length: 6 }).map((_, i) => (
            <div
              key={i}
              className="animate-pulse bg-gray-100 rounded-lg p-6 h-40"
            />
          ))
        ) : (
          filteredResources.map((resource: Resource) => (
            <div
              key={resource.id}
              onClick={() => setSelectedResource(resource)}
              className="cursor-pointer p-4 border rounded-lg hover:border-blue-500 transition-colors"
            >
              <div className="flex items-start gap-3">
                <div className="p-2 bg-gray-50 rounded-lg">
                  {getResourceIcon(resource.type)}
                </div>
                <div className="flex-1">
                  <div className="flex justify-between items-start">
                    <div>
                      <h3 className="font-medium">{resource.name}</h3>
                      <p className="text-sm text-gray-500">{resource.id}</p>
                    </div>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      resource.status === 'running' ? 'bg-green-100 text-green-800' :
                      resource.status === 'stopped' ? 'bg-red-100 text-red-800' :
                      'bg-gray-100 text-gray-800'
                    }`}>
                      {resource.status}
                    </span>
                  </div>
                  
                  <div className="mt-2">
                    <p className="text-sm text-gray-600">{resource.region}</p>
                    <p className="text-sm text-gray-600">{resource.type}</p>
                  </div>
                  
                  {resource.estimated_savings > 0 && (
                    <div className="mt-2 flex items-center text-amber-600 text-sm">
                      <AlertCircle className="w-4 h-4 mr-1" />
                      Potential savings: ${resource.estimated_savings.toFixed(2)}/month
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Resource Details Side Panel */}
      {selectedResource && (
        <ResourcePanel
          resourceId={selectedResource.id}
          resourceType={selectedResource.type}
          onClose={() => setSelectedResource(null)}
        />
      )}
    </div>
  );
};

export default ResourcesPage;
};

export default ResourcesPage;