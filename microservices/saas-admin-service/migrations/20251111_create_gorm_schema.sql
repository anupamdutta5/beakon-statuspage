-- Create "saas_backups" table
CREATE TABLE "public"."saas_backups" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "size" bigint NULL,
  "location" text NOT NULL,
  "checksum" text NULL,
  "expires_at" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_backups_deleted_at" to table: "saas_backups"
CREATE INDEX "idx_saas_backups_deleted_at" ON "public"."saas_backups" ("deleted_at");
-- Create index "idx_saas_backups_type" to table: "saas_backups"
CREATE INDEX "idx_saas_backups_type" ON "public"."saas_backups" ("type");
-- Create "saas_integrations" table
CREATE TABLE "public"."saas_integrations" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "status" text NULL DEFAULT 'active',
  "config" text NULL,
  "last_used" timestamptz NULL,
  "last_error" text NULL,
  "event_types" text NULL,
  "enabled" boolean NULL DEFAULT true,
  "description" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_integrations_deleted_at" to table: "saas_integrations"
CREATE INDEX "idx_saas_integrations_deleted_at" ON "public"."saas_integrations" ("deleted_at");
-- Create index "idx_saas_integrations_tenant_id" to table: "saas_integrations"
CREATE INDEX "idx_saas_integrations_tenant_id" ON "public"."saas_integrations" ("tenant_id");
-- Create index "idx_saas_integrations_type" to table: "saas_integrations"
CREATE INDEX "idx_saas_integrations_type" ON "public"."saas_integrations" ("type");
-- Create "domain_stats" table
CREATE TABLE "public"."domain_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "total_domains" bigint NULL,
  "active_domains" bigint NULL,
  "pending_domains" bigint NULL,
  "verified_domains" bigint NULL,
  "failed_domains" bigint NULL,
  "s_slactive_domains" bigint NULL,
  "ssl_pending_domains" bigint NULL,
  "ssl_expired_domains" bigint NULL,
  "last_updated" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "saas_stats" table
CREATE TABLE "public"."saas_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "total_tenants" bigint NULL,
  "active_tenants" bigint NULL,
  "total_users" bigint NULL,
  "active_users" bigint NULL,
  "total_revenue" numeric NULL,
  "monthly_revenue" numeric NULL,
  "total_plans" bigint NULL,
  "active_plans" bigint NULL,
  "total_features" bigint NULL,
  "active_features" bigint NULL,
  "last_updated" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "saas_services" table
CREATE TABLE "public"."saas_services" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "status" text NOT NULL,
  "category" text NULL,
  "position" bigint NULL DEFAULT 0,
  "tenant_id" text NOT NULL,
  "is_visible" boolean NULL DEFAULT true,
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_services_category" to table: "saas_services"
CREATE INDEX "idx_saas_services_category" ON "public"."saas_services" ("category");
-- Create index "idx_saas_services_deleted_at" to table: "saas_services"
CREATE INDEX "idx_saas_services_deleted_at" ON "public"."saas_services" ("deleted_at");
-- Create index "idx_saas_services_status" to table: "saas_services"
CREATE INDEX "idx_saas_services_status" ON "public"."saas_services" ("status");
-- Create index "idx_saas_services_tenant_id" to table: "saas_services"
CREATE INDEX "idx_saas_services_tenant_id" ON "public"."saas_services" ("tenant_id");
-- Create "platforms" table
CREATE TABLE "public"."platforms" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "url" text NOT NULL,
  "description" text NULL,
  "version" text NOT NULL,
  "status" text NULL DEFAULT 'active',
  "admin_email" text NOT NULL,
  "support_email" text NOT NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_platforms_deleted_at" to table: "platforms"
CREATE INDEX "idx_platforms_deleted_at" ON "public"."platforms" ("deleted_at");
-- Create index "idx_platforms_name" to table: "platforms"
CREATE UNIQUE INDEX "idx_platforms_name" ON "public"."platforms" ("name");
-- Create "saas_notifications" table
CREATE TABLE "public"."saas_notifications" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "message" text NOT NULL,
  "type" text NOT NULL,
  "priority" text NULL DEFAULT 'normal',
  "status" text NULL DEFAULT 'active',
  "is_global" boolean NULL DEFAULT false,
  "target_roles" text NULL,
  "expires_at" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_notifications_deleted_at" to table: "saas_notifications"
