-- Create "incident_templates" table
CREATE TABLE "incident_templates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "title" text NOT NULL,
  "message" text NULL,
  "impact" text NULL DEFAULT 'minor',
  "severity" text NULL DEFAULT 'low',
  "is_active" boolean NULL DEFAULT true,
  "created_by" bigint NULL,
  "updated_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_incident_templates_deleted_at" to table: "incident_templates"
CREATE INDEX "idx_incident_templates_deleted_at" ON "incident_templates" ("deleted_at");
-- Create index "idx_incident_templates_tenant_id" to table: "incident_templates"
CREATE INDEX "idx_incident_templates_tenant_id" ON "incident_templates" ("tenant_id");
-- Create "incidents" table
CREATE TABLE "incidents" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "status" text NULL DEFAULT 'investigating',
  "impact" text NULL DEFAULT 'minor',
  "severity" text NULL DEFAULT 'low',
  "is_visible" boolean NULL DEFAULT true,
  "started_at" timestamptz NOT NULL,
  "resolved_at" timestamptz NULL,
  "created_by" bigint NULL,
  "updated_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_incidents_deleted_at" to table: "incidents"
CREATE INDEX "idx_incidents_deleted_at" ON "incidents" ("deleted_at");
-- Create index "idx_resolved_at" to table: "incidents"
CREATE INDEX "idx_resolved_at" ON "incidents" ("resolved_at");
-- Create index "idx_started_at" to table: "incidents"
CREATE INDEX "idx_started_at" ON "incidents" ("started_at");
-- Create index "idx_status_impact" to table: "incidents"
CREATE INDEX "idx_status_impact" ON "incidents" ("status", "impact");
-- Create index "idx_tenant_visible" to table: "incidents"
CREATE INDEX "idx_tenant_visible" ON "incidents" ("tenant_id", "is_visible");
-- Create "incident_components" table
CREATE TABLE "incident_components" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "incident_id" bigint NOT NULL,
  "component_id" bigint NOT NULL,
  "component_name" text NOT NULL,
  "status" text NOT NULL,
  "created_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_incidents_components" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_incident_components_component_id" to table: "incident_components"
CREATE INDEX "idx_incident_components_component_id" ON "incident_components" ("component_id");
-- Create index "idx_incident_components_deleted_at" to table: "incident_components"
CREATE INDEX "idx_incident_components_deleted_at" ON "incident_components" ("deleted_at");
-- Create index "idx_incident_components_incident_id" to table: "incident_components"
CREATE INDEX "idx_incident_components_incident_id" ON "incident_components" ("incident_id");
-- Create "incident_metrics" table
CREATE TABLE "incident_metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "incident_id" bigint NOT NULL,
  "metric_type" text NOT NULL,
  "value" numeric NOT NULL,
  "unit" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_incident_metrics_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_incident_metrics_deleted_at" to table: "incident_metrics"
CREATE INDEX "idx_incident_metrics_deleted_at" ON "incident_metrics" ("deleted_at");
-- Create index "idx_incident_metrics_incident_id" to table: "incident_metrics"
CREATE INDEX "idx_incident_metrics_incident_id" ON "incident_metrics" ("incident_id");
-- Create index "idx_incident_metrics_timestamp" to table: "incident_metrics"
CREATE INDEX "idx_incident_metrics_timestamp" ON "incident_metrics" ("timestamp");
-- Create "incident_notifications" table
CREATE TABLE "incident_notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "incident_id" bigint NOT NULL,
  "type" text NOT NULL,
  "recipient" text NOT NULL,
  "subject" text NULL,
  "message" text NULL,
  "status" text NULL DEFAULT 'pending',
  "sent_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_incident_notifications_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_incident_notifications_deleted_at" to table: "incident_notifications"
CREATE INDEX "idx_incident_notifications_deleted_at" ON "incident_notifications" ("deleted_at");
-- Create index "idx_incident_notifications_incident_id" to table: "incident_notifications"
CREATE INDEX "idx_incident_notifications_incident_id" ON "incident_notifications" ("incident_id");
-- Create "incident_updates" table
CREATE TABLE "incident_updates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "incident_id" bigint NOT NULL,
  "status" text NOT NULL,
  "message" text NOT NULL,
  "is_visible" boolean NULL DEFAULT true,
  "created_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_incidents_updates" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_incident_updates_deleted_at" to table: "incident_updates"
CREATE INDEX "idx_incident_updates_deleted_at" ON "incident_updates" ("deleted_at");
-- Create index "idx_incident_updates_incident_id" to table: "incident_updates"
CREATE INDEX "idx_incident_updates_incident_id" ON "incident_updates" ("incident_id");
-- Create "advanced_incident_templates" table
CREATE TABLE "advanced_incident_templates" (
  "id" bigserial NOT NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "title" character varying(500) NOT NULL,
  "body" text NULL,
  "severity" character varying(50) NOT NULL DEFAULT 'medium',
  "status" character varying(50) NOT NULL DEFAULT 'investigating',
  "is_public" boolean NULL DEFAULT true,
  "is_active" boolean NULL DEFAULT true,
  "created_by" bigint NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "component_ids" text NULL,
  "notification_settings" text NULL,
  "automation_settings" text NULL,
  "usage_count" bigint NULL DEFAULT 0,
  "last_used_at" timestamptz NULL,
  "last_used_by" bigint NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_advanced_incident_templates_tenant_id" to table: "advanced_incident_templates"
CREATE INDEX "idx_advanced_incident_templates_tenant_id" ON "advanced_incident_templates" ("tenant_id");
-- Create "workflow_steps" table
CREATE TABLE "workflow_steps" (
  "id" bigserial NOT NULL,
  "template_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "step_type" character varying(50) NOT NULL,
  "order" bigint NOT NULL,
  "trigger_type" character varying(50) NOT NULL DEFAULT 'manual',
  "trigger_delay" bigint NULL DEFAULT 0,
  "trigger_status" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "configuration" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_advanced_incident_templates_workflow_steps" FOREIGN KEY ("template_id") REFERENCES "advanced_incident_templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_workflow_steps_template_id" to table: "workflow_steps"
CREATE INDEX "idx_workflow_steps_template_id" ON "workflow_steps" ("template_id");
-- Create "step_executions" table
CREATE TABLE "step_executions" (
  "id" bigserial NOT NULL,
  "step_id" bigint NOT NULL,
  "incident_id" bigint NOT NULL,
  "tenant_id" bigint NOT NULL,
  "status" character varying(50) NOT NULL,
  "started_at" timestamptz NULL,
  "completed_at" timestamptz NULL,
  "executed_by" bigint NULL,
  "execution_type" character varying(50) NOT NULL,
  "result" text NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_step_executions_incident" FOREIGN KEY ("incident_id") REFERENCES "incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_workflow_steps_execution_logs" FOREIGN KEY ("step_id") REFERENCES "workflow_steps" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_step_executions_incident_id" to table: "step_executions"
CREATE INDEX "idx_step_executions_incident_id" ON "step_executions" ("incident_id");
-- Create index "idx_step_executions_step_id" to table: "step_executions"
CREATE INDEX "idx_step_executions_step_id" ON "step_executions" ("step_id");
-- Create index "idx_step_executions_tenant_id" to table: "step_executions"
CREATE INDEX "idx_step_executions_tenant_id" ON "step_executions" ("tenant_id");
