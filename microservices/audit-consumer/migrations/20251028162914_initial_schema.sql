-- Create "audit_logs" table
CREATE TABLE "audit_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "session_id" text NULL,
  "action" text NOT NULL,
  "resource" text NOT NULL,
  "resource_id" text NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "request_id" text NULL,
  "status" text NOT NULL,
  "error_message" text NULL,
  "changes" text NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_audit_logs_action" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_action" ON "audit_logs" ("action");
-- Create index "idx_audit_logs_deleted_at" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_deleted_at" ON "audit_logs" ("deleted_at");
-- Create index "idx_audit_logs_request_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_request_id" ON "audit_logs" ("request_id");
-- Create index "idx_audit_logs_resource" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_resource" ON "audit_logs" ("resource");
-- Create index "idx_audit_logs_resource_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_resource_id" ON "audit_logs" ("resource_id");
-- Create index "idx_audit_logs_session_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_session_id" ON "audit_logs" ("session_id");
-- Create index "idx_audit_logs_status" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_status" ON "audit_logs" ("status");
-- Create index "idx_audit_logs_tenant_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_tenant_id" ON "audit_logs" ("tenant_id");
-- Create index "idx_audit_logs_timestamp" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_timestamp" ON "audit_logs" ("timestamp");
-- Create index "idx_audit_logs_user_id" to table: "audit_logs"
CREATE INDEX "idx_audit_logs_user_id" ON "audit_logs" ("user_id");
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
-- Create "compliance_logs" table
CREATE TABLE "compliance_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "audit_log_id" bigint NOT NULL,
  "compliance_type" text NOT NULL,
  "rule" text NOT NULL,
  "status" text NOT NULL,
  "severity" text NOT NULL,
  "message" text NULL,
  "timestamp" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_compliance_logs_audit_log" FOREIGN KEY ("audit_log_id") REFERENCES "audit_logs" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_compliance_logs_audit_log_id" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_audit_log_id" ON "compliance_logs" ("audit_log_id");
-- Create index "idx_compliance_logs_compliance_type" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_compliance_type" ON "compliance_logs" ("compliance_type");
-- Create index "idx_compliance_logs_deleted_at" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_deleted_at" ON "compliance_logs" ("deleted_at");
-- Create index "idx_compliance_logs_rule" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_rule" ON "compliance_logs" ("rule");
-- Create index "idx_compliance_logs_severity" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_severity" ON "compliance_logs" ("severity");
-- Create index "idx_compliance_logs_status" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_status" ON "compliance_logs" ("status");
-- Create index "idx_compliance_logs_tenant_id" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_tenant_id" ON "compliance_logs" ("tenant_id");
-- Create index "idx_compliance_logs_timestamp" to table: "compliance_logs"
CREATE INDEX "idx_compliance_logs_timestamp" ON "compliance_logs" ("timestamp");
