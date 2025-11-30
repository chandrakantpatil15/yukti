import React, { useState } from 'react';
import { Tabs, TabsList, TabsTrigger, TabsContent } from '../ui/tabs';
import { Card } from '../ui/card';
import { ScrollArea } from '../ui/scroll-area';
import { Button } from '../ui/button';
import { ChevronRight, X, Code, FileText, History, Tag } from 'lucide-react';
import IaCGeneratorTab from './IaCGeneratorTab';
import ResourceInfoTab from './ResourceInfoTab';
import ResourceHistoryTab from './ResourceHistoryTab';
import ResourceTagsTab from './ResourceTagsTab';

interface ResourcePanelProps {
  resourceId: string;
  resourceType: string;
  onClose: () => void;
}

const ResourcePanel: React.FC<ResourcePanelProps> = ({
  resourceId,
  resourceType,
  onClose,
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div
      className={`fixed right-0 top-0 h-screen bg-white border-l border-gray-200 shadow-lg transition-all duration-300 ${
        isExpanded ? 'w-2/3' : 'w-1/3'
      }`}
    >
      {/* Header */}
      <div className="h-16 border-b border-gray-200 flex items-center justify-between px-4">
        <div className="flex items-center gap-2">
          <h2 className="text-lg font-semibold">{resourceType}</h2>
          <span className="text-sm text-gray-500">{resourceId}</span>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setIsExpanded(!isExpanded)}
          >
            <ChevronRight
              className={`h-4 w-4 transition-transform ${
                isExpanded ? 'rotate-180' : ''
              }`}
            />
          </Button>
          <Button variant="ghost" size="icon" onClick={onClose}>
            <X className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Content */}
      <ScrollArea className="h-[calc(100vh-4rem)]">
        <div className="p-4">
          <Tabs defaultValue="details" className="w-full">
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="details">Details</TabsTrigger>
              <TabsTrigger value="iac">IaC</TabsTrigger>
              <TabsTrigger value="history">History</TabsTrigger>
              <TabsTrigger value="tags">Tags</TabsTrigger>
            </TabsList>

            <TabsContent value="details">
              <ResourceInfoTab resourceId={resourceId} />
            </TabsContent>

            <TabsContent value="iac">
              <IaCGeneratorTab resourceId={resourceId} resourceType={resourceType} />
            </TabsContent>

            <TabsContent value="history">
              <ResourceHistoryTab resourceId={resourceId} />
            </TabsContent>

            <TabsContent value="tags">
              <ResourceTagsTab resourceId={resourceId} />
            </TabsContent>
          </Tabs>
        </div>
      </ScrollArea>
    </div>
  );
};

export default ResourcePanel;