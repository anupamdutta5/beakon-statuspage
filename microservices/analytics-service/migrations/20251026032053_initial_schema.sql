-- Create "sla_targets" table
CREATE TABLE "sla_targets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "service_tier" text NOT NULL,
  "target_value" numeric NOT NULL,
  "unit" text NOT NULL,
  "is_default" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "settings" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_sla_targets_deleted_at" to table: "sla_targets"
CREATE INDEX "idx_sla_targets_deleted_at" ON "sla_targets" ("deleted_at");
-- Create index "idx_sla_targets_tenant_id" to table: "sla_targets"
CREATE INDEX "idx_sla_targets_tenant_id" ON "sla_targets" ("tenant_id");
-- Create "dashboards" table
CREATE TABLE "dashboards" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "is_public" boolean NULL DEFAULT false,
  "is_default" boolean NULL DEFAULT false,
  "layout" text NULL,
  "settings" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_dashboards_deleted_at" to table: "dashboards"
CREATE INDEX "idx_dashboards_deleted_at" ON "dashboards" ("deleted_at");
-- Create index "idx_dashboards_tenant_id" to table: "dashboards"
CREATE INDEX "idx_dashboards_tenant_id" ON "dashboards" ("tenant_id");
-- Create index "idx_dashboards_user_id" to table: "dashboards"
CREATE INDEX "idx_dashboards_user_id" ON "dashboards" ("user_id");
-- Create "data_exports" table
CREATE TABLE "data_exports" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "type" text NOT NULL,
  "format" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "filters" text NULL,
  "date_range" text NULL,
  "file_size" bigint NULL,
  "file_path" text NULL,
  "download_url" text NULL,
  "expires_at" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_data_exports_deleted_at" to table: "data_exports"
CREATE INDEX "idx_data_exports_deleted_at" ON "data_exports" ("deleted_at");
-- Create index "idx_data_exports_tenant_id" to table: "data_exports"
CREATE INDEX "idx_data_exports_tenant_id" ON "data_exports" ("tenant_id");
-- Create index "idx_data_exports_user_id" to table: "data_exports"
CREATE INDEX "idx_data_exports_user_id" ON "data_exports" ("user_id");
-- Create "analytics_events" table
CREATE TABLE "analytics_events" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "event_type" text NOT NULL,
  "event_name" text NOT NULL,
  "properties" text NULL,
  "session_id" text NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "referrer" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_analytics_events_deleted_at" to table: "analytics_events"
CREATE INDEX "idx_analytics_events_deleted_at" ON "analytics_events" ("deleted_at");
-- Create index "idx_analytics_events_event_type" to table: "analytics_events"
CREATE INDEX "idx_analytics_events_event_type" ON "analytics_events" ("event_type");
-- Create index "idx_analytics_events_session_id" to table: "analytics_events"
CREATE INDEX "idx_analytics_events_session_id" ON "analytics_events" ("session_id");
-- Create index "idx_analytics_events_tenant_id" to table: "analytics_events"
CREATE INDEX "idx_analytics_events_tenant_id" ON "analytics_events" ("tenant_id");
-- Create index "idx_analytics_events_user_id" to table: "analytics_events"
CREATE INDEX "idx_analytics_events_user_id" ON "analytics_events" ("user_id");
-- Create "response_time_metrics" table
CREATE TABLE "response_time_metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "component_id" bigint NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "response_time" numeric NOT NULL,
  "status_code" bigint NULL,
  "is_successful" boolean NOT NULL,
  "endpoint" text NULL,
  "method" text NULL,
  "location" text NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_response_time_metrics_component_id" to table: "response_time_metrics"
CREATE INDEX "idx_response_time_metrics_component_id" ON "response_time_metrics" ("component_id");
-- Create index "idx_response_time_metrics_deleted_at" to table: "response_time_metrics"
CREATE INDEX "idx_response_time_metrics_deleted_at" ON "response_time_metrics" ("deleted_at");
-- Create index "idx_response_time_metrics_tenant_id" to table: "response_time_metrics"
CREATE INDEX "idx_response_time_metrics_tenant_id" ON "response_time_metrics" ("tenant_id");
-- Create index "idx_response_time_metrics_timestamp" to table: "response_time_metrics"
CREATE INDEX "idx_response_time_metrics_timestamp" ON "response_time_metrics" ("timestamp");
-- Create "uptime_calculations" table
CREATE TABLE "uptime_calculations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "component_id" bigint NOT NULL,
  "period_start" timestamptz NOT NULL,
  "period_end" timestamptz NOT NULL,
  "total_minutes" bigint NOT NULL,
  "uptime_minutes" bigint NOT NULL,
  "downtime_minutes" bigint NOT NULL,
  "uptime_percent" numeric NOT NULL,
  "incident_count" bigint NULL,
  "maintenance_minutes" bigint NULL,
  "calculation_type" text NOT NULL,
  "status" text NULL DEFAULT 'active',
  "details" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_uptime_calculations_component_id" to table: "uptime_calculations"
