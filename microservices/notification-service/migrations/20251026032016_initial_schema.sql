-- Create "channels" table
CREATE TABLE "channels" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "provider" text NULL,
  "config" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_default" boolean NULL DEFAULT false,
  "priority" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_channels_deleted_at" to table: "channels"
CREATE INDEX "idx_channels_deleted_at" ON "channels" ("deleted_at");
-- Create index "idx_channels_tenant_id" to table: "channels"
CREATE INDEX "idx_channels_tenant_id" ON "channels" ("tenant_id");
-- Create "webhook_events" table
CREATE TABLE "webhook_events" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "event_type" text NOT NULL,
  "event_data" text NULL,
  "source" text NULL,
  "status" text NULL DEFAULT 'pending',
  "processed_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "max_retries" bigint NULL DEFAULT 3,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_webhook_events_deleted_at" to table: "webhook_events"
CREATE INDEX "idx_webhook_events_deleted_at" ON "webhook_events" ("deleted_at");
-- Create index "idx_webhook_events_event_type" to table: "webhook_events"
CREATE INDEX "idx_webhook_events_event_type" ON "webhook_events" ("event_type");
-- Create index "idx_webhook_events_tenant_id" to table: "webhook_events"
CREATE INDEX "idx_webhook_events_tenant_id" ON "webhook_events" ("tenant_id");
-- Create "templates" table
CREATE TABLE "templates" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "type" text NOT NULL,
  "category" text NULL,
  "subject" text NULL,
  "content" text NULL,
  "variables" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_templates_deleted_at" to table: "templates"
CREATE INDEX "idx_templates_deleted_at" ON "templates" ("deleted_at");
-- Create index "idx_templates_tenant_id" to table: "templates"
CREATE INDEX "idx_templates_tenant_id" ON "templates" ("tenant_id");
-- Create "notifications" table
CREATE TABLE "notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "template_id" bigint NULL,
  "type" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "priority" text NULL DEFAULT 'normal',
  "subject" text NULL,
  "content" text NULL,
  "recipients" text NULL,
  "metadata" text NULL,
  "scheduled_at" timestamptz NULL,
  "sent_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "max_retries" bigint NULL DEFAULT 3,
  "error" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_templates_notifications" FOREIGN KEY ("template_id") REFERENCES "templates" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notifications_deleted_at" to table: "notifications"
CREATE INDEX "idx_notifications_deleted_at" ON "notifications" ("deleted_at");
-- Create index "idx_notifications_template_id" to table: "notifications"
CREATE INDEX "idx_notifications_template_id" ON "notifications" ("template_id");
-- Create index "idx_notifications_tenant_id" to table: "notifications"
CREATE INDEX "idx_notifications_tenant_id" ON "notifications" ("tenant_id");
-- Create index "idx_notifications_user_id" to table: "notifications"
CREATE INDEX "idx_notifications_user_id" ON "notifications" ("user_id");
-- Create "deliveries" table
CREATE TABLE "deliveries" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "notification_id" bigint NOT NULL,
  "channel_id" bigint NOT NULL,
  "recipient" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "attempt" bigint NULL DEFAULT 1,
  "max_attempts" bigint NULL DEFAULT 3,
  "sent_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "response_code" bigint NULL,
  "response_message" text NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_deliveries_channel" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_notifications_deliveries" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_deliveries_channel_id" to table: "deliveries"
CREATE INDEX "idx_deliveries_channel_id" ON "deliveries" ("channel_id");
-- Create index "idx_deliveries_deleted_at" to table: "deliveries"
CREATE INDEX "idx_deliveries_deleted_at" ON "deliveries" ("deleted_at");
-- Create index "idx_deliveries_notification_id" to table: "deliveries"
CREATE INDEX "idx_deliveries_notification_id" ON "deliveries" ("notification_id");
-- Create "notification_logs" table
CREATE TABLE "notification_logs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "notification_id" bigint NULL,
  "level" text NOT NULL,
  "message" text NOT NULL,
  "details" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_notification_logs_notification" FOREIGN KEY ("notification_id") REFERENCES "notifications" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_notification_logs_deleted_at" to table: "notification_logs"
CREATE INDEX "idx_notification_logs_deleted_at" ON "notification_logs" ("deleted_at");
-- Create index "idx_notification_logs_notification_id" to table: "notification_logs"
CREATE INDEX "idx_notification_logs_notification_id" ON "notification_logs" ("notification_id");
-- Create index "idx_notification_logs_tenant_id" to table: "notification_logs"
CREATE INDEX "idx_notification_logs_tenant_id" ON "notification_logs" ("tenant_id");
-- Create "subscriptions" table
CREATE TABLE "subscriptions" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "channel_id" bigint NOT NULL,
  "event_types" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "preferences" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_subscriptions_channel" FOREIGN KEY ("channel_id") REFERENCES "channels" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_subscriptions_channel_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_channel_id" ON "subscriptions" ("channel_id");
-- Create index "idx_subscriptions_deleted_at" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_deleted_at" ON "subscriptions" ("deleted_at");
-- Create index "idx_subscriptions_tenant_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_tenant_id" ON "subscriptions" ("tenant_id");
-- Create index "idx_subscriptions_user_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_user_id" ON "subscriptions" ("user_id");
