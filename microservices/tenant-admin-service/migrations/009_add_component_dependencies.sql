-- Migration: Add component dependency tracking
-- Created: 2025-01-25
-- Description: Adds dependency_edges table for service dependency mapping

-- Create dependency_edges table for tracking component dependencies
CREATE TABLE IF NOT EXISTS dependency_edges (
    id BIGSERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL,
    from_component_id UUID NOT NULL,
    to_component_id UUID NOT NULL,
    dependency_type VARCHAR(20) NOT NULL DEFAULT 'hard', -- 'hard' or 'soft'
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    -- Prevent duplicate dependencies
    UNIQUE(from_component_id, to_component_id),

    -- Foreign key constraints
    CONSTRAINT fk_from_component FOREIGN KEY (from_component_id)
        REFERENCES saas_components(id) ON DELETE CASCADE,
    CONSTRAINT fk_to_component FOREIGN KEY (to_component_id)
        REFERENCES saas_components(id) ON DELETE CASCADE,

    -- Check constraints
    CONSTRAINT chk_dependency_type CHECK (dependency_type IN ('hard', 'soft')),
    CONSTRAINT chk_different_components CHECK (from_component_id != to_component_id)
);

-- Create indexes for performance
CREATE INDEX idx_dependency_edges_tenant_id ON dependency_edges(tenant_id);
CREATE INDEX idx_dependency_edges_from_component ON dependency_edges(from_component_id);
CREATE INDEX idx_dependency_edges_to_component ON dependency_edges(to_component_id);
CREATE INDEX idx_dependency_edges_type ON dependency_edges(dependency_type);

-- Create composite index for graph queries
CREATE INDEX idx_dependency_edges_tenant_from ON dependency_edges(tenant_id, from_component_id);
CREATE INDEX idx_dependency_edges_tenant_to ON dependency_edges(tenant_id, to_component_id);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_dependency_edges_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_dependency_edges_updated_at
    BEFORE UPDATE ON dependency_edges
    FOR EACH ROW
    EXECUTE FUNCTION update_dependency_edges_updated_at();

-- Add comment for documentation
COMMENT ON TABLE dependency_edges IS 'Tracks service dependencies between components for dependency mapping and impact analysis';
COMMENT ON COLUMN dependency_edges.dependency_type IS 'Type of dependency: hard (required) or soft (optional/graceful degradation)';
COMMENT ON COLUMN dependency_edges.from_component_id IS 'Component that depends on another (source)';
COMMENT ON COLUMN dependency_edges.to_component_id IS 'Component being depended upon (target/dependency)';

-- Create view for easy dependency graph queries
CREATE OR REPLACE VIEW component_dependency_graph AS
SELECT
    de.id,
    de.tenant_id,
    de.from_component_id,
    from_comp.name AS from_component_name,
    from_comp.status AS from_component_status,
    de.to_component_id,
    to_comp.name AS to_component_name,
    to_comp.status AS to_component_status,
    de.dependency_type,
    de.description,
    de.created_at,
    de.updated_at
FROM dependency_edges de
JOIN saas_components from_comp ON de.from_component_id = from_comp.id
JOIN saas_components to_comp ON de.to_component_id = to_comp.id
WHERE from_comp.deleted_at IS NULL
  AND to_comp.deleted_at IS NULL;

COMMENT ON VIEW component_dependency_graph IS 'Enriched view of component dependencies with component details';
