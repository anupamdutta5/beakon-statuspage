-- Create "component_groups" table
CREATE TABLE "component_groups" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "position" bigint NULL DEFAULT 0,
  "is_visible" boolean NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_component_groups_deleted_at" to table: "component_groups"
CREATE INDEX "idx_component_groups_deleted_at" ON "component_groups" ("deleted_at");
-- Create index "idx_component_groups_tenant_id" to table: "component_groups"
CREATE INDEX "idx_component_groups_tenant_id" ON "component_groups" ("tenant_id");
-- Create "components" table
CREATE TABLE "components" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "status" text NULL DEFAULT 'operational',
  "position" bigint NULL DEFAULT 0,
  "is_visible" boolean NULL DEFAULT true,
  "group_id" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_groups_components" FOREIGN KEY ("group_id") REFERENCES "component_groups" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_components_deleted_at" to table: "components"
CREATE INDEX "idx_components_deleted_at" ON "components" ("deleted_at");
-- Create index "idx_components_group_id" to table: "components"
CREATE INDEX "idx_components_group_id" ON "components" ("group_id");
-- Create index "idx_components_tenant_id" to table: "components"
CREATE INDEX "idx_components_tenant_id" ON "components" ("tenant_id");
-- Create "component_alerts" table
CREATE TABLE "component_alerts" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "alert_type" text NOT NULL,
  "status" text NOT NULL,
  "message" text NULL,
  "threshold" numeric NULL,
  "current_value" numeric NULL,
  "resolved_at" timestamptz NULL,
  "resolved_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_alerts_component" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_alerts_component_id" to table: "component_alerts"
CREATE INDEX "idx_component_alerts_component_id" ON "component_alerts" ("component_id");
-- Create index "idx_component_alerts_deleted_at" to table: "component_alerts"
CREATE INDEX "idx_component_alerts_deleted_at" ON "component_alerts" ("deleted_at");
-- Create "component_history" table
CREATE TABLE "component_history" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "old_status" text NULL,
  "new_status" text NOT NULL,
  "message" text NULL,
  "updated_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_history_component" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_history_component_id" to table: "component_history"
CREATE INDEX "idx_component_history_component_id" ON "component_history" ("component_id");
-- Create index "idx_component_history_deleted_at" to table: "component_history"
CREATE INDEX "idx_component_history_deleted_at" ON "component_history" ("deleted_at");
-- Create "component_metrics" table
CREATE TABLE "component_metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "metric_type" text NOT NULL,
  "value" numeric NOT NULL,
  "unit" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_metrics_component" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_metrics_component_id" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_component_id" ON "component_metrics" ("component_id");
-- Create index "idx_component_metrics_deleted_at" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_deleted_at" ON "component_metrics" ("deleted_at");
-- Create index "idx_component_metrics_timestamp" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_timestamp" ON "component_metrics" ("timestamp");
-- Create "component_statuses" table
CREATE TABLE "component_statuses" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "status" text NOT NULL,
  "message" text NULL,
  "updated_by" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_statuses_component" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_statuses_component_id" to table: "component_statuses"
CREATE INDEX "idx_component_statuses_component_id" ON "component_statuses" ("component_id");
-- Create index "idx_component_statuses_deleted_at" to table: "component_statuses"
CREATE INDEX "idx_component_statuses_deleted_at" ON "component_statuses" ("deleted_at");
-- Create "component_webhooks" table
CREATE TABLE "component_webhooks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "url" text NOT NULL,
  "events" text NULL,
  "secret" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_triggered" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_component_webhooks_component" FOREIGN KEY ("component_id") REFERENCES "components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_webhooks_component_id" to table: "component_webhooks"
CREATE INDEX "idx_component_webhooks_component_id" ON "component_webhooks" ("component_id");
-- Create index "idx_component_webhooks_deleted_at" to table: "component_webhooks"
CREATE INDEX "idx_component_webhooks_deleted_at" ON "component_webhooks" ("deleted_at");
