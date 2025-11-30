-- Migration to fix optimization_recommendations foreign key constraint
BEGIN;

-- 1. First remove the existing foreign key constraint
ALTER TABLE optimization_recommendations 
DROP CONSTRAINT IF EXISTS optimization_recommendations_resource_id_fkey;

-- 2. Update the resource_id column type in optimization_recommendations if needed
ALTER TABLE optimization_recommendations 
ALTER COLUMN resource_id TYPE bigint 
USING resource_id::bigint;

-- 3. Update the id column type in resources table to match
ALTER TABLE resources 
ALTER COLUMN id TYPE bigint;

-- 4. Add the foreign key constraint back with matching types
ALTER TABLE optimization_recommendations 
ADD CONSTRAINT optimization_recommendations_resource_id_fkey 
FOREIGN KEY (resource_id) 
REFERENCES resources(id) 
ON DELETE CASCADE;

COMMIT;