CREATE INDEX "idx_saas_notifications_deleted_at" ON "public"."saas_notifications" ("deleted_at");
-- Create index "idx_saas_notifications_type" to table: "saas_notifications"
CREATE INDEX "idx_saas_notifications_type" ON "public"."saas_notifications" ("type");
-- Create "base_models" table
CREATE TABLE "public"."base_models" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_base_models_deleted_at" to table: "base_models"
CREATE INDEX "idx_base_models_deleted_at" ON "public"."base_models" ("deleted_at");
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" bigserial NOT NULL,
  "email" text NOT NULL,
  "password_hash" text NOT NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  "tenant_id" uuid NULL,
  "role" character varying(50) NOT NULL DEFAULT 'admin',
  "is_active" boolean NULL DEFAULT true,
  "last_login_at" timestamptz NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "public"."users" ("email");
-- Create index "idx_users_tenant_id" to table: "users"
CREATE INDEX "idx_users_tenant_id" ON "public"."users" ("tenant_id");
-- Create "saas_feature_flags" table
CREATE TABLE "public"."saas_feature_flags" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "is_enabled" boolean NULL DEFAULT false,
  "config" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_feature_flags_deleted_at" to table: "saas_feature_flags"
CREATE INDEX "idx_saas_feature_flags_deleted_at" ON "public"."saas_feature_flags" ("deleted_at");
-- Create index "idx_saas_feature_flags_name" to table: "saas_feature_flags"
CREATE UNIQUE INDEX "idx_saas_feature_flags_name" ON "public"."saas_feature_flags" ("name");
-- Create "custom_domains" table
CREATE TABLE "public"."custom_domains" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "domain" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "verification_key" text NOT NULL,
  "verification_value" text NOT NULL,
  "verification_type" text NULL DEFAULT 'dns',
  "dns_record_type" text NULL DEFAULT 'CNAME',
  "dns_record_value" text NULL,
  "ssl_status" text NULL DEFAULT 'pending',
  "ssl_provider" text NULL DEFAULT 'lets_encrypt',
  "ssl_issue_date" timestamptz NULL,
  "ssl_expiry_date" timestamptz NULL,
  "s_slauto_renew" boolean NULL DEFAULT true,
  "last_verified" timestamptz NULL,
  "verification_tries" bigint NULL DEFAULT 0,
  "max_tries" bigint NULL DEFAULT 5,
  "is_active" boolean NULL DEFAULT false,
  "is_wildcard" boolean NULL DEFAULT false,
  "config" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_custom_domains_deleted_at" to table: "custom_domains"
CREATE INDEX "idx_custom_domains_deleted_at" ON "public"."custom_domains" ("deleted_at");
-- Create index "idx_custom_domains_domain" to table: "custom_domains"
CREATE UNIQUE INDEX "idx_custom_domains_domain" ON "public"."custom_domains" ("domain");
-- Create index "idx_custom_domains_is_active" to table: "custom_domains"
CREATE INDEX "idx_custom_domains_is_active" ON "public"."custom_domains" ("is_active");
-- Create index "idx_custom_domains_tenant_id" to table: "custom_domains"
CREATE INDEX "idx_custom_domains_tenant_id" ON "public"."custom_domains" ("tenant_id");
-- Create "saas_components" table
CREATE TABLE "public"."saas_components" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "status" text NULL DEFAULT 'operational',
  "group_name" text NULL,
  "order" bigint NULL DEFAULT 0,
  "visible" boolean NULL DEFAULT true,
  "show_uptime" boolean NULL DEFAULT true,
  "uptime" numeric NULL DEFAULT 100,
  "link" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_components_deleted_at" to table: "saas_components"
