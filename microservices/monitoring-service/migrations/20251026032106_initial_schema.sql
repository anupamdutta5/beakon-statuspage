-- Create "maintenance_templates" table
CREATE TABLE "maintenance_templates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "title" character varying(255) NOT NULL,
  "message" text NULL,
  "type" character varying(50) NOT NULL,
  "impact" character varying(50) NOT NULL,
  "duration" bigint NULL DEFAULT 60,
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_maintenance_templates_deleted_at" to table: "maintenance_templates"
CREATE INDEX "idx_maintenance_templates_deleted_at" ON "maintenance_templates" ("deleted_at");
-- Create index "idx_maintenance_templates_tenant_id" to table: "maintenance_templates"
CREATE INDEX "idx_maintenance_templates_tenant_id" ON "maintenance_templates" ("tenant_id");
-- Create "on_call_schedules" table
CREATE TABLE "on_call_schedules" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "rotation_type" character varying(20) NOT NULL,
  "rotation_start" timestamptz NOT NULL,
  "rotation_interval_hours" bigint NULL DEFAULT 168,
  "participants" jsonb NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_on_call_active" to table: "on_call_schedules"
CREATE INDEX "idx_on_call_active" ON "on_call_schedules" ("is_active");
-- Create index "idx_on_call_tenant" to table: "on_call_schedules"
CREATE INDEX "idx_on_call_tenant" ON "on_call_schedules" ("tenant_id");
-- Create "heartbeat_monitors" table
CREATE TABLE "heartbeat_monitors" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "unique_key" character varying(255) NOT NULL,
  "expected_interval_seconds" bigint NOT NULL,
  "grace_period_seconds" bigint NULL DEFAULT 300,
  "last_ping" timestamptz NULL,
  "is_alive" boolean NULL DEFAULT false,
  "consecutive_misses" bigint NULL DEFAULT 0,
  "alert_sent" boolean NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "idx_heartbeat_status" to table: "heartbeat_monitors"
CREATE INDEX "idx_heartbeat_status" ON "heartbeat_monitors" ("is_alive");
-- Create index "idx_heartbeat_tenant" to table: "heartbeat_monitors"
CREATE INDEX "idx_heartbeat_tenant" ON "heartbeat_monitors" ("tenant_id");
-- Create "auto_incidents" table
CREATE TABLE "auto_incidents" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "monitor_id" bigint NOT NULL,
  "incident_id" uuid NOT NULL,
  "component_id" uuid NOT NULL,
  "triggered_by_result_id" bigint NULL,
  "failure_count" bigint NOT NULL,
  "resolved" boolean NULL DEFAULT false,
  "resolved_at" timestamptz NULL,
  "resolved_by_result_id" bigint NULL,
  "auto_resolved" boolean NULL DEFAULT false,
  "error_message" text NULL,
  "affected_locations" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_auto_incidents_incident" to table: "auto_incidents"
CREATE INDEX "idx_auto_incidents_incident" ON "auto_incidents" ("incident_id");
-- Create index "idx_auto_incidents_monitor" to table: "auto_incidents"
CREATE INDEX "idx_auto_incidents_monitor" ON "auto_incidents" ("monitor_id");
-- Create index "idx_auto_incidents_tenant" to table: "auto_incidents"
CREATE INDEX "idx_auto_incidents_tenant" ON "auto_incidents" ("tenant_id");
-- Create index "idx_auto_incidents_unresolved" to table: "auto_incidents"
CREATE INDEX "idx_auto_incidents_unresolved" ON "auto_incidents" ("resolved");
-- Create "monitored_services" table
CREATE TABLE "monitored_services" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "url" text NULL,
  "host" text NULL,
  "port" bigint NULL,
  "status" text NULL DEFAULT 'unknown',
  "is_active" boolean NULL DEFAULT true,
  "check_interval" bigint NULL DEFAULT 60,
  "timeout" bigint NULL DEFAULT 30,
  "retries" bigint NULL DEFAULT 3,
  "last_checked" timestamptz NULL,
  "last_healthy" timestamptz NULL,
  "last_unhealthy" timestamptz NULL,
  "uptime" numeric NULL DEFAULT 0,
  "response_time" numeric NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_active_status" to table: "monitored_services"
CREATE INDEX "idx_active_status" ON "monitored_services" ("is_active");
-- Create index "idx_monitored_services_deleted_at" to table: "monitored_services"
CREATE INDEX "idx_monitored_services_deleted_at" ON "monitored_services" ("deleted_at");
-- Create index "idx_tenant_status" to table: "monitored_services"
CREATE INDEX "idx_tenant_status" ON "monitored_services" ("tenant_id", "status");
-- Create index "idx_type_status" to table: "monitored_services"
CREATE INDEX "idx_type_status" ON "monitored_services" ("type", "status");
-- Create "monitoring_results_daily" table
CREATE TABLE "monitoring_results_daily" (
  "monitor_id" bigint NOT NULL,
  "day" date NOT NULL,
  "avg_response_time_ms" bigint NULL,
  "p95_response_time_ms" bigint NULL,
  "p99_response_time_ms" bigint NULL,
  "uptime_percentage" numeric(5,2) NULL,
  "total_checks" bigint NULL,
  "total_failures" bigint NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("monitor_id", "day")
);
-- Create index "idx_monitoring_daily_day" to table: "monitoring_results_daily"
CREATE INDEX "idx_monitoring_daily_day" ON "monitoring_results_daily" ("day");
-- Create "monitor_notifications" table
CREATE TABLE "monitor_notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "monitor_id" bigint NOT NULL,
  "notify_email" boolean NULL DEFAULT true,
  "notify_sms" boolean NULL DEFAULT false,
  "notify_slack" boolean NULL DEFAULT false,
  "notify_webhook" boolean NULL DEFAULT false,
  "notify_on_failure" boolean NULL DEFAULT true,
  "notify_on_recovery" boolean NULL DEFAULT true,
  "notify_on_degraded" boolean NULL DEFAULT false,
  "escalation_policy_id" bigint NULL,
  "on_call_schedule_id" bigint NULL,
  "custom_email_recipients" text NULL,
  "custom_webhook_url" character varying(1000) NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_monitor_notifications_monitor" to table: "monitor_notifications"