CREATE INDEX "idx_uptime_calculations_component_id" ON "uptime_calculations" ("component_id");
-- Create index "idx_uptime_calculations_deleted_at" to table: "uptime_calculations"
CREATE INDEX "idx_uptime_calculations_deleted_at" ON "uptime_calculations" ("deleted_at");
-- Create index "idx_uptime_calculations_period_end" to table: "uptime_calculations"
CREATE INDEX "idx_uptime_calculations_period_end" ON "uptime_calculations" ("period_end");
-- Create index "idx_uptime_calculations_period_start" to table: "uptime_calculations"
CREATE INDEX "idx_uptime_calculations_period_start" ON "uptime_calculations" ("period_start");
-- Create index "idx_uptime_calculations_tenant_id" to table: "uptime_calculations"
CREATE INDEX "idx_uptime_calculations_tenant_id" ON "uptime_calculations" ("tenant_id");
-- Create "dashboard_widgets" table
CREATE TABLE "dashboard_widgets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "dashboard_id" bigint NOT NULL,
  "type" text NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "position" bigint NULL DEFAULT 0,
  "size" text NULL DEFAULT 'medium',
  "config" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_dashboards_widgets" FOREIGN KEY ("dashboard_id") REFERENCES "dashboards" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_dashboard_widgets_dashboard_id" to table: "dashboard_widgets"
CREATE INDEX "idx_dashboard_widgets_dashboard_id" ON "dashboard_widgets" ("dashboard_id");
-- Create index "idx_dashboard_widgets_deleted_at" to table: "dashboard_widgets"
CREATE INDEX "idx_dashboard_widgets_deleted_at" ON "dashboard_widgets" ("deleted_at");
-- Create "metrics" table
CREATE TABLE "metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "unit" text NULL,
  "category" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_public" boolean NULL DEFAULT false,
  "aggregation_type" text NULL DEFAULT 'sum',
  "retention_days" bigint NULL DEFAULT 365,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_metrics_deleted_at" to table: "metrics"
CREATE INDEX "idx_metrics_deleted_at" ON "metrics" ("deleted_at");
-- Create index "idx_metrics_tenant_id" to table: "metrics"
CREATE INDEX "idx_metrics_tenant_id" ON "metrics" ("tenant_id");
-- Create "metric_data_points" table
CREATE TABLE "metric_data_points" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "metric_id" bigint NOT NULL,
  "value" numeric NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "labels" text NULL,
  "source" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_metrics_data_points" FOREIGN KEY ("metric_id") REFERENCES "metrics" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_metric_data_points_deleted_at" to table: "metric_data_points"
CREATE INDEX "idx_metric_data_points_deleted_at" ON "metric_data_points" ("deleted_at");
-- Create index "idx_metric_data_points_metric_id" to table: "metric_data_points"
CREATE INDEX "idx_metric_data_points_metric_id" ON "metric_data_points" ("metric_id");
-- Create index "idx_metric_data_points_timestamp" to table: "metric_data_points"
CREATE INDEX "idx_metric_data_points_timestamp" ON "metric_data_points" ("timestamp");
-- Create "reports" table
CREATE TABLE "reports" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "status" text NULL DEFAULT 'draft',
  "schedule" text NULL,
  "format" text NULL DEFAULT 'pdf',
  "is_public" boolean NULL DEFAULT false,
  "last_generated" timestamptz NULL,
  "next_generation" timestamptz NULL,
  "config" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_reports_deleted_at" to table: "reports"
CREATE INDEX "idx_reports_deleted_at" ON "reports" ("deleted_at");
-- Create index "idx_reports_tenant_id" to table: "reports"
CREATE INDEX "idx_reports_tenant_id" ON "reports" ("tenant_id");
-- Create index "idx_reports_user_id" to table: "reports"
CREATE INDEX "idx_reports_user_id" ON "reports" ("user_id");
-- Create "report_generations" table
CREATE TABLE "report_generations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "report_id" bigint NOT NULL,
  "status" text NULL DEFAULT 'generating',
  "started_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "file_size" bigint NULL,
  "file_path" text NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_reports_generations" FOREIGN KEY ("report_id") REFERENCES "reports" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_report_generations_deleted_at" to table: "report_generations"
