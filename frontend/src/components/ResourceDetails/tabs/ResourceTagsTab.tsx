import React, { useState, useEffect } from 'react';
import { Plus, X } from 'lucide-react';
import api from '../../../services/api';
import { Button } from '../../ui/button';
import { Input } from '../../ui/input';

interface ResourceTagsTabProps {
  resourceId: string;
}

interface Tag {
  key: string;
  value: string;
}

const ResourceTagsTab: React.FC<ResourceTagsTabProps> = ({ resourceId }) => {
  const [tags, setTags] = useState<Tag[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [newTag, setNewTag] = useState<Tag>({ key: '', value: '' });
  const [isAdding, setIsAdding] = useState(false);

  useEffect(() => {
    fetchTags();
  }, [resourceId]);

  const fetchTags = async () => {
    try {
      setLoading(true);
      const response = await api.get(`/api/v1/resources/${resourceId}/tags`);
      setTags(response.data.tags);
    } catch (err) {
      setError('Failed to load resource tags');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleAddTag = async () => {
    try {
      await api.post(`/api/v1/resources/${resourceId}/tags`, {
        key: newTag.key,
        value: newTag.value,
      });
      setNewTag({ key: '', value: '' });
      setIsAdding(false);
      fetchTags();
    } catch (err) {
      setError('Failed to add tag');
      console.error(err);
    }
  };

  const handleDeleteTag = async (key: string) => {
    try {
      await api.delete(`/api/v1/resources/${resourceId}/tags/${key}`);
      fetchTags();
    } catch (err) {
      setError('Failed to delete tag');
      console.error(err);
    }
  };

  if (loading) {
    return <div>Loading tags...</div>;
  }

  return (
    <div className="space-y-4">
      {/* Add Tag Button */}
      {!isAdding && (
        <Button
          onClick={() => setIsAdding(true)}
          variant="outline"
          className="w-full flex items-center justify-center gap-2"
        >
          <Plus className="h-4 w-4" />
          Add New Tag
        </Button>
      )}

      {/* Add Tag Form */}
      {isAdding && (
        <div className="border rounded-lg p-4 space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Key
              </label>
              <Input
                type="text"
                value={newTag.key}
                onChange={(e) => setNewTag({ ...newTag, key: e.target.value })}
                placeholder="Enter key"
                className="mt-1"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700">
                Value
              </label>
              <Input
                type="text"
                value={newTag.value}
                onChange={(e) => setNewTag({ ...newTag, value: e.target.value })}
                placeholder="Enter value"
                className="mt-1"
              />
            </div>
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="ghost"
              onClick={() => {
                setIsAdding(false);
                setNewTag({ key: '', value: '' });
              }}
            >
              Cancel
            </Button>
            <Button
              onClick={handleAddTag}
              disabled={!newTag.key || !newTag.value}
            >
              Add Tag
            </Button>
          </div>
        </div>
      )}

      {/* Tags List */}
      <div className="space-y-2">
        {tags.map((tag) => (
          <div
            key={tag.key}
            className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
          >
            <div>
              <span className="font-medium">{tag.key}</span>
              <span className="mx-2 text-gray-400">=</span>
              <span>{tag.value}</span>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => handleDeleteTag(tag.key)}
              className="text-red-500 hover:text-red-700"
            >
              <X className="h-4 w-4" />
            </Button>
          </div>
        ))}
      </div>

      {tags.length === 0 && !isAdding && (
        <div className="text-center text-gray-500 py-4">
          No tags found for this resource
        </div>
      )}
    </div>
  );
};

export default ResourceTagsTab;