CREATE INDEX "idx_saas_components_deleted_at" ON "public"."saas_components" ("deleted_at");
-- Create index "idx_saas_components_tenant_id" to table: "saas_components"
CREATE INDEX "idx_saas_components_tenant_id" ON "public"."saas_components" ("tenant_id");
-- Create "user_sessions" table
CREATE TABLE "public"."user_sessions" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "token" character varying(255) NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "is_active" boolean NULL DEFAULT true,
  PRIMARY KEY ("id")
);
-- Create index "idx_user_sessions_expires_at" to table: "user_sessions"
CREATE INDEX "idx_user_sessions_expires_at" ON "public"."user_sessions" ("expires_at");
-- Create index "idx_user_sessions_is_active" to table: "user_sessions"
CREATE INDEX "idx_user_sessions_is_active" ON "public"."user_sessions" ("is_active");
-- Create index "idx_user_sessions_token" to table: "user_sessions"
CREATE UNIQUE INDEX "idx_user_sessions_token" ON "public"."user_sessions" ("token");
-- Create index "idx_user_sessions_user_id" to table: "user_sessions"
CREATE INDEX "idx_user_sessions_user_id" ON "public"."user_sessions" ("user_id");
-- Create "saas_features" table
CREATE TABLE "public"."saas_features" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NULL,
  "category" text NOT NULL,
  "is_enabled" boolean NULL DEFAULT true,
  "is_public" boolean NULL DEFAULT true,
  "config" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_features_category" to table: "saas_features"