CREATE INDEX "idx_report_generations_deleted_at" ON "report_generations" ("deleted_at");
-- Create index "idx_report_generations_report_id" to table: "report_generations"
CREATE INDEX "idx_report_generations_report_id" ON "report_generations" ("report_id");
-- Create "slas" table
CREATE TABLE "slas" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "component_id" bigint NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "target_value" numeric NOT NULL,
  "unit" text NOT NULL,
  "period_type" text NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "alert_threshold" numeric NULL,
  "settings" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_slas_component_id" to table: "slas"
CREATE INDEX "idx_slas_component_id" ON "slas" ("component_id");
-- Create index "idx_slas_deleted_at" to table: "slas"
CREATE INDEX "idx_slas_deleted_at" ON "slas" ("deleted_at");
-- Create index "idx_slas_tenant_id" to table: "slas"
CREATE INDEX "idx_slas_tenant_id" ON "slas" ("tenant_id");
-- Create "sla_measurements" table
CREATE TABLE "sla_measurements" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "sla_id" bigint NOT NULL,
  "period_start" timestamptz NOT NULL,
  "period_end" timestamptz NOT NULL,
  "actual_value" numeric NOT NULL,
  "target_value" numeric NOT NULL,
  "compliance_rate" numeric NOT NULL,
  "is_compliant" boolean NOT NULL,
  "total_samples" bigint NULL,
  "valid_samples" bigint NULL,
  "failed_samples" bigint NULL,
  "status" text NULL DEFAULT 'active',
  "details" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_slas_measurements" FOREIGN KEY ("sla_id") REFERENCES "slas" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sla_measurements_deleted_at" to table: "sla_measurements"
CREATE INDEX "idx_sla_measurements_deleted_at" ON "sla_measurements" ("deleted_at");
-- Create index "idx_sla_measurements_period_end" to table: "sla_measurements"
CREATE INDEX "idx_sla_measurements_period_end" ON "sla_measurements" ("period_end");
-- Create index "idx_sla_measurements_period_start" to table: "sla_measurements"
CREATE INDEX "idx_sla_measurements_period_start" ON "sla_measurements" ("period_start");
-- Create index "idx_sla_measurements_sla_id" to table: "sla_measurements"
CREATE INDEX "idx_sla_measurements_sla_id" ON "sla_measurements" ("sla_id");
-- Create "sla_breaches" table
CREATE TABLE "sla_breaches" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "sla_id" bigint NOT NULL,
  "measurement_id" bigint NOT NULL,
  "severity" text NOT NULL,
  "status" text NULL DEFAULT 'open',
  "triggered_at" timestamptz NOT NULL,
  "detected_at" timestamptz NOT NULL,
  "resolved_at" timestamptz NULL,
  "duration" bigint NULL,
  "impact_value" numeric NULL,
  "description" text NULL,
  "root_cause" text NULL,
  "resolution" text NULL,
  "notified_users" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_sla_breaches_measurement" FOREIGN KEY ("measurement_id") REFERENCES "sla_measurements" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_slas_breaches" FOREIGN KEY ("sla_id") REFERENCES "slas" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sla_breaches_deleted_at" to table: "sla_breaches"
CREATE INDEX "idx_sla_breaches_deleted_at" ON "sla_breaches" ("deleted_at");
-- Create index "idx_sla_breaches_measurement_id" to table: "sla_breaches"
CREATE INDEX "idx_sla_breaches_measurement_id" ON "sla_breaches" ("measurement_id");
-- Create index "idx_sla_breaches_sla_id" to table: "sla_breaches"
CREATE INDEX "idx_sla_breaches_sla_id" ON "sla_breaches" ("sla_id");
-- Create "sla_reports" table
CREATE TABLE "sla_reports" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "sla_id" bigint NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "period" text NOT NULL,
  "period_start" timestamptz NOT NULL,
  "period_end" timestamptz NOT NULL,
  "status" text NULL DEFAULT 'generating',
  "format" text NULL DEFAULT 'pdf',
  "generated_at" timestamptz NULL,
  "file_path" text NULL,
  "file_size" bigint NULL,
  "download_url" text NULL,
  "expires_at" timestamptz NULL,
  "summary" text NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_slas_reports" FOREIGN KEY ("sla_id") REFERENCES "slas" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_sla_reports_deleted_at" to table: "sla_reports"
CREATE INDEX "idx_sla_reports_deleted_at" ON "sla_reports" ("deleted_at");
-- Create index "idx_sla_reports_sla_id" to table: "sla_reports"
CREATE INDEX "idx_sla_reports_sla_id" ON "sla_reports" ("sla_id");
-- Create index "idx_sla_reports_tenant_id" to table: "sla_reports"
CREATE INDEX "idx_sla_reports_tenant_id" ON "sla_reports" ("tenant_id");