CREATE INDEX "idx_monitor_notifications_monitor" ON "monitor_notifications" ("monitor_id");
-- Create "ssl_certificates" table
CREATE TABLE "ssl_certificates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "domain" character varying(255) NOT NULL,
  "issuer" character varying(255) NULL,
  "subject" character varying(255) NULL,
  "serial_number" character varying(255) NULL,
  "valid_from" timestamptz NULL,
  "valid_until" timestamptz NULL,
  "days_until_expiry" bigint NULL,
  "last_checked" timestamptz NULL,
  "is_valid" boolean NULL DEFAULT true,
  "is_self_signed" boolean NULL DEFAULT false,
  "warning_sent_30d" boolean NULL DEFAULT false,
  "warning_sent_14d" boolean NULL DEFAULT false,
  "warning_sent_7d" boolean NULL DEFAULT false,
  "error_message" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_ssl_tenant" to table: "ssl_certificates"
CREATE INDEX "idx_ssl_tenant" ON "ssl_certificates" ("tenant_id", "is_valid");
-- Create "anomaly_baselines" table
CREATE TABLE "anomaly_baselines" (
  "id" bigserial NOT NULL,
  "tenant_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "metric_type" character varying(50) NOT NULL,
  "baseline_type" character varying(20) NOT NULL,
  "hour_of_day" bigint NULL,
  "day_of_week" bigint NULL,
  "mean_value" numeric NOT NULL,
  "std_dev" numeric NOT NULL,
  "min_value" numeric NULL,
  "max_value" numeric NULL,
  "p50" numeric NULL,
  "p95" numeric NULL,
  "p99" numeric NULL,
  "sample_count" bigint NOT NULL,
  "calculated_at" timestamptz NULL,
  "expires_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_anomaly_baselines_expires_at" to table: "anomaly_baselines"
CREATE INDEX "idx_anomaly_baselines_expires_at" ON "anomaly_baselines" ("expires_at");
-- Create index "idx_anomaly_baselines_metric_type" to table: "anomaly_baselines"
CREATE INDEX "idx_anomaly_baselines_metric_type" ON "anomaly_baselines" ("metric_type");
-- Create index "idx_anomaly_baselines_monitor_id" to table: "anomaly_baselines"
CREATE INDEX "idx_anomaly_baselines_monitor_id" ON "anomaly_baselines" ("monitor_id");
-- Create index "idx_anomaly_baselines_tenant_id" to table: "anomaly_baselines"
CREATE INDEX "idx_anomaly_baselines_tenant_id" ON "anomaly_baselines" ("tenant_id");
-- Create "detected_anomalies" table
CREATE TABLE "detected_anomalies" (
  "id" bigserial NOT NULL,
  "tenant_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "metric_type" character varying(50) NOT NULL,
  "detected_at" timestamptz NOT NULL,
  "actual_value" numeric NOT NULL,
  "expected_value" numeric NOT NULL,
  "deviation_score" numeric NOT NULL,
  "severity" character varying(20) NOT NULL,
  "detection_method" character varying(50) NULL DEFAULT 'z_score',
  "status" character varying(20) NOT NULL DEFAULT 'open',
  "acknowledged_at" timestamptz NULL,
  "acknowledged_by" bigint NULL,
  "resolved_at" timestamptz NULL,
  "resolution_notes" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_detected_anomalies_detected_at" to table: "detected_anomalies"
CREATE INDEX "idx_detected_anomalies_detected_at" ON "detected_anomalies" ("detected_at" DESC);
-- Create index "idx_detected_anomalies_monitor_id" to table: "detected_anomalies"
CREATE INDEX "idx_detected_anomalies_monitor_id" ON "detected_anomalies" ("monitor_id");
-- Create index "idx_detected_anomalies_severity" to table: "detected_anomalies"
CREATE INDEX "idx_detected_anomalies_severity" ON "detected_anomalies" ("severity");
-- Create index "idx_detected_anomalies_status" to table: "detected_anomalies"
CREATE INDEX "idx_detected_anomalies_status" ON "detected_anomalies" ("status");
-- Create index "idx_detected_anomalies_tenant_id" to table: "detected_anomalies"
CREATE INDEX "idx_detected_anomalies_tenant_id" ON "detected_anomalies" ("tenant_id");
-- Create "monitor_status_history" table
CREATE TABLE "monitor_status_history" (
  "id" bigserial NOT NULL,
  "changed_at" timestamptz NOT NULL,
  "monitor_id" bigint NOT NULL,
  "tenant_id" uuid NOT NULL,
  "previous_status" character varying(20) NULL,
  "new_status" character varying(20) NOT NULL,
  "duration_seconds" bigint NULL,
  "triggered_by" character varying(50) NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_status_history_monitor" to table: "monitor_status_history"
CREATE INDEX "idx_status_history_monitor" ON "monitor_status_history" ("monitor_id", "changed_at");
-- Create index "idx_status_history_tenant" to table: "monitor_status_history"
CREATE INDEX "idx_status_history_tenant" ON "monitor_status_history" ("tenant_id", "changed_at");
-- Create "monitors" table
CREATE TABLE "monitors" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "component_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "monitor_type" character varying(50) NOT NULL,
  "check_url" character varying(1000) NULL,
  "check_interval_seconds" bigint NOT NULL DEFAULT 60,
  "timeout_seconds" bigint NOT NULL DEFAULT 30,
  "http_method" character varying(10) NULL DEFAULT 'GET',
  "http_headers" text NULL,
  "http_body" text NULL,
  "expected_status_codes" text NULL DEFAULT '200,201,204',
  "follow_redirects" boolean NULL DEFAULT true,
  "verify_ssl" boolean NULL DEFAULT true,
  "enabled_locations" text NULL,
  "auto_create_incidents" boolean NULL DEFAULT true,
  "failure_threshold" bigint NULL DEFAULT 3,
  "consecutive_failures" bigint NULL DEFAULT 0,
  "last_incident_id" uuid NULL,
  "is_active" boolean NULL DEFAULT true,
  "current_status" character varying(20) NULL DEFAULT 'unknown',
  "last_check_at" timestamptz NULL,
  "last_success_at" timestamptz NULL,
  "last_failure_at" timestamptz NULL,
  "uptime_percentage" numeric(5,2) NULL DEFAULT 100,
  "in_maintenance" boolean NULL DEFAULT false,
  "maintenance_until" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_monitors_active" to table: "monitors"
CREATE INDEX "idx_monitors_active" ON "monitors" ("is_active");
-- Create index "idx_monitors_component" to table: "monitors"
CREATE INDEX "idx_monitors_component" ON "monitors" ("component_id");
-- Create index "idx_monitors_deleted_at" to table: "monitors"
CREATE INDEX "idx_monitors_deleted_at" ON "monitors" ("deleted_at");
-- Create index "idx_monitors_failures" to table: "monitors"
CREATE INDEX "idx_monitors_failures" ON "monitors" ("consecutive_failures");
-- Create index "idx_monitors_status" to table: "monitors"
CREATE INDEX "idx_monitors_status" ON "monitors" ("current_status");
-- Create index "idx_monitors_tenant" to table: "monitors"
CREATE INDEX "idx_monitors_tenant" ON "monitors" ("tenant_id");
-- Create "metric_snapshots" table
CREATE TABLE "metric_snapshots" (
  "id" bigserial NOT NULL,
  "tenant_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "metric_type" character varying(50) NOT NULL,
  "metric_value" numeric NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_metric_snapshots_metric_type" to table: "metric_snapshots"
CREATE INDEX "idx_metric_snapshots_metric_type" ON "metric_snapshots" ("metric_type");
-- Create index "idx_metric_snapshots_monitor_id" to table: "metric_snapshots"
CREATE INDEX "idx_metric_snapshots_monitor_id" ON "metric_snapshots" ("monitor_id");
-- Create index "idx_metric_snapshots_tenant_id" to table: "metric_snapshots"
CREATE INDEX "idx_metric_snapshots_tenant_id" ON "metric_snapshots" ("tenant_id");
-- Create index "idx_metric_snapshots_timestamp" to table: "metric_snapshots"
CREATE INDEX "idx_metric_snapshots_timestamp" ON "metric_snapshots" ("timestamp" DESC);
-- Create "alerts" table
CREATE TABLE "alerts" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "service_id" bigint NULL,
  "type" text NOT NULL,
  "severity" text NOT NULL,
  "status" text NULL DEFAULT 'active',
  "title" text NOT NULL,
  "description" text NULL,
  "message" text NULL,
  "triggered_at" timestamptz NOT NULL,
  "acknowledged_at" timestamptz NULL,
  "acknowledged_by" bigint NULL,
  "resolved_at" timestamptz NULL,
  "resolved_by" bigint NULL,
  "resolution_type" text NULL,
  "resolution_note" text NULL,
  "dedup_key" text NULL,
  "error_type" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitored_services_alerts" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_alerts_deleted_at" to table: "alerts"
CREATE INDEX "idx_alerts_deleted_at" ON "alerts" ("deleted_at");
-- Create index "idx_alerts_service_id" to table: "alerts"
CREATE INDEX "idx_alerts_service_id" ON "alerts" ("service_id");
-- Create index "idx_alerts_tenant_id" to table: "alerts"
CREATE INDEX "idx_alerts_tenant_id" ON "alerts" ("tenant_id");
-- Create "escalation_policies" table
CREATE TABLE "escalation_policies" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "levels" jsonb NOT NULL,
  "is_default" boolean NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "idx_escalation_default" to table: "escalation_policies"
CREATE INDEX "idx_escalation_default" ON "escalation_policies" ("is_default");
-- Create index "idx_escalation_tenant" to table: "escalation_policies"
CREATE INDEX "idx_escalation_tenant" ON "escalation_policies" ("tenant_id");
-- Create "anomaly_detection_config" table
CREATE TABLE "anomaly_detection_config" (
  "id" bigserial NOT NULL,
  "tenant_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "metric_type" character varying(50) NULL,
  "enabled" boolean NULL DEFAULT true,
  "sensitivity" character varying(20) NULL DEFAULT 'medium',
  "min_baseline_samples" bigint NULL DEFAULT 50,
  "z_score_threshold_minor" numeric NULL DEFAULT 2,
  "z_score_threshold_major" numeric NULL DEFAULT 3,
  "z_score_threshold_critical" numeric NULL DEFAULT 4,
  "notification_enabled" boolean NULL DEFAULT true,
  "notification_cooldown_minutes" bigint NULL DEFAULT 30,
  "require_consecutive_anomalies" bigint NULL DEFAULT 1,
  "escalation_policy_id" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_anomaly_detection_config_escalation_policy" FOREIGN KEY ("escalation_policy_id") REFERENCES "escalation_policies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_anomaly_detection_config_enabled" to table: "anomaly_detection_config"
CREATE INDEX "idx_anomaly_detection_config_enabled" ON "anomaly_detection_config" ("enabled");
-- Create index "idx_anomaly_detection_config_monitor_id" to table: "anomaly_detection_config"
CREATE INDEX "idx_anomaly_detection_config_monitor_id" ON "anomaly_detection_config" ("monitor_id");
-- Create index "idx_anomaly_detection_config_tenant_id" to table: "anomaly_detection_config"
CREATE INDEX "idx_anomaly_detection_config_tenant_id" ON "anomaly_detection_config" ("tenant_id");
-- Create "integrations" table
CREATE TABLE "integrations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "description" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "configuration" text NULL,
  "last_sync_at" timestamptz NULL,
  "sync_status" text NULL DEFAULT 'pending',
  "sync_error" text NULL,
  "sync_interval" bigint NULL DEFAULT 300,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_integrations_deleted_at" to table: "integrations"
CREATE INDEX "idx_integrations_deleted_at" ON "integrations" ("deleted_at");
-- Create index "idx_integrations_tenant_id" to table: "integrations"
CREATE INDEX "idx_integrations_tenant_id" ON "integrations" ("tenant_id");
-- Create "component_mappings" table
CREATE TABLE "component_mappings" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "integration_id" bigint NOT NULL,
  "component_id" bigint NOT NULL,
  "external_service_id" text NOT NULL,
  "external_service_name" text NOT NULL,
  "mapping_config" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_status_sync" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_integrations_component_mappings" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_component_mappings_component_id" to table: "component_mappings"
CREATE INDEX "idx_component_mappings_component_id" ON "component_mappings" ("component_id");
-- Create index "idx_component_mappings_deleted_at" to table: "component_mappings"
CREATE INDEX "idx_component_mappings_deleted_at" ON "component_mappings" ("deleted_at");
-- Create index "idx_component_mappings_integration_id" to table: "component_mappings"
CREATE INDEX "idx_component_mappings_integration_id" ON "component_mappings" ("integration_id");
-- Create "monitored_components" table
CREATE TABLE "monitored_components" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "type" character varying(50) NOT NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_monitored_components_deleted_at" to table: "monitored_components"
CREATE INDEX "idx_monitored_components_deleted_at" ON "monitored_components" ("deleted_at");
-- Create index "idx_monitored_components_tenant_id" to table: "monitored_components"
CREATE INDEX "idx_monitored_components_tenant_id" ON "monitored_components" ("tenant_id");
-- Create "component_metrics" table
CREATE TABLE "component_metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "value" numeric NOT NULL,
  "unit" character varying(50) NULL,
  "timestamp" timestamptz NOT NULL,
  "labels" text NULL,
  "source" character varying(100) NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitored_components_metrics" FOREIGN KEY ("component_id") REFERENCES "monitored_components" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_component_metrics_component_id" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_component_id" ON "component_metrics" ("component_id");
-- Create index "idx_component_metrics_deleted_at" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_deleted_at" ON "component_metrics" ("deleted_at");
-- Create index "idx_component_metrics_timestamp" to table: "component_metrics"
CREATE INDEX "idx_component_metrics_timestamp" ON "component_metrics" ("timestamp");
-- Create "monitored_containers" table
CREATE TABLE "monitored_containers" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "component_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "type" character varying(50) NOT NULL,
  "location" character varying(100) NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitored_components_containers" FOREIGN KEY ("component_id") REFERENCES "monitored_components" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_monitored_containers_component_id" to table: "monitored_containers"
CREATE INDEX "idx_monitored_containers_component_id" ON "monitored_containers" ("component_id");
-- Create index "idx_monitored_containers_deleted_at" to table: "monitored_containers"
CREATE INDEX "idx_monitored_containers_deleted_at" ON "monitored_containers" ("deleted_at");
-- Create "container_health_checks" table
CREATE TABLE "container_health_checks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "container_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "type" character varying(50) NOT NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "last_checked" timestamptz NULL,
  "message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitored_containers_health_checks" FOREIGN KEY ("container_id") REFERENCES "monitored_containers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_container_health_checks_container_id" to table: "container_health_checks"
CREATE INDEX "idx_container_health_checks_container_id" ON "container_health_checks" ("container_id");
-- Create index "idx_container_health_checks_deleted_at" to table: "container_health_checks"
CREATE INDEX "idx_container_health_checks_deleted_at" ON "container_health_checks" ("deleted_at");
-- Create "custom_metrics" table
CREATE TABLE "custom_metrics" (
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
  "retention_days" bigint NULL DEFAULT 30,
  "labels" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_custom_metrics_deleted_at" to table: "custom_metrics"
CREATE INDEX "idx_custom_metrics_deleted_at" ON "custom_metrics" ("deleted_at");
-- Create index "idx_custom_metrics_tenant_id" to table: "custom_metrics"
CREATE INDEX "idx_custom_metrics_tenant_id" ON "custom_metrics" ("tenant_id");
-- Create index "idx_tenant_name" to table: "custom_metrics"
CREATE UNIQUE INDEX "idx_tenant_name" ON "custom_metrics" ("name");
-- Create "custom_metric_data_points" table
CREATE TABLE "custom_metric_data_points" (
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
  CONSTRAINT "fk_custom_metrics_data_points" FOREIGN KEY ("metric_id") REFERENCES "custom_metrics" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_custom_metric_data_points_deleted_at" to table: "custom_metric_data_points"
CREATE INDEX "idx_custom_metric_data_points_deleted_at" ON "custom_metric_data_points" ("deleted_at");
-- Create index "idx_custom_metric_data_points_metric_id" to table: "custom_metric_data_points"
CREATE INDEX "idx_custom_metric_data_points_metric_id" ON "custom_metric_data_points" ("metric_id");
-- Create index "idx_custom_metric_data_points_timestamp" to table: "custom_metric_data_points"
CREATE INDEX "idx_custom_metric_data_points_timestamp" ON "custom_metric_data_points" ("timestamp");
-- Create "discord_integrations" table
CREATE TABLE "discord_integrations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "webhook_url" text NOT NULL,
  "webhook_name" text NULL,
  "avatar_url" text NULL,
  "default_channel_id" text NULL,
  "default_channel_name" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_used_at" timestamptz NULL,
  "notify_on_down" boolean NULL DEFAULT true,
  "notify_on_up" boolean NULL DEFAULT true,
  "notify_on_degraded" boolean NULL DEFAULT true,
  "notify_on_maintenance" boolean NULL DEFAULT false,
  "mention_users" text[] NULL,
  "mention_roles" text[] NULL,
  "mention_everyone" boolean NULL DEFAULT false,
  "custom_color" text NULL,
  "include_monitor_url" boolean NULL DEFAULT true,
  "include_timestamp" boolean NULL DEFAULT true,
  "retry_count" bigint NULL DEFAULT 3,
  "retry_interval_seconds" bigint NULL DEFAULT 5,
  PRIMARY KEY ("id")
);
-- Create index "idx_discord_integrations_deleted_at" to table: "discord_integrations"
CREATE INDEX "idx_discord_integrations_deleted_at" ON "discord_integrations" ("deleted_at");
-- Create index "idx_discord_integrations_tenant_id" to table: "discord_integrations"
CREATE INDEX "idx_discord_integrations_tenant_id" ON "discord_integrations" ("tenant_id");
-- Create "discord_channel_subscriptions" table
CREATE TABLE "discord_channel_subscriptions" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "integration_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "channel_id" text NULL,
  "channel_name" text NULL,
  "notify_on_down" boolean NULL DEFAULT true,
  "notify_on_up" boolean NULL DEFAULT true,
  "notify_on_degraded" boolean NULL DEFAULT true,
  "notify_on_maintenance" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discord_integrations_channel_subscriptions" FOREIGN KEY ("integration_id") REFERENCES "discord_integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_discord_channel_subscriptions_deleted_at" to table: "discord_channel_subscriptions"
CREATE INDEX "idx_discord_channel_subscriptions_deleted_at" ON "discord_channel_subscriptions" ("deleted_at");
-- Create index "idx_discord_channel_subscriptions_integration_id" to table: "discord_channel_subscriptions"
CREATE INDEX "idx_discord_channel_subscriptions_integration_id" ON "discord_channel_subscriptions" ("integration_id");
-- Create index "idx_discord_channel_subscriptions_monitor_id" to table: "discord_channel_subscriptions"
CREATE INDEX "idx_discord_channel_subscriptions_monitor_id" ON "discord_channel_subscriptions" ("monitor_id");
-- Create "discord_notifications" table
CREATE TABLE "discord_notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "integration_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "event_type" character varying(50) NOT NULL,
  "monitor_name" text NULL,
  "monitor_url" text NULL,
  "status" character varying(50) NOT NULL,
  "http_status_code" bigint NULL,
  "message_content" text NULL,
  "embed_data" jsonb NULL,
  "sent_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "error_message" text NULL,
  "error_code" text NULL,
  "discord_message_id" text NULL,
  "discord_channel_id" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_discord_integrations_notifications" FOREIGN KEY ("integration_id") REFERENCES "discord_integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_discord_notifications_deleted_at" to table: "discord_notifications"
CREATE INDEX "idx_discord_notifications_deleted_at" ON "discord_notifications" ("deleted_at");
-- Create index "idx_discord_notifications_event_type" to table: "discord_notifications"
CREATE INDEX "idx_discord_notifications_event_type" ON "discord_notifications" ("event_type");
-- Create index "idx_discord_notifications_integration_id" to table: "discord_notifications"
CREATE INDEX "idx_discord_notifications_integration_id" ON "discord_notifications" ("integration_id");
-- Create index "idx_discord_notifications_monitor_id" to table: "discord_notifications"
CREATE INDEX "idx_discord_notifications_monitor_id" ON "discord_notifications" ("monitor_id");
-- Create index "idx_discord_notifications_status" to table: "discord_notifications"
CREATE INDEX "idx_discord_notifications_status" ON "discord_notifications" ("status");
-- Create "docker_containers" table
CREATE TABLE "docker_containers" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "image" character varying(500) NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "state" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_checked" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_docker_containers_deleted_at" to table: "docker_containers"
CREATE INDEX "idx_docker_containers_deleted_at" ON "docker_containers" ("deleted_at");
-- Create index "idx_docker_containers_tenant_id" to table: "docker_containers"
CREATE INDEX "idx_docker_containers_tenant_id" ON "docker_containers" ("tenant_id");
-- Create "docker_container_health_checks" table
CREATE TABLE "docker_container_health_checks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "container_id" bigint NOT NULL,
  "type" character varying(50) NOT NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "last_checked" timestamptz NULL,
  "message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_docker_containers_health_checks" FOREIGN KEY ("container_id") REFERENCES "docker_containers" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_docker_container_health_checks_container_id" to table: "docker_container_health_checks"
CREATE INDEX "idx_docker_container_health_checks_container_id" ON "docker_container_health_checks" ("container_id");
-- Create index "idx_docker_container_health_checks_deleted_at" to table: "docker_container_health_checks"
CREATE INDEX "idx_docker_container_health_checks_deleted_at" ON "docker_container_health_checks" ("deleted_at");
-- Create "external_services" table
CREATE TABLE "external_services" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "type" character varying(50) NOT NULL,
  "provider" character varying(100) NULL,
  "url" character varying(500) NULL,
  "api_key" character varying(500) NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "is_active" boolean NULL DEFAULT true,
  "check_interval" bigint NULL DEFAULT 300,
  "last_checked" timestamptz NULL,
  "last_healthy" timestamptz NULL,
  "last_unhealthy" timestamptz NULL,
  "uptime" numeric NULL DEFAULT 0,
  "response_time" numeric NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_external_services_deleted_at" to table: "external_services"
CREATE INDEX "idx_external_services_deleted_at" ON "external_services" ("deleted_at");
-- Create index "idx_external_services_tenant_id" to table: "external_services"
CREATE INDEX "idx_external_services_tenant_id" ON "external_services" ("tenant_id");
-- Create "external_service_health_checks" table
CREATE TABLE "external_service_health_checks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "external_service_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "type" character varying(50) NOT NULL,
  "url" character varying(500) NULL,
  "method" character varying(10) NULL DEFAULT 'GET',
  "headers" text NULL,
  "body" text NULL,
  "expected_status" bigint NULL DEFAULT 200,
  "expected_body" text NULL,
  "timeout" bigint NULL DEFAULT 30,
  "is_active" boolean NULL DEFAULT true,
  "last_checked" timestamptz NULL,
  "last_result" character varying(50) NULL DEFAULT 'unknown',
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_external_services_health_checks" FOREIGN KEY ("external_service_id") REFERENCES "external_services" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_external_service_health_checks_deleted_at" to table: "external_service_health_checks"
CREATE INDEX "idx_external_service_health_checks_deleted_at" ON "external_service_health_checks" ("deleted_at");
-- Create index "idx_external_service_health_checks_external_service_id" to table: "external_service_health_checks"
CREATE INDEX "idx_external_service_health_checks_external_service_id" ON "external_service_health_checks" ("external_service_id");
-- Create "health_checks" table
CREATE TABLE "health_checks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "service_id" bigint NOT NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "url" text NULL,
  "method" text NULL DEFAULT 'GET',
  "headers" text NULL,
  "body" text NULL,
  "expected_status" bigint NULL DEFAULT 200,
  "expected_body" text NULL,
  "timeout" bigint NULL DEFAULT 30,
  "is_active" boolean NULL DEFAULT true,
  "last_checked" timestamptz NULL,
  "last_result" text NULL DEFAULT 'unknown',
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitored_services_health_checks" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_health_checks_deleted_at" to table: "health_checks"
CREATE INDEX "idx_health_checks_deleted_at" ON "health_checks" ("deleted_at");
-- Create index "idx_last_checked" to table: "health_checks"
CREATE INDEX "idx_last_checked" ON "health_checks" ("last_checked");
-- Create index "idx_result_checked" to table: "health_checks"
CREATE INDEX "idx_result_checked" ON "health_checks" ("last_result");
-- Create index "idx_service_active" to table: "health_checks"
CREATE INDEX "idx_service_active" ON "health_checks" ("service_id", "is_active");
-- Create "health_check_results" table
CREATE TABLE "health_check_results" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "service_id" bigint NOT NULL,
  "health_check_id" bigint NOT NULL,
  "status" text NOT NULL,
  "response_time" numeric NULL,
  "status_code" bigint NULL,
  "response_body" text NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_health_check_results_health_check" FOREIGN KEY ("health_check_id") REFERENCES "health_checks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_health_check_results_service" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_health_check_results_deleted_at" to table: "health_check_results"
CREATE INDEX "idx_health_check_results_deleted_at" ON "health_check_results" ("deleted_at");
-- Create index "idx_health_check_results_health_check_id" to table: "health_check_results"
CREATE INDEX "idx_health_check_results_health_check_id" ON "health_check_results" ("health_check_id");
-- Create index "idx_health_check_results_service_id" to table: "health_check_results"
CREATE INDEX "idx_health_check_results_service_id" ON "health_check_results" ("service_id");
-- Create "integration_sync_logs" table
CREATE TABLE "integration_sync_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "integration_id" bigint NOT NULL,
  "sync_type" text NOT NULL,
  "status" text NOT NULL,
  "started_at" timestamptz NOT NULL,
  "completed_at" timestamptz NULL,
  "duration" bigint NULL,
  "records_sync" bigint NULL,
  "error_message" text NULL,
  "details" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_integrations_sync_logs" FOREIGN KEY ("integration_id") REFERENCES "integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_integration_sync_logs_deleted_at" to table: "integration_sync_logs"
CREATE INDEX "idx_integration_sync_logs_deleted_at" ON "integration_sync_logs" ("deleted_at");
-- Create index "idx_integration_sync_logs_integration_id" to table: "integration_sync_logs"
CREATE INDEX "idx_integration_sync_logs_integration_id" ON "integration_sync_logs" ("integration_id");
-- Create "kubernetes_resources" table
CREATE TABLE "kubernetes_resources" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "namespace" character varying(100) NULL,
  "type" character varying(50) NOT NULL,
  "kind" character varying(50) NOT NULL,
  "status" character varying(50) NULL DEFAULT 'unknown',
  "is_active" boolean NULL DEFAULT true,
  "last_checked" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_kubernetes_resources_deleted_at" to table: "kubernetes_resources"
CREATE INDEX "idx_kubernetes_resources_deleted_at" ON "kubernetes_resources" ("deleted_at");
-- Create index "idx_kubernetes_resources_tenant_id" to table: "kubernetes_resources"
CREATE INDEX "idx_kubernetes_resources_tenant_id" ON "kubernetes_resources" ("tenant_id");
-- Create "kubernetes_events" table
CREATE TABLE "kubernetes_events" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "resource_id" bigint NOT NULL,
  "type" character varying(50) NOT NULL,
  "reason" character varying(100) NULL,
  "message" text NULL,
  "count" bigint NULL DEFAULT 1,
  "first_seen" timestamptz NOT NULL,
  "last_seen" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_kubernetes_resources_events" FOREIGN KEY ("resource_id") REFERENCES "kubernetes_resources" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_kubernetes_events_deleted_at" to table: "kubernetes_events"
CREATE INDEX "idx_kubernetes_events_deleted_at" ON "kubernetes_events" ("deleted_at");
-- Create index "idx_kubernetes_events_resource_id" to table: "kubernetes_events"
CREATE INDEX "idx_kubernetes_events_resource_id" ON "kubernetes_events" ("resource_id");
-- Create "log_entries" table
CREATE TABLE "log_entries" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "service_id" bigint NULL,
  "level" text NOT NULL,
  "message" text NOT NULL,
  "source" text NULL,
  "timestamp" timestamptz NOT NULL,
  "fields" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_log_entries_service" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_log_entries_deleted_at" to table: "log_entries"
CREATE INDEX "idx_log_entries_deleted_at" ON "log_entries" ("deleted_at");
-- Create index "idx_log_entries_level" to table: "log_entries"
CREATE INDEX "idx_log_entries_level" ON "log_entries" ("level");
-- Create index "idx_log_entries_service_id" to table: "log_entries"
CREATE INDEX "idx_log_entries_service_id" ON "log_entries" ("service_id");
-- Create index "idx_log_entries_tenant_id" to table: "log_entries"
CREATE INDEX "idx_log_entries_tenant_id" ON "log_entries" ("tenant_id");
-- Create index "idx_log_entries_timestamp" to table: "log_entries"
CREATE INDEX "idx_log_entries_timestamp" ON "log_entries" ("timestamp");
-- Create "maintenance_windows" table
CREATE TABLE "maintenance_windows" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "tenant_id" uuid NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "starts_at" timestamptz NOT NULL,
  "ends_at" timestamptz NOT NULL,
  "status" character varying(50) NULL DEFAULT 'scheduled',
  "is_active" boolean NULL DEFAULT true,
  "created_by" uuid NULL,
  "reminder_sent" boolean NULL DEFAULT false,
  "auto_started" boolean NULL DEFAULT false,
  "auto_completed" boolean NULL DEFAULT false,
  "actual_start_time" timestamptz NULL,
  "actual_end_time" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_maintenance_windows_tenant_id" to table: "maintenance_windows"
CREATE INDEX "idx_maintenance_windows_tenant_id" ON "maintenance_windows" ("tenant_id");
-- Create "maintenance_components" table
CREATE TABLE "maintenance_components" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "maintenance_id" bigint NOT NULL,
  "component_id" bigint NOT NULL,
  "status" character varying(50) NULL DEFAULT 'operational',
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maintenance_components_component" FOREIGN KEY ("component_id") REFERENCES "monitored_components" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_maintenance_windows_components" FOREIGN KEY ("maintenance_id") REFERENCES "maintenance_windows" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_maintenance_components_component_id" to table: "maintenance_components"
CREATE INDEX "idx_maintenance_components_component_id" ON "maintenance_components" ("component_id");
-- Create index "idx_maintenance_components_deleted_at" to table: "maintenance_components"
CREATE INDEX "idx_maintenance_components_deleted_at" ON "maintenance_components" ("deleted_at");
-- Create index "idx_maintenance_components_maintenance_id" to table: "maintenance_components"
CREATE INDEX "idx_maintenance_components_maintenance_id" ON "maintenance_components" ("maintenance_id");
-- Create "maintenance_updates" table
CREATE TABLE "maintenance_updates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "maintenance_id" bigint NOT NULL,
  "status" character varying(50) NOT NULL,
  "message" text NULL,
  "is_public" boolean NULL DEFAULT true,
  "created_by" bigint NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_maintenance_windows_updates" FOREIGN KEY ("maintenance_id") REFERENCES "maintenance_windows" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_maintenance_updates_deleted_at" to table: "maintenance_updates"
CREATE INDEX "idx_maintenance_updates_deleted_at" ON "maintenance_updates" ("deleted_at");
-- Create index "idx_maintenance_updates_maintenance_id" to table: "maintenance_updates"
CREATE INDEX "idx_maintenance_updates_maintenance_id" ON "maintenance_updates" ("maintenance_id");
-- Create "monitoring_locations" table
CREATE TABLE "monitoring_locations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" character varying(100) NOT NULL,
  "city" character varying(100) NULL,
  "country" character varying(100) NOT NULL,
  "region" character varying(50) NOT NULL,
  "latitude" numeric(9,6) NULL,
  "longitude" numeric(9,6) NULL,
  "is_active" boolean NULL DEFAULT true,
  "provider" character varying(50) NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_monitoring_locations_deleted_at" to table: "monitoring_locations"
CREATE INDEX "idx_monitoring_locations_deleted_at" ON "monitoring_locations" ("deleted_at");
-- Create index "idx_monitoring_locations_is_active" to table: "monitoring_locations"
CREATE INDEX "idx_monitoring_locations_is_active" ON "monitoring_locations" ("is_active");
-- Create index "idx_monitoring_locations_region" to table: "monitoring_locations"
CREATE INDEX "idx_monitoring_locations_region" ON "monitoring_locations" ("region");
-- Create "monitoring_results" table
CREATE TABLE "monitoring_results" (
  "id" bigserial NOT NULL,
  "monitor_id" bigint NOT NULL,
  "location_id" bigint NULL,
  "checked_at" timestamptz NOT NULL,
  "status" character varying(20) NOT NULL,
  "response_time_ms" bigint NULL,
  "ttfb_ms" bigint NULL,
  "dns_time_ms" bigint NULL,
  "connection_time_ms" bigint NULL,
  "ssl_handshake_ms" bigint NULL,
  "status_code" bigint NULL,
  "error_message" text NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_monitoring_results_location" FOREIGN KEY ("location_id") REFERENCES "monitoring_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_monitoring_results_location" to table: "monitoring_results"
CREATE INDEX "idx_monitoring_results_location" ON "monitoring_results" ("location_id");
-- Create index "idx_monitoring_results_monitor_time" to table: "monitoring_results"
CREATE INDEX "idx_monitoring_results_monitor_time" ON "monitoring_results" ("monitor_id", "checked_at");
-- Create index "idx_monitoring_results_status" to table: "monitoring_results"
CREATE INDEX "idx_monitoring_results_status" ON "monitoring_results" ("status");
-- Create "monitoring_results_hourly" table
CREATE TABLE "monitoring_results_hourly" (
  "monitor_id" bigint NOT NULL,
  "location_id" bigint NOT NULL,
  "hour" timestamptz NOT NULL,
  "avg_response_time_ms" bigint NULL,
  "min_response_time_ms" bigint NULL,
  "max_response_time_ms" bigint NULL,
  "p50_response_time_ms" bigint NULL,
  "p95_response_time_ms" bigint NULL,
  "p99_response_time_ms" bigint NULL,
  "uptime_percentage" numeric(5,2) NULL,
  "check_count" bigint NULL,
  "failure_count" bigint NULL,
  "created_at" timestamptz NULL,
  PRIMARY KEY ("monitor_id", "location_id", "hour"),
  CONSTRAINT "fk_monitoring_results_hourly_location" FOREIGN KEY ("location_id") REFERENCES "monitoring_locations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_monitoring_hourly_time" to table: "monitoring_results_hourly"
CREATE INDEX "idx_monitoring_hourly_time" ON "monitoring_results_hourly" ("hour");
-- Create "performance_metrics" table
CREATE TABLE "performance_metrics" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "service_id" bigint NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "unit" text NULL,
  "category" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "retention_days" bigint NULL DEFAULT 30,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_performance_metrics_service" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_performance_metrics_deleted_at" to table: "performance_metrics"
CREATE INDEX "idx_performance_metrics_deleted_at" ON "performance_metrics" ("deleted_at");
-- Create index "idx_performance_metrics_service_id" to table: "performance_metrics"
CREATE INDEX "idx_performance_metrics_service_id" ON "performance_metrics" ("service_id");
-- Create index "idx_performance_metrics_tenant_id" to table: "performance_metrics"
CREATE INDEX "idx_performance_metrics_tenant_id" ON "performance_metrics" ("tenant_id");
-- Create "performance_data_points" table
CREATE TABLE "performance_data_points" (
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
  CONSTRAINT "fk_performance_metrics_data_points" FOREIGN KEY ("metric_id") REFERENCES "performance_metrics" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_performance_data_points_deleted_at" to table: "performance_data_points"
CREATE INDEX "idx_performance_data_points_deleted_at" ON "performance_data_points" ("deleted_at");
-- Create index "idx_performance_data_points_metric_id" to table: "performance_data_points"
CREATE INDEX "idx_performance_data_points_metric_id" ON "performance_data_points" ("metric_id");
-- Create index "idx_performance_data_points_timestamp" to table: "performance_data_points"
CREATE INDEX "idx_performance_data_points_timestamp" ON "performance_data_points" ("timestamp");
-- Create "status_automations" table
CREATE TABLE "status_automations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" character varying(255) NOT NULL,
  "description" text NULL,
  "type" character varying(50) NOT NULL,
  "provider" character varying(100) NULL,
  "config" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_sync" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_status_automations_deleted_at" to table: "status_automations"
CREATE INDEX "idx_status_automations_deleted_at" ON "status_automations" ("deleted_at");
-- Create index "idx_status_automations_tenant_id" to table: "status_automations"
CREATE INDEX "idx_status_automations_tenant_id" ON "status_automations" ("tenant_id");
-- Create "status_automation_integrations" table
CREATE TABLE "status_automation_integrations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "automation_id" bigint NOT NULL,
  "service_id" bigint NULL,
  "external_id" character varying(255) NULL,
  "config" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_sync" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_status_automation_integrations_service" FOREIGN KEY ("service_id") REFERENCES "monitored_services" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_status_automations_integrations" FOREIGN KEY ("automation_id") REFERENCES "status_automations" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);
-- Create index "idx_status_automation_integrations_automation_id" to table: "status_automation_integrations"
CREATE INDEX "idx_status_automation_integrations_automation_id" ON "status_automation_integrations" ("automation_id");
-- Create index "idx_status_automation_integrations_deleted_at" to table: "status_automation_integrations"
CREATE INDEX "idx_status_automation_integrations_deleted_at" ON "status_automation_integrations" ("deleted_at");
-- Create index "idx_status_automation_integrations_service_id" to table: "status_automation_integrations"
CREATE INDEX "idx_status_automation_integrations_service_id" ON "status_automation_integrations" ("service_id");
-- Create "telegram_integrations" table
CREATE TABLE "telegram_integrations" (
  "id" bigserial NOT NULL,
  "tenant_id" uuid NOT NULL,
  "bot_token" text NOT NULL,
  "bot_username" character varying(255) NULL,
  "bot_name" character varying(255) NULL,
  "default_chat_id" character varying(255) NULL,
  "default_chat_name" character varying(255) NULL,
  "default_chat_type" character varying(50) NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_used_at" timestamptz NULL,
  "notify_on_down" boolean NULL DEFAULT true,
  "notify_on_up" boolean NULL DEFAULT true,
  "notify_on_degraded" boolean NULL DEFAULT true,
  "notify_on_maintenance" boolean NULL DEFAULT false,
  "use_markdown" boolean NULL DEFAULT true,
  "include_monitor_url" boolean NULL DEFAULT true,
  "include_timestamp" boolean NULL DEFAULT true,
  "silent_notifications" boolean NULL DEFAULT false,
  "disable_preview" boolean NULL DEFAULT false,
  "retry_count" bigint NULL DEFAULT 3,
  "retry_interval_seconds" bigint NULL DEFAULT 5,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_telegram_integrations_deleted_at" to table: "telegram_integrations"
CREATE INDEX "idx_telegram_integrations_deleted_at" ON "telegram_integrations" ("deleted_at");
-- Create index "idx_telegram_integrations_tenant_id" to table: "telegram_integrations"
CREATE INDEX "idx_telegram_integrations_tenant_id" ON "telegram_integrations" ("tenant_id");
-- Create "telegram_chat_subscriptions" table
CREATE TABLE "telegram_chat_subscriptions" (
  "id" bigserial NOT NULL,
  "integration_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "chat_id" character varying(255) NULL,
  "chat_name" character varying(255) NULL,
  "chat_type" character varying(50) NULL,
  "notify_on_down" boolean NULL DEFAULT true,
  "notify_on_up" boolean NULL DEFAULT true,
  "notify_on_degraded" boolean NULL DEFAULT true,
  "notify_on_maintenance" boolean NULL DEFAULT false,
  "message_thread_id" bigint NULL,
  "custom_message_prefix" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_telegram_chat_subscriptions_integration" FOREIGN KEY ("integration_id") REFERENCES "telegram_integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_telegram_chat_subscriptions_integration_id" to table: "telegram_chat_subscriptions"
CREATE INDEX "idx_telegram_chat_subscriptions_integration_id" ON "telegram_chat_subscriptions" ("integration_id");
-- Create index "idx_telegram_chat_subscriptions_monitor_id" to table: "telegram_chat_subscriptions"
CREATE INDEX "idx_telegram_chat_subscriptions_monitor_id" ON "telegram_chat_subscriptions" ("monitor_id");
-- Create "telegram_notifications" table
CREATE TABLE "telegram_notifications" (
  "id" bigserial NOT NULL,
  "integration_id" bigint NOT NULL,
  "monitor_id" bigint NULL,
  "event_type" character varying(50) NOT NULL,
  "monitor_name" character varying(255) NULL,
  "monitor_url" text NULL,
  "status" character varying(50) NOT NULL,
  "http_status_code" bigint NULL,
  "message_text" text NULL,
  "parse_mode" character varying(50) NULL,
  "sent_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "error_message" text NULL,
  "error_code" character varying(50) NULL,
  "telegram_message_id" bigint NULL,
  "telegram_chat_id" character varying(255) NULL,
  "telegram_chat_type" character varying(50) NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_telegram_notifications_integration" FOREIGN KEY ("integration_id") REFERENCES "telegram_integrations" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_telegram_notifications_created_at" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_created_at" ON "telegram_notifications" ("created_at");
-- Create index "idx_telegram_notifications_event_type" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_event_type" ON "telegram_notifications" ("event_type");
-- Create index "idx_telegram_notifications_integration_id" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_integration_id" ON "telegram_notifications" ("integration_id");
-- Create index "idx_telegram_notifications_monitor_id" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_monitor_id" ON "telegram_notifications" ("monitor_id");
-- Create index "idx_telegram_notifications_status" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_status" ON "telegram_notifications" ("status");
-- Create index "idx_telegram_notifications_telegram_chat_id" to table: "telegram_notifications"
CREATE INDEX "idx_telegram_notifications_telegram_chat_id" ON "telegram_notifications" ("telegram_chat_id");
-- Create "uptime_checks" table
CREATE TABLE "uptime_checks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "url" text NOT NULL,
  "method" text NULL DEFAULT 'GET',
  "headers" text NULL,
  "body" text NULL,
  "expected_status" bigint NULL DEFAULT 200,
  "expected_body" text NULL,
  "check_interval" bigint NULL DEFAULT 60,
  "timeout" bigint NULL DEFAULT 30,
  "retries" bigint NULL DEFAULT 3,
  "is_active" boolean NULL DEFAULT true,
  "last_checked" timestamptz NULL,
  "last_success" timestamptz NULL,
  "last_failure" timestamptz NULL,
  "uptime" numeric NULL DEFAULT 0,
  "average_response_time" numeric NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_uptime_checks_deleted_at" to table: "uptime_checks"
CREATE INDEX "idx_uptime_checks_deleted_at" ON "uptime_checks" ("deleted_at");
-- Create index "idx_uptime_checks_tenant_id" to table: "uptime_checks"
CREATE INDEX "idx_uptime_checks_tenant_id" ON "uptime_checks" ("tenant_id");
-- Create "uptime_results" table
CREATE TABLE "uptime_results" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "uptime_check_id" bigint NOT NULL,
  "status" text NOT NULL,
  "response_time" numeric NULL,
  "status_code" bigint NULL,
  "response_body" text NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_uptime_checks_results" FOREIGN KEY ("uptime_check_id") REFERENCES "uptime_checks" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_uptime_results_deleted_at" to table: "uptime_results"
CREATE INDEX "idx_uptime_results_deleted_at" ON "uptime_results" ("deleted_at");
-- Create index "idx_uptime_results_uptime_check_id" to table: "uptime_results"
CREATE INDEX "idx_uptime_results_uptime_check_id" ON "uptime_results" ("uptime_check_id");
-- Create "webhook_endpoints" table
CREATE TABLE "webhook_endpoints" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "url" text NOT NULL,
  "secret_key" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "events" text NULL,
  "headers" text NULL,
  "timeout" bigint NULL DEFAULT 30,
  "retry_count" bigint NULL DEFAULT 3,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_webhook_endpoints_deleted_at" to table: "webhook_endpoints"
CREATE INDEX "idx_webhook_endpoints_deleted_at" ON "webhook_endpoints" ("deleted_at");
-- Create index "idx_webhook_endpoints_tenant_id" to table: "webhook_endpoints"
CREATE INDEX "idx_webhook_endpoints_tenant_id" ON "webhook_endpoints" ("tenant_id");
-- Create "webhook_deliveries" table
CREATE TABLE "webhook_deliveries" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "webhook_id" bigint NOT NULL,
  "event_type" text NOT NULL,
  "event_id" text NOT NULL,
  "payload" text NULL,
  "request_headers" text NULL,
  "response_status" bigint NULL,
  "response_headers" text NULL,
  "response_body" text NULL,
  "duration" bigint NULL,
  "success" boolean NULL,
  "error_message" text NULL,
  "attempt_count" bigint NULL DEFAULT 1,
  "next_retry_at" timestamptz NULL,
  "delivered_at" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_webhook_endpoints_deliveries" FOREIGN KEY ("webhook_id") REFERENCES "webhook_endpoints" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_webhook_deliveries_deleted_at" to table: "webhook_deliveries"
CREATE INDEX "idx_webhook_deliveries_deleted_at" ON "webhook_deliveries" ("deleted_at");
-- Create index "idx_webhook_deliveries_webhook_id" to table: "webhook_deliveries"
CREATE INDEX "idx_webhook_deliveries_webhook_id" ON "webhook_deliveries" ("webhook_id");
