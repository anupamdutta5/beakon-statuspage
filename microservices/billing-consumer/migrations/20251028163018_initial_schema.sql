-- Create "billing_records" table
CREATE TABLE "billing_records" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "type" text NOT NULL,
  "subscription_id" text NULL,
  "payment_id" text NULL,
  "invoice_id" text NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL,
  "description" text NULL,
  "status" text NOT NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_billing_records_deleted_at" to table: "billing_records"
CREATE INDEX "idx_billing_records_deleted_at" ON "billing_records" ("deleted_at");
-- Create index "idx_billing_records_invoice_id" to table: "billing_records"
CREATE INDEX "idx_billing_records_invoice_id" ON "billing_records" ("invoice_id");
-- Create index "idx_billing_records_payment_id" to table: "billing_records"
CREATE INDEX "idx_billing_records_payment_id" ON "billing_records" ("payment_id");
-- Create index "idx_billing_records_status" to table: "billing_records"
CREATE INDEX "idx_billing_records_status" ON "billing_records" ("status");
-- Create index "idx_billing_records_subscription_id" to table: "billing_records"
CREATE INDEX "idx_billing_records_subscription_id" ON "billing_records" ("subscription_id");
-- Create index "idx_billing_records_tenant_id" to table: "billing_records"
CREATE INDEX "idx_billing_records_tenant_id" ON "billing_records" ("tenant_id");
-- Create index "idx_billing_records_timestamp" to table: "billing_records"
CREATE INDEX "idx_billing_records_timestamp" ON "billing_records" ("timestamp");
-- Create index "idx_billing_records_type" to table: "billing_records"
CREATE INDEX "idx_billing_records_type" ON "billing_records" ("type");
-- Create index "idx_billing_records_user_id" to table: "billing_records"
CREATE INDEX "idx_billing_records_user_id" ON "billing_records" ("user_id");
-- Create "invoice_records" table
CREATE TABLE "invoice_records" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "invoice_id" text NOT NULL,
  "subscription_id" text NULL,
  "amount" numeric NOT NULL,
  "tax_amount" numeric NULL,
  "total_amount" numeric NOT NULL,
  "currency" text NOT NULL,
  "status" text NOT NULL,
  "due_date" timestamptz NULL,
  "paid_at" timestamptz NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_invoice_records_deleted_at" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_deleted_at" ON "invoice_records" ("deleted_at");
-- Create index "idx_invoice_records_due_date" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_due_date" ON "invoice_records" ("due_date");
-- Create index "idx_invoice_records_invoice_id" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_invoice_id" ON "invoice_records" ("invoice_id");
-- Create index "idx_invoice_records_status" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_status" ON "invoice_records" ("status");
-- Create index "idx_invoice_records_subscription_id" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_subscription_id" ON "invoice_records" ("subscription_id");
-- Create index "idx_invoice_records_tenant_id" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_tenant_id" ON "invoice_records" ("tenant_id");
-- Create index "idx_invoice_records_timestamp" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_timestamp" ON "invoice_records" ("timestamp");
-- Create index "idx_invoice_records_user_id" to table: "invoice_records"
CREATE INDEX "idx_invoice_records_user_id" ON "invoice_records" ("user_id");
-- Create "payment_records" table
CREATE TABLE "payment_records" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NULL,
  "payment_id" text NOT NULL,
  "invoice_id" text NULL,
  "subscription_id" text NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL,
  "payment_method" text NOT NULL,
  "status" text NOT NULL,
  "transaction_id" text NULL,
  "gateway_response" text NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_payment_records_deleted_at" to table: "payment_records"
CREATE INDEX "idx_payment_records_deleted_at" ON "payment_records" ("deleted_at");
-- Create index "idx_payment_records_invoice_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_invoice_id" ON "payment_records" ("invoice_id");
-- Create index "idx_payment_records_payment_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_payment_id" ON "payment_records" ("payment_id");
-- Create index "idx_payment_records_payment_method" to table: "payment_records"
CREATE INDEX "idx_payment_records_payment_method" ON "payment_records" ("payment_method");
-- Create index "idx_payment_records_status" to table: "payment_records"
CREATE INDEX "idx_payment_records_status" ON "payment_records" ("status");
-- Create index "idx_payment_records_subscription_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_subscription_id" ON "payment_records" ("subscription_id");
-- Create index "idx_payment_records_tenant_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_tenant_id" ON "payment_records" ("tenant_id");
-- Create index "idx_payment_records_timestamp" to table: "payment_records"
CREATE INDEX "idx_payment_records_timestamp" ON "payment_records" ("timestamp");
-- Create index "idx_payment_records_transaction_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_transaction_id" ON "payment_records" ("transaction_id");
-- Create index "idx_payment_records_user_id" to table: "payment_records"
CREATE INDEX "idx_payment_records_user_id" ON "payment_records" ("user_id");
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
