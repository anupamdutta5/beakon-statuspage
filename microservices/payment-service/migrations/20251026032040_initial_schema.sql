-- Create "payment_webhooks" table
CREATE TABLE "payment_webhooks" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "gateway" text NOT NULL,
  "event_type" text NOT NULL,
  "gateway_id" text NULL,
  "payload" text NULL,
  "processed" boolean NULL DEFAULT false,
  "processed_at" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_payment_webhooks_deleted_at" to table: "payment_webhooks"
CREATE INDEX "idx_payment_webhooks_deleted_at" ON "payment_webhooks" ("deleted_at");
-- Create index "idx_payment_webhooks_gateway_id" to table: "payment_webhooks"
CREATE INDEX "idx_payment_webhooks_gateway_id" ON "payment_webhooks" ("gateway_id");
-- Create "plans" table
CREATE TABLE "plans" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "price" numeric NOT NULL,
  "currency" text NOT NULL DEFAULT 'USD',
  "billing_cycle" text NOT NULL,
  "is_active" boolean NULL DEFAULT true,
  "is_public" boolean NULL DEFAULT true,
  "features" text NULL,
  "limits" text NULL,
  "trial_days" bigint NULL DEFAULT 0,
  "sort_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_plans_deleted_at" to table: "plans"
CREATE INDEX "idx_plans_deleted_at" ON "plans" ("deleted_at");
-- Create index "idx_plans_tenant_id" to table: "plans"
CREATE INDEX "idx_plans_tenant_id" ON "plans" ("tenant_id");
-- Create "subscriptions" table
CREATE TABLE "subscriptions" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "plan_id" bigint NOT NULL,
  "status" text NULL DEFAULT 'active',
  "billing_cycle" text NOT NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL DEFAULT 'USD',
  "started_at" timestamptz NOT NULL,
  "current_period_start" timestamptz NOT NULL,
  "current_period_end" timestamptz NOT NULL,
  "next_billing_date" timestamptz NOT NULL,
  "cancelled_at" timestamptz NULL,
  "cancel_reason" text NULL,
  "gateway" text NOT NULL,
  "gateway_id" text NULL,
  "gateway_response" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_plans_subscriptions" FOREIGN KEY ("plan_id") REFERENCES "plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_subscriptions_deleted_at" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_deleted_at" ON "subscriptions" ("deleted_at");
-- Create index "idx_subscriptions_gateway_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_gateway_id" ON "subscriptions" ("gateway_id");
-- Create index "idx_subscriptions_plan_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_plan_id" ON "subscriptions" ("plan_id");
-- Create index "idx_subscriptions_tenant_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_tenant_id" ON "subscriptions" ("tenant_id");
-- Create index "idx_subscriptions_user_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_user_id" ON "subscriptions" ("user_id");
-- Create "billing_usage" table
CREATE TABLE "billing_usage" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "subscription_id" bigint NOT NULL,
  "metric_type" text NOT NULL,
  "usage" numeric NOT NULL,
  "limit" numeric NOT NULL,
  "overage" numeric NULL DEFAULT 0,
  "billing_period" timestamptz NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_billing_usage_subscription" FOREIGN KEY ("subscription_id") REFERENCES "subscriptions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_billing_usage_billing_period" to table: "billing_usage"
CREATE INDEX "idx_billing_usage_billing_period" ON "billing_usage" ("billing_period");
-- Create index "idx_billing_usage_deleted_at" to table: "billing_usage"
CREATE INDEX "idx_billing_usage_deleted_at" ON "billing_usage" ("deleted_at");
-- Create index "idx_billing_usage_subscription_id" to table: "billing_usage"
CREATE INDEX "idx_billing_usage_subscription_id" ON "billing_usage" ("subscription_id");
-- Create index "idx_billing_usage_tenant_id" to table: "billing_usage"
CREATE INDEX "idx_billing_usage_tenant_id" ON "billing_usage" ("tenant_id");
-- Create index "idx_billing_usage_user_id" to table: "billing_usage"
CREATE INDEX "idx_billing_usage_user_id" ON "billing_usage" ("user_id");
-- Create "payments" table
CREATE TABLE "payments" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "subscription_id" bigint NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL DEFAULT 'USD',
  "status" text NULL DEFAULT 'pending',
  "payment_method" text NOT NULL,
  "gateway" text NOT NULL,
  "gateway_id" text NULL,
  "gateway_response" text NULL,
  "description" text NULL,
  "metadata" text NULL,
  "processed_at" timestamptz NULL,
  "failed_at" timestamptz NULL,
  "refunded_at" timestamptz NULL,
  "refund_amount" numeric NULL DEFAULT 0,
  "refund_reason" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_subscriptions_payments" FOREIGN KEY ("subscription_id") REFERENCES "subscriptions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_payments_deleted_at" to table: "payments"
