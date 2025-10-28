-- Create "aggregated_data" table
CREATE TABLE "aggregated_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "metric_name" text NOT NULL,
  "aggregation_type" text NOT NULL,
  "value" numeric NOT NULL,
  "count" bigint NOT NULL,
  "period" text NOT NULL,
  "start_time" timestamptz NOT NULL,
  "end_time" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_aggregated_data_aggregation_type" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_aggregation_type" ON "aggregated_data" ("aggregation_type");
-- Create index "idx_aggregated_data_deleted_at" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_deleted_at" ON "aggregated_data" ("deleted_at");
-- Create index "idx_aggregated_data_end_time" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_end_time" ON "aggregated_data" ("end_time");
-- Create index "idx_aggregated_data_metric_name" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_metric_name" ON "aggregated_data" ("metric_name");
-- Create index "idx_aggregated_data_period" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_period" ON "aggregated_data" ("period");
-- Create index "idx_aggregated_data_start_time" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_start_time" ON "aggregated_data" ("start_time");
-- Create index "idx_aggregated_data_tenant_id" to table: "aggregated_data"
CREATE INDEX "idx_aggregated_data_tenant_id" ON "aggregated_data" ("tenant_id");
-- Create "error_data" table
CREATE TABLE "error_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "error_type" text NOT NULL,
  "error_message" text NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "page" text NULL,
  "stack" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_error_data_deleted_at" to table: "error_data"
CREATE INDEX "idx_error_data_deleted_at" ON "error_data" ("deleted_at");
-- Create index "idx_error_data_error_type" to table: "error_data"
CREATE INDEX "idx_error_data_error_type" ON "error_data" ("error_type");
-- Create index "idx_error_data_page" to table: "error_data"
CREATE INDEX "idx_error_data_page" ON "error_data" ("page");
-- Create index "idx_error_data_session_id" to table: "error_data"
CREATE INDEX "idx_error_data_session_id" ON "error_data" ("session_id");
-- Create index "idx_error_data_tenant_id" to table: "error_data"
CREATE INDEX "idx_error_data_tenant_id" ON "error_data" ("tenant_id");
-- Create index "idx_error_data_timestamp" to table: "error_data"
CREATE INDEX "idx_error_data_timestamp" ON "error_data" ("timestamp");
-- Create index "idx_error_data_user_id" to table: "error_data"
CREATE INDEX "idx_error_data_user_id" ON "error_data" ("user_id");
-- Create "metric_data" table
CREATE TABLE "metric_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "metric_name" text NOT NULL,
  "metric_value" numeric NOT NULL,
  "timestamp" timestamptz NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_metric_data_deleted_at" to table: "metric_data"
CREATE INDEX "idx_metric_data_deleted_at" ON "metric_data" ("deleted_at");
-- Create index "idx_metric_data_metric_name" to table: "metric_data"
CREATE INDEX "idx_metric_data_metric_name" ON "metric_data" ("metric_name");
-- Create index "idx_metric_data_session_id" to table: "metric_data"
CREATE INDEX "idx_metric_data_session_id" ON "metric_data" ("session_id");
-- Create index "idx_metric_data_tenant_id" to table: "metric_data"
CREATE INDEX "idx_metric_data_tenant_id" ON "metric_data" ("tenant_id");
-- Create index "idx_metric_data_timestamp" to table: "metric_data"
CREATE INDEX "idx_metric_data_timestamp" ON "metric_data" ("timestamp");
-- Create index "idx_metric_data_user_id" to table: "metric_data"
CREATE INDEX "idx_metric_data_user_id" ON "metric_data" ("user_id");
-- Create "page_view_data" table
CREATE TABLE "page_view_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "page" text NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "user_agent" text NULL,
  "ip_address" text NULL,
  "referrer" text NULL,
  "duration" bigint NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_page_view_data_deleted_at" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_deleted_at" ON "page_view_data" ("deleted_at");