CREATE INDEX "idx_saas_features_category" ON "public"."saas_features" ("category");
-- Create index "idx_saas_features_deleted_at" to table: "saas_features"
CREATE INDEX "idx_saas_features_deleted_at" ON "public"."saas_features" ("deleted_at");
-- Create index "idx_saas_features_name" to table: "saas_features"
CREATE UNIQUE INDEX "idx_saas_features_name" ON "public"."saas_features" ("name");
-- Create index "idx_saas_features_slug" to table: "saas_features"
CREATE UNIQUE INDEX "idx_saas_features_slug" ON "public"."saas_features" ("slug");
-- Create "sessions" table
CREATE TABLE "public"."sessions" (
  "id" character varying(128) NOT NULL,
  "user_id" bigint NOT NULL,
  "tenant_id" bigint NOT NULL,
  "ip_address" character varying(45) NULL,
  "user_agent" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "last_seen" timestamptz NULL DEFAULT CURRENT_TIMESTAMP,
  "expires_at" timestamptz NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_sessions_expires_at" to table: "sessions"
CREATE INDEX "idx_sessions_expires_at" ON "public"."sessions" ("expires_at");
-- Create index "idx_sessions_is_active" to table: "sessions"
CREATE INDEX "idx_sessions_is_active" ON "public"."sessions" ("is_active");
-- Create index "idx_sessions_tenant_id" to table: "sessions"
CREATE INDEX "idx_sessions_tenant_id" ON "public"."sessions" ("tenant_id");
-- Create index "idx_sessions_user_id" to table: "sessions"
CREATE INDEX "idx_sessions_user_id" ON "public"."sessions" ("user_id");
-- Create "saas_webhooks" table
CREATE TABLE "public"."saas_webhooks" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "name" text NOT NULL,
  "url" text NOT NULL,
  "secret" text NULL,
  "event_types" text NULL,
  "headers" text NULL,
  "method" text NULL DEFAULT 'POST',
  "status" text NULL DEFAULT 'active',
  "last_status" bigint NULL,
  "last_error" text NULL,
  "last_sent" timestamptz NULL,
  "enabled" boolean NULL DEFAULT true,
  "retries" bigint NULL DEFAULT 3,
  "timeout" bigint NULL DEFAULT 30,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_webhooks_deleted_at" to table: "saas_webhooks"
CREATE INDEX "idx_saas_webhooks_deleted_at" ON "public"."saas_webhooks" ("deleted_at");
-- Create index "idx_saas_webhooks_tenant_id" to table: "saas_webhooks"
CREATE INDEX "idx_saas_webhooks_tenant_id" ON "public"."saas_webhooks" ("tenant_id");
-- Create "saas_subscribers" table
CREATE TABLE "public"."saas_subscribers" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "email" text NOT NULL,
  "phone" text NULL,
  "status" text NULL DEFAULT 'active',
  "event_types" text NULL,
  "components" text NULL,
  "verified_at" timestamptz NULL,
  "unsubscribe_token" text NULL,
  "preferences" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_subscribers_deleted_at" to table: "saas_subscribers"
CREATE INDEX "idx_saas_subscribers_deleted_at" ON "public"."saas_subscribers" ("deleted_at");
-- Create index "idx_saas_subscribers_email" to table: "saas_subscribers"
CREATE INDEX "idx_saas_subscribers_email" ON "public"."saas_subscribers" ("email");
-- Create index "idx_saas_subscribers_tenant_id" to table: "saas_subscribers"
CREATE INDEX "idx_saas_subscribers_tenant_id" ON "public"."saas_subscribers" ("tenant_id");
-- Create index "idx_saas_subscribers_unsubscribe_token" to table: "saas_subscribers"
CREATE UNIQUE INDEX "idx_saas_subscribers_unsubscribe_token" ON "public"."saas_subscribers" ("unsubscribe_token");
-- Create "saas_maintenance_windows" table
CREATE TABLE "public"."saas_maintenance_windows" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "status" text NOT NULL,
  "impact" text NOT NULL,
  "scheduled_for" timestamptz NOT NULL,
  "estimated_end" timestamptz NULL,
  "actual_start" timestamptz NULL,
  "actual_end" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "created_by" text NULL,
  "updated_by" text NULL,
  "is_visible" boolean NULL DEFAULT true,
  "notify_subscribers" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_maintenance_windows_deleted_at" to table: "saas_maintenance_windows"
CREATE INDEX "idx_saas_maintenance_windows_deleted_at" ON "public"."saas_maintenance_windows" ("deleted_at");
-- Create index "idx_saas_maintenance_windows_status" to table: "saas_maintenance_windows"
CREATE INDEX "idx_saas_maintenance_windows_status" ON "public"."saas_maintenance_windows" ("status");
-- Create index "idx_saas_maintenance_windows_tenant_id" to table: "saas_maintenance_windows"
CREATE INDEX "idx_saas_maintenance_windows_tenant_id" ON "public"."saas_maintenance_windows" ("tenant_id");
-- Create "domain_verification_challenges" table
CREATE TABLE "public"."domain_verification_challenges" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "domain_id" uuid NOT NULL,
  "type" text NOT NULL,
  "token" text NOT NULL,
  "key_auth" text NOT NULL,
  "value" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "validated_at" timestamptz NULL,
  "expires_at" timestamptz NOT NULL,
  "retry_count" bigint NULL DEFAULT 0,
  "last_error" text NULL,
  "challenge_data" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_domain_verification_challenges_domain" FOREIGN KEY ("domain_id") REFERENCES "public"."custom_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_domain_verification_challenges_deleted_at" to table: "domain_verification_challenges"
CREATE INDEX "idx_domain_verification_challenges_deleted_at" ON "public"."domain_verification_challenges" ("deleted_at");
-- Create index "idx_domain_verification_challenges_domain_id" to table: "domain_verification_challenges"
CREATE INDEX "idx_domain_verification_challenges_domain_id" ON "public"."domain_verification_challenges" ("domain_id");
-- Create index "idx_domain_verification_challenges_type" to table: "domain_verification_challenges"
CREATE INDEX "idx_domain_verification_challenges_type" ON "public"."domain_verification_challenges" ("type");
-- Create "pricing_features" table
CREATE TABLE "public"."pricing_features" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "category" text NOT NULL,
  "icon" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_pricing_features_category" to table: "pricing_features"
CREATE INDEX "idx_pricing_features_category" ON "public"."pricing_features" ("category");
-- Create index "idx_pricing_features_deleted_at" to table: "pricing_features"
CREATE INDEX "idx_pricing_features_deleted_at" ON "public"."pricing_features" ("deleted_at");
-- Create index "idx_pricing_features_is_active" to table: "pricing_features"
CREATE INDEX "idx_pricing_features_is_active" ON "public"."pricing_features" ("is_active");
-- Create "saas_plans" table
CREATE TABLE "public"."saas_plans" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NULL,
  "price" numeric NOT NULL,
  "currency" text NULL DEFAULT 'USD',
  "billing_interval" text NULL DEFAULT 'monthly',
  "max_tenants" bigint NULL DEFAULT 1,
  "max_users" bigint NULL DEFAULT 5,
  "max_services" bigint NULL DEFAULT 10,
  "max_monitors" bigint NULL DEFAULT 50,
  "max_subscribers" bigint NULL DEFAULT 1000,
  "max_incidents" bigint NULL DEFAULT 100,
  "max_maintenance" bigint NULL DEFAULT 50,
  "custom_domain" boolean NULL DEFAULT false,
  "white_label" boolean NULL DEFAULT false,
  "api" boolean NULL DEFAULT false,
  "integrations" boolean NULL DEFAULT false,
  "analytics" boolean NULL DEFAULT false,
  "support" text NULL DEFAULT 'email',
  "is_active" boolean NULL DEFAULT true,
  "is_public" boolean NULL DEFAULT true,
  "is_popular" boolean NULL DEFAULT false,
  "button_text" text NULL DEFAULT 'Get Started',
  "button_url" text NULL DEFAULT '/signup',
  "display_order" bigint NULL DEFAULT 0,
  "features" text NULL,
  "limits" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_plans_deleted_at" to table: "saas_plans"
CREATE INDEX "idx_saas_plans_deleted_at" ON "public"."saas_plans" ("deleted_at");
-- Create index "idx_saas_plans_is_active" to table: "saas_plans"
CREATE INDEX "idx_saas_plans_is_active" ON "public"."saas_plans" ("is_active");
-- Create index "idx_saas_plans_is_popular" to table: "saas_plans"
CREATE INDEX "idx_saas_plans_is_popular" ON "public"."saas_plans" ("is_popular");
-- Create index "idx_saas_plans_is_public" to table: "saas_plans"
CREATE INDEX "idx_saas_plans_is_public" ON "public"."saas_plans" ("is_public");
-- Create index "idx_saas_plans_name" to table: "saas_plans"
CREATE UNIQUE INDEX "idx_saas_plans_name" ON "public"."saas_plans" ("name");
-- Create index "idx_saas_plans_slug" to table: "saas_plans"
CREATE UNIQUE INDEX "idx_saas_plans_slug" ON "public"."saas_plans" ("slug");
-- Create "plan_features" table
CREATE TABLE "public"."plan_features" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "plan_id" uuid NOT NULL,
  "feature_id" bigint NOT NULL,
  "is_enabled" boolean NULL DEFAULT true,
  "order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_plan_features_feature" FOREIGN KEY ("feature_id") REFERENCES "public"."pricing_features" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_saas_plans_plan_features" FOREIGN KEY ("plan_id") REFERENCES "public"."saas_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_plan_features_deleted_at" to table: "plan_features"
CREATE INDEX "idx_plan_features_deleted_at" ON "public"."plan_features" ("deleted_at");
-- Create index "idx_plan_features_feature_id" to table: "plan_features"
CREATE INDEX "idx_plan_features_feature_id" ON "public"."plan_features" ("feature_id");
-- Create index "idx_plan_features_is_enabled" to table: "plan_features"
CREATE INDEX "idx_plan_features_is_enabled" ON "public"."plan_features" ("is_enabled");
-- Create index "idx_plan_features_plan_id" to table: "plan_features"
CREATE INDEX "idx_plan_features_plan_id" ON "public"."plan_features" ("plan_id");
-- Create "pricing_tiers" table
CREATE TABLE "public"."pricing_tiers" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "plan_id" uuid NOT NULL,
  "billing_interval" text NOT NULL,
  "price" numeric NOT NULL,
  "currency" text NULL DEFAULT 'USD',
  "discount_percent" numeric NULL DEFAULT 0,
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_saas_plans_pricing_tiers" FOREIGN KEY ("plan_id") REFERENCES "public"."saas_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_pricing_tiers_deleted_at" to table: "pricing_tiers"
CREATE INDEX "idx_pricing_tiers_deleted_at" ON "public"."pricing_tiers" ("deleted_at");
-- Create index "idx_pricing_tiers_is_active" to table: "pricing_tiers"
CREATE INDEX "idx_pricing_tiers_is_active" ON "public"."pricing_tiers" ("is_active");
-- Create index "idx_pricing_tiers_plan_id" to table: "pricing_tiers"
CREATE INDEX "idx_pricing_tiers_plan_id" ON "public"."pricing_tiers" ("plan_id");
-- Create "saas_admin_users" table
CREATE TABLE "public"."saas_admin_users" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "email" text NOT NULL,
  "username" text NOT NULL,
  "password" text NOT NULL,
  "first_name" text NOT NULL,
  "last_name" text NOT NULL,
  "role" text NULL DEFAULT 'admin',
  "status" text NULL DEFAULT 'active',
  "last_login_at" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_admin_users_deleted_at" to table: "saas_admin_users"
CREATE INDEX "idx_saas_admin_users_deleted_at" ON "public"."saas_admin_users" ("deleted_at");
-- Create index "idx_saas_admin_users_email" to table: "saas_admin_users"
CREATE UNIQUE INDEX "idx_saas_admin_users_email" ON "public"."saas_admin_users" ("email");
-- Create index "idx_saas_admin_users_username" to table: "saas_admin_users"
CREATE UNIQUE INDEX "idx_saas_admin_users_username" ON "public"."saas_admin_users" ("username");
-- Create "saas_activities" table
CREATE TABLE "public"."saas_activities" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "user_id" bigint NOT NULL,
  "action" text NOT NULL,
  "resource" text NOT NULL,
  "resource_id" text NULL,
  "description" text NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_saas_activities_user" FOREIGN KEY ("user_id") REFERENCES "public"."saas_admin_users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_saas_activities_action" to table: "saas_activities"
CREATE INDEX "idx_saas_activities_action" ON "public"."saas_activities" ("action");
-- Create index "idx_saas_activities_deleted_at" to table: "saas_activities"
CREATE INDEX "idx_saas_activities_deleted_at" ON "public"."saas_activities" ("deleted_at");
-- Create index "idx_saas_activities_resource" to table: "saas_activities"
CREATE INDEX "idx_saas_activities_resource" ON "public"."saas_activities" ("resource");
-- Create index "idx_saas_activities_resource_id" to table: "saas_activities"
CREATE INDEX "idx_saas_activities_resource_id" ON "public"."saas_activities" ("resource_id");
-- Create index "idx_saas_activities_user_id" to table: "saas_activities"
CREATE INDEX "idx_saas_activities_user_id" ON "public"."saas_activities" ("user_id");
-- Create "saas_incidents" table
CREATE TABLE "public"."saas_incidents" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "status" text NOT NULL,
  "impact" text NOT NULL,
  "component_status" text NULL DEFAULT 'operational',
  "started_at" timestamptz NOT NULL,
  "resolved_at" timestamptz NULL,
  "tenant_id" text NOT NULL,
  "created_by" text NULL,
  "updated_by" text NULL,
  "is_visible" boolean NULL DEFAULT true,
  "external_id" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_saas_incidents_deleted_at" to table: "saas_incidents"
CREATE INDEX "idx_saas_incidents_deleted_at" ON "public"."saas_incidents" ("deleted_at");
-- Create index "idx_saas_incidents_external_id" to table: "saas_incidents"
CREATE INDEX "idx_saas_incidents_external_id" ON "public"."saas_incidents" ("external_id");
-- Create index "idx_saas_incidents_impact" to table: "saas_incidents"
CREATE INDEX "idx_saas_incidents_impact" ON "public"."saas_incidents" ("impact");
-- Create index "idx_saas_incidents_status" to table: "saas_incidents"
CREATE INDEX "idx_saas_incidents_status" ON "public"."saas_incidents" ("status");
-- Create index "idx_saas_incidents_tenant_id" to table: "saas_incidents"
CREATE INDEX "idx_saas_incidents_tenant_id" ON "public"."saas_incidents" ("tenant_id");
-- Create "saas_incident_updates" table
CREATE TABLE "public"."saas_incident_updates" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "incident_id" uuid NOT NULL,
  "status" text NOT NULL,
  "message" text NOT NULL,
  "created_by" text NULL,
  "is_public" boolean NULL DEFAULT true,
  "update_type" text NULL DEFAULT 'status_update',
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_saas_incidents_updates" FOREIGN KEY ("incident_id") REFERENCES "public"."saas_incidents" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_saas_incident_updates_deleted_at" to table: "saas_incident_updates"
CREATE INDEX "idx_saas_incident_updates_deleted_at" ON "public"."saas_incident_updates" ("deleted_at");
-- Create index "idx_saas_incident_updates_incident_id" to table: "saas_incident_updates"
CREATE INDEX "idx_saas_incident_updates_incident_id" ON "public"."saas_incident_updates" ("incident_id");
-- Create "ssl_certificates" table
CREATE TABLE "public"."ssl_certificates" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "domain_id" uuid NOT NULL,
  "provider" text NOT NULL,
  "certificate_data" text NOT NULL,
  "private_key_data" text NOT NULL,
  "chain_data" text NULL,
  "serial_number" text NULL,
  "fingerprint" text NULL,
  "algorithm" text NULL,
  "key_size" bigint NULL,
  "issued_at" timestamptz NOT NULL,
  "expires_at" timestamptz NOT NULL,
  "is_active" boolean NULL DEFAULT false,
  "auto_renew" boolean NULL DEFAULT true,
  "renewal_days" bigint NULL DEFAULT 30,
  "last_renewal" timestamptz NULL,
  "renewal_status" text NULL DEFAULT 'none',
  "subject_alt_names" text NULL,
  "ocsp" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_ssl_certificates_domain" FOREIGN KEY ("domain_id") REFERENCES "public"."custom_domains" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_ssl_certificates_deleted_at" to table: "ssl_certificates"
CREATE INDEX "idx_ssl_certificates_deleted_at" ON "public"."ssl_certificates" ("deleted_at");
-- Create index "idx_ssl_certificates_domain_id" to table: "ssl_certificates"
CREATE INDEX "idx_ssl_certificates_domain_id" ON "public"."ssl_certificates" ("domain_id");
-- Create index "idx_ssl_certificates_expires_at" to table: "ssl_certificates"
CREATE INDEX "idx_ssl_certificates_expires_at" ON "public"."ssl_certificates" ("expires_at");
-- Create index "idx_ssl_certificates_is_active" to table: "ssl_certificates"
CREATE INDEX "idx_ssl_certificates_is_active" ON "public"."ssl_certificates" ("is_active");
-- Create index "idx_ssl_certificates_serial_number" to table: "ssl_certificates"
CREATE UNIQUE INDEX "idx_ssl_certificates_serial_number" ON "public"."ssl_certificates" ("serial_number");
-- Create "tenants" table
CREATE TABLE "public"."tenants" (
  "id" uuid NOT NULL DEFAULT public.uuid_generate_v4(),
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "domain" text NULL,
  "subdomain" text NULL,
  "contact_email" text NOT NULL,
  "billing_email" text NULL,
  "plan_id" uuid NOT NULL,
  "status" text NULL DEFAULT 'active',
  "is_active" boolean NULL DEFAULT true,
  "max_users" bigint NULL,
  "settings" text NULL,
  "branding" text NULL,
  "features" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_tenants_plan" FOREIGN KEY ("plan_id") REFERENCES "public"."saas_plans" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_tenants_deleted_at" to table: "tenants"
CREATE INDEX "idx_tenants_deleted_at" ON "public"."tenants" ("deleted_at");
-- Create index "idx_tenants_domain" to table: "tenants"
CREATE UNIQUE INDEX "idx_tenants_domain" ON "public"."tenants" ("domain");
-- Create index "idx_tenants_plan_id" to table: "tenants"
CREATE INDEX "idx_tenants_plan_id" ON "public"."tenants" ("plan_id");
-- Create index "idx_tenants_slug" to table: "tenants"
CREATE UNIQUE INDEX "idx_tenants_slug" ON "public"."tenants" ("slug");
-- Create index "idx_tenants_subdomain" to table: "tenants"
CREATE UNIQUE INDEX "idx_tenants_subdomain" ON "public"."tenants" ("subdomain");