CREATE INDEX "idx_payments_deleted_at" ON "payments" ("deleted_at");
-- Create index "idx_payments_gateway_id" to table: "payments"
CREATE INDEX "idx_payments_gateway_id" ON "payments" ("gateway_id");
-- Create index "idx_payments_subscription_id" to table: "payments"
CREATE INDEX "idx_payments_subscription_id" ON "payments" ("subscription_id");
-- Create index "idx_payments_tenant_id" to table: "payments"
CREATE INDEX "idx_payments_tenant_id" ON "payments" ("tenant_id");
-- Create index "idx_payments_user_id" to table: "payments"
CREATE INDEX "idx_payments_user_id" ON "payments" ("user_id");
-- Create "invoices" table
CREATE TABLE "invoices" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "user_id" bigint NOT NULL,
  "subscription_id" bigint NULL,
  "invoice_number" text NOT NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL DEFAULT 'USD',
  "status" text NULL DEFAULT 'draft',
  "due_date" timestamptz NOT NULL,
  "paid_at" timestamptz NULL,
  "payment_id" bigint NULL,
  "description" text NULL,
  "items" text NULL,
  "tax_amount" numeric NULL DEFAULT 0,
  "discount_amount" numeric NULL DEFAULT 0,
  "total_amount" numeric NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_invoices_payment" FOREIGN KEY ("payment_id") REFERENCES "payments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_invoices_subscription" FOREIGN KEY ("subscription_id") REFERENCES "subscriptions" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_invoices_deleted_at" to table: "invoices"
CREATE INDEX "idx_invoices_deleted_at" ON "invoices" ("deleted_at");
-- Create index "idx_invoices_invoice_number" to table: "invoices"
CREATE UNIQUE INDEX "idx_invoices_invoice_number" ON "invoices" ("invoice_number");
-- Create index "idx_invoices_payment_id" to table: "invoices"
CREATE INDEX "idx_invoices_payment_id" ON "invoices" ("payment_id");
-- Create index "idx_invoices_subscription_id" to table: "invoices"
CREATE INDEX "idx_invoices_subscription_id" ON "invoices" ("subscription_id");
-- Create index "idx_invoices_tenant_id" to table: "invoices"
CREATE INDEX "idx_invoices_tenant_id" ON "invoices" ("tenant_id");
-- Create index "idx_invoices_user_id" to table: "invoices"
CREATE INDEX "idx_invoices_user_id" ON "invoices" ("user_id");
-- Create "payment_transactions" table
CREATE TABLE "payment_transactions" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "payment_id" bigint NOT NULL,
  "type" text NOT NULL,
  "amount" numeric NOT NULL,
  "currency" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "gateway_id" text NULL,
  "description" text NULL,
  "metadata" text NULL,
  "processed_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_payments_transactions" FOREIGN KEY ("payment_id") REFERENCES "payments" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_payment_transactions_deleted_at" to table: "payment_transactions"
CREATE INDEX "idx_payment_transactions_deleted_at" ON "payment_transactions" ("deleted_at");
-- Create index "idx_payment_transactions_gateway_id" to table: "payment_transactions"
CREATE INDEX "idx_payment_transactions_gateway_id" ON "payment_transactions" ("gateway_id");
-- Create index "idx_payment_transactions_payment_id" to table: "payment_transactions"
CREATE INDEX "idx_payment_transactions_payment_id" ON "payment_transactions" ("payment_id");