-- Create index "idx_page_view_data_page" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_page" ON "page_view_data" ("page");
-- Create index "idx_page_view_data_session_id" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_session_id" ON "page_view_data" ("session_id");
-- Create index "idx_page_view_data_tenant_id" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_tenant_id" ON "page_view_data" ("tenant_id");
-- Create index "idx_page_view_data_timestamp" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_timestamp" ON "page_view_data" ("timestamp");
-- Create index "idx_page_view_data_user_id" to table: "page_view_data"
CREATE INDEX "idx_page_view_data_user_id" ON "page_view_data" ("user_id");
-- Create "performance_data" table
CREATE TABLE "performance_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "metric_name" text NOT NULL,
  "metric_value" numeric NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "page" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_performance_data_deleted_at" to table: "performance_data"
CREATE INDEX "idx_performance_data_deleted_at" ON "performance_data" ("deleted_at");
-- Create index "idx_performance_data_metric_name" to table: "performance_data"
CREATE INDEX "idx_performance_data_metric_name" ON "performance_data" ("metric_name");
-- Create index "idx_performance_data_page" to table: "performance_data"
CREATE INDEX "idx_performance_data_page" ON "performance_data" ("page");
-- Create index "idx_performance_data_session_id" to table: "performance_data"
CREATE INDEX "idx_performance_data_session_id" ON "performance_data" ("session_id");
-- Create index "idx_performance_data_tenant_id" to table: "performance_data"
CREATE INDEX "idx_performance_data_tenant_id" ON "performance_data" ("tenant_id");
-- Create index "idx_performance_data_timestamp" to table: "performance_data"
CREATE INDEX "idx_performance_data_timestamp" ON "performance_data" ("timestamp");
-- Create index "idx_performance_data_user_id" to table: "performance_data"
CREATE INDEX "idx_performance_data_user_id" ON "performance_data" ("user_id");
-- Create "processing_logs" table
CREATE TABLE "processing_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "message_id" text NOT NULL,
  "event_id" text NOT NULL,
  "event_type" text NOT NULL,
  "status" text NOT NULL,
  "processing_time" bigint NULL,
  "error" text NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_processing_logs_deleted_at" to table: "processing_logs"
CREATE INDEX "idx_processing_logs_deleted_at" ON "processing_logs" ("deleted_at");
-- Create index "idx_processing_logs_event_id" to table: "processing_logs"
CREATE INDEX "idx_processing_logs_event_id" ON "processing_logs" ("event_id");
-- Create index "idx_processing_logs_event_type" to table: "processing_logs"
CREATE INDEX "idx_processing_logs_event_type" ON "processing_logs" ("event_type");
-- Create index "idx_processing_logs_message_id" to table: "processing_logs"
CREATE INDEX "idx_processing_logs_message_id" ON "processing_logs" ("message_id");
-- Create "queue_messages" table
CREATE TABLE "queue_messages" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "queue_name" text NOT NULL,
  "event_type" text NOT NULL,
  "data" text NOT NULL,
  "priority" bigint NULL DEFAULT 0,
  "status" text NULL DEFAULT 'pending',
  "retry_count" bigint NULL DEFAULT 0,
  "max_retries" bigint NULL DEFAULT 3,
  "processed_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_queue_messages_deleted_at" to table: "queue_messages"
CREATE INDEX "idx_queue_messages_deleted_at" ON "queue_messages" ("deleted_at");
-- Create index "idx_queue_messages_event_type" to table: "queue_messages"
CREATE INDEX "idx_queue_messages_event_type" ON "queue_messages" ("event_type");
-- Create index "idx_queue_messages_queue_name" to table: "queue_messages"
CREATE INDEX "idx_queue_messages_queue_name" ON "queue_messages" ("queue_name");
-- Create "user_action_data" table
CREATE TABLE "user_action_data" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "action" text NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "page" text NULL,
  "element" text NULL,
  "value" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_action_data_action" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_action" ON "user_action_data" ("action");
-- Create index "idx_user_action_data_deleted_at" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_deleted_at" ON "user_action_data" ("deleted_at");
-- Create index "idx_user_action_data_page" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_page" ON "user_action_data" ("page");
-- Create index "idx_user_action_data_session_id" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_session_id" ON "user_action_data" ("session_id");
-- Create index "idx_user_action_data_tenant_id" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_tenant_id" ON "user_action_data" ("tenant_id");
-- Create index "idx_user_action_data_timestamp" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_timestamp" ON "user_action_data" ("timestamp");
-- Create index "idx_user_action_data_user_id" to table: "user_action_data"
CREATE INDEX "idx_user_action_data_user_id" ON "user_action_data" ("user_id");
