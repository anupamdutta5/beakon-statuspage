-- Create "ab_test_assignments" table
CREATE TABLE "ab_test_assignments" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "test_id" bigint NOT NULL,
  "session_id" text NOT NULL,
  "variant" text NOT NULL,
  "user_id" text NULL,
  "converted" boolean NULL DEFAULT false,
  PRIMARY KEY ("id")
);
-- Create index "idx_ab_test_assignments_user_id" to table: "ab_test_assignments"
CREATE INDEX "idx_ab_test_assignments_user_id" ON "ab_test_assignments" ("user_id");
-- Create index "idx_test_session" to table: "ab_test_assignments"
CREATE INDEX "idx_test_session" ON "ab_test_assignments" ("test_id", "session_id");
-- Create "articles" table
CREATE TABLE "articles" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "slug" text NOT NULL,
  "excerpt" text NULL,
  "content" text NOT NULL,
  "author" text NOT NULL,
  "author_email" text NULL,
  "author_avatar" text NULL,
  "featured_image" text NULL,
  "category" text NULL,
  "tags" text NULL,
  "status" text NULL DEFAULT 'draft',
  "is_featured" boolean NULL DEFAULT false,
  "view_count" bigint NULL DEFAULT 0,
  "published_at" timestamptz NULL,
  "meta_title" text NULL,
  "meta_description" text NULL,
  "canonical_url" text NULL,
  "schema_markup" text NULL,
  "reading_time_minutes" bigint NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_articles_category" to table: "articles"
CREATE INDEX "idx_articles_category" ON "articles" ("category");
-- Create index "idx_articles_deleted_at" to table: "articles"
CREATE INDEX "idx_articles_deleted_at" ON "articles" ("deleted_at");
-- Create index "idx_articles_slug" to table: "articles"
CREATE UNIQUE INDEX "idx_articles_slug" ON "articles" ("slug");
-- Create "contact_forms" table
CREATE TABLE "contact_forms" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "email" text NOT NULL,
  "company" text NULL,
  "subject" text NULL,
  "message" text NOT NULL,
  "status" text NULL DEFAULT 'pending',
  "ip_address" text NULL,
  "user_agent" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_contact_forms_deleted_at" to table: "contact_forms"
CREATE INDEX "idx_contact_forms_deleted_at" ON "contact_forms" ("deleted_at");
-- Create "faqs" table
CREATE TABLE "faqs" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "question" text NOT NULL,
  "answer" text NOT NULL,
  "category" text NULL,
  "status" text NULL DEFAULT 'active',
  "sort_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_faqs_category" to table: "faqs"
CREATE INDEX "idx_faqs_category" ON "faqs" ("category");
-- Create index "idx_faqs_deleted_at" to table: "faqs"
CREATE INDEX "idx_faqs_deleted_at" ON "faqs" ("deleted_at");
-- Create "feature_sections" table
CREATE TABLE "feature_sections" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "icon" text NULL,
  "image_url" text NULL,
  "status" text NULL DEFAULT 'active',
  "sort_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_feature_sections_deleted_at" to table: "feature_sections"
CREATE INDEX "idx_feature_sections_deleted_at" ON "feature_sections" ("deleted_at");
-- Create "hero_sections" table
CREATE TABLE "hero_sections" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "title" text NOT NULL,
  "subtitle" text NULL,
  "description" text NULL,
  "button_text" text NULL,
  "button_url" text NULL,
  "image_url" text NULL,
  "video_url" text NULL,
  "background_color" text NULL,
  "text_color" text NULL,
  "status" text NULL DEFAULT 'active',
  "sort_order" bigint NULL DEFAULT 0,
  "ab_test_id" bigint NULL,
  "variant" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_hero_sections_ab_test_id" to table: "hero_sections"
CREATE INDEX "idx_hero_sections_ab_test_id" ON "hero_sections" ("ab_test_id");
-- Create index "idx_hero_sections_deleted_at" to table: "hero_sections"
CREATE INDEX "idx_hero_sections_deleted_at" ON "hero_sections" ("deleted_at");
-- Create "landing_page_ab_tests" table
CREATE TABLE "landing_page_ab_tests" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "element_type" text NOT NULL,
  "element_id" bigint NULL,
  "status" text NULL DEFAULT 'draft',
  "start_date" timestamptz NULL,
  "end_date" timestamptz NULL,
  "variant_a_config" text NULL,
  "variant_b_config" text NULL,
  "traffic_split" bigint NULL DEFAULT 50,
  "variant_a_views" bigint NULL DEFAULT 0,
  "variant_a_conversions" bigint NULL DEFAULT 0,
  "variant_b_views" bigint NULL DEFAULT 0,
  "variant_b_conversions" bigint NULL DEFAULT 0,
  "winner" text NULL,
  "confidence_level" numeric NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_page_ab_tests_deleted_at" to table: "landing_page_ab_tests"
CREATE INDEX "idx_landing_page_ab_tests_deleted_at" ON "landing_page_ab_tests" ("deleted_at");
-- Create index "idx_landing_page_ab_tests_element_id" to table: "landing_page_ab_tests"
CREATE INDEX "idx_landing_page_ab_tests_element_id" ON "landing_page_ab_tests" ("element_id");
-- Create index "idx_landing_page_ab_tests_element_type" to table: "landing_page_ab_tests"
CREATE INDEX "idx_landing_page_ab_tests_element_type" ON "landing_page_ab_tests" ("element_type");
-- Create index "idx_landing_page_ab_tests_status" to table: "landing_page_ab_tests"
CREATE INDEX "idx_landing_page_ab_tests_status" ON "landing_page_ab_tests" ("status");
-- Create "landing_page_cta_buttons" table
CREATE TABLE "landing_page_cta_buttons" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "text" text NOT NULL,
  "url" text NOT NULL,
  "style" text NULL DEFAULT 'primary',
  "size" text NULL DEFAULT 'medium',
  "icon" text NULL,
  "click_count" bigint NULL DEFAULT 0,
  "conversion_count" bigint NULL DEFAULT 0,
  "locations" text NULL,
  "is_active" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_page_cta_buttons_deleted_at" to table: "landing_page_cta_buttons"
CREATE INDEX "idx_landing_page_cta_buttons_deleted_at" ON "landing_page_cta_buttons" ("deleted_at");
-- Create "landing_page_integrations" table
CREATE TABLE "landing_page_integrations" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "integration_type" text NOT NULL,
  "provider" text NOT NULL,
  "is_enabled" boolean NULL DEFAULT false,
  "config" text NULL,
  "api_key" text NULL,
  "last_sync_at" timestamptz NULL,
  "sync_status" text NULL,
  "error_message" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_integration_type_provider" to table: "landing_page_integrations"
CREATE UNIQUE INDEX "idx_integration_type_provider" ON "landing_page_integrations" ("integration_type", "provider");
-- Create "landing_page_media" table
CREATE TABLE "landing_page_media" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "filename" text NOT NULL,
  "original_filename" text NOT NULL,
  "mime_type" text NOT NULL,
  "size_bytes" bigint NOT NULL,
  "width" bigint NULL,
  "height" bigint NULL,
  "storage_path" text NOT NULL,
  "cdn_url" text NULL,
  "is_optimized" boolean NULL DEFAULT false,
  "web_p_path" text NULL,
  "avif_path" text NULL,
  "thumbnail_path" text NULL,
  "alt_text" text NULL,
  "caption" text NULL,
  "title" text NULL,
  "category" text NULL,
  "usage_count" bigint NULL DEFAULT 0,
  "last_used_at" timestamptz NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_page_media_category" to table: "landing_page_media"
CREATE INDEX "idx_landing_page_media_category" ON "landing_page_media" ("category");
-- Create index "idx_landing_page_media_deleted_at" to table: "landing_page_media"
CREATE INDEX "idx_landing_page_media_deleted_at" ON "landing_page_media" ("deleted_at");
-- Create "landing_page_sections" table
CREATE TABLE "landing_page_sections" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "section_type" text NOT NULL,
  "title" text NULL,
  "content" text NULL,
  "position" text NULL,
  "order" bigint NULL DEFAULT 0,
  "background_color" text NULL,
  "status" text NULL DEFAULT 'active',
  "layout" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_page_sections_deleted_at" to table: "landing_page_sections"
CREATE INDEX "idx_landing_page_sections_deleted_at" ON "landing_page_sections" ("deleted_at");
-- Create index "idx_landing_page_sections_position" to table: "landing_page_sections"
CREATE INDEX "idx_landing_page_sections_position" ON "landing_page_sections" ("position");
-- Create index "idx_landing_page_sections_section_type" to table: "landing_page_sections"
CREATE INDEX "idx_landing_page_sections_section_type" ON "landing_page_sections" ("section_type");
-- Create "landing_page_seo_config" table
CREATE TABLE "landing_page_seo_config" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "site_name" text NOT NULL,
  "site_url" text NOT NULL,
  "default_meta_description" text NULL,
  "default_og_image" text NULL,
  "google_analytics_id" text NULL,
  "google_tag_manager_id" text NULL,
  "facebook_pixel_id" text NULL,
  "plausible_domain" text NULL,
  "organization_name" text NULL,
  "organization_logo" text NULL,
  "organization_url" text NULL,
  "organization_description" text NULL,
  "organization_email" text NULL,
  "organization_phone" text NULL,
  "organization_address" text NULL,
  "organization_social_links" text NULL,
  "sitemap_enabled" boolean NULL DEFAULT true,
  "sitemap_change_freq" text NULL DEFAULT 'daily',
  "sitemap_priority" numeric NULL DEFAULT 0.8,
  "robots_txt" text NULL,
  "robots_enabled" boolean NULL DEFAULT true,
  "schema_enabled" boolean NULL DEFAULT true,
  "twitter_card" text NULL,
  "default_title" text NULL,
  "default_keywords" text NULL,
  "og_default_title" text NULL,
  "og_default_description" text NULL,
  "google_site_verification" text NULL,
  "bing_site_verification" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create "landing_page_stats" table
CREATE TABLE "landing_page_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "date" timestamptz NOT NULL,
  "page_views" bigint NULL DEFAULT 0,
  "unique_visitors" bigint NULL DEFAULT 0,
  "contact_forms" bigint NULL DEFAULT 0,
  "newsletter_signups" bigint NULL DEFAULT 0,
  "plan_views" bigint NULL DEFAULT 0,
  "plan_clicks" bigint NULL DEFAULT 0,
  "article_views" bigint NULL DEFAULT 0,
  "bounce_rate" numeric NULL DEFAULT 0,
  "avg_session_time" numeric NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_page_stats_date" to table: "landing_page_stats"
CREATE INDEX "idx_landing_page_stats_date" ON "landing_page_stats" ("date");
-- Create "landing_pages" table
CREATE TABLE "landing_pages" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "title" text NOT NULL,
  "description" text NULL,
  "content" text NULL,
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "meta_title" text NULL,
  "meta_description" text NULL,
  "meta_keywords" text NULL,
  "canonical_url" text NULL,
  "meta_robots" text NULL,
  "og_title" text NULL,
  "og_description" text NULL,
  "og_image" text NULL,
  "twitter_card" text NULL,
  "schema_markup" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_landing_pages_deleted_at" to table: "landing_pages"
CREATE INDEX "idx_landing_pages_deleted_at" ON "landing_pages" ("deleted_at");
-- Create index "idx_landing_pages_name" to table: "landing_pages"
CREATE UNIQUE INDEX "idx_landing_pages_name" ON "landing_pages" ("name");
-- Create index "idx_landing_pages_slug" to table: "landing_pages"
CREATE UNIQUE INDEX "idx_landing_pages_slug" ON "landing_pages" ("slug");
-- Create "newsletters" table
CREATE TABLE "newsletters" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "email" text NOT NULL,
  "name" text NULL,
  "status" text NULL DEFAULT 'active',
  "source" text NULL,
  "ip_address" text NULL,
  "user_agent" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_newsletters_deleted_at" to table: "newsletters"
CREATE INDEX "idx_newsletters_deleted_at" ON "newsletters" ("deleted_at");
-- Create index "idx_newsletters_email" to table: "newsletters"
CREATE UNIQUE INDEX "idx_newsletters_email" ON "newsletters" ("email");
-- Create "pricing_plans" table
CREATE TABLE "pricing_plans" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "plan_id" text NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NULL,
  "price" numeric NOT NULL,
  "currency" text NULL DEFAULT 'USD',
  "billing_interval" text NULL DEFAULT 'monthly',
  "features" text NULL,
  "is_popular" boolean NULL DEFAULT false,
  "is_active" boolean NULL DEFAULT true,
  "button_text" text NULL,
  "button_url" text NULL,
  "status" text NULL DEFAULT 'active',
  "sort_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_pricing_plans_deleted_at" to table: "pricing_plans"
CREATE INDEX "idx_pricing_plans_deleted_at" ON "pricing_plans" ("deleted_at");
-- Create index "idx_pricing_plans_plan_id" to table: "pricing_plans"
CREATE UNIQUE INDEX "idx_pricing_plans_plan_id" ON "pricing_plans" ("plan_id");
-- Create index "idx_pricing_plans_slug" to table: "pricing_plans"
CREATE UNIQUE INDEX "idx_pricing_plans_slug" ON "pricing_plans" ("slug");
-- Create "seo_contents" table
CREATE TABLE "seo_contents" (
  "id" bigserial NOT NULL,
  "type" text NULL,
  "slug" text NULL,
  "title" text NULL,
  "content" text NULL,
  "excerpt" text NULL,
  "seo_title" text NULL,
  "seo_description" text NULL,
  "seo_keywords" text NULL,
  "focus_keyword" text NULL,
  "author" text NULL,
  "published_at" timestamptz NULL,
  "last_modified" timestamptz NULL,
  "status" text NULL,
  "view_count" bigint NULL,
  "seo_score" bigint NULL,
  "reading_time" bigint NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_seo_contents_slug" to table: "seo_contents"
CREATE UNIQUE INDEX "idx_seo_contents_slug" ON "seo_contents" ("slug");
-- Create "seo_redirects" table
CREATE TABLE "seo_redirects" (
  "id" bigserial NOT NULL,
  "from_url" text NULL,
  "to_url" text NULL,
  "status_code" bigint NULL,
  "is_active" boolean NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_seo_redirects_from_url" to table: "seo_redirects"
CREATE UNIQUE INDEX "idx_seo_redirects_from_url" ON "seo_redirects" ("from_url");
-- Create "testimonials" table
CREATE TABLE "testimonials" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "company" text NULL,
  "position" text NULL,
  "avatar" text NULL,
  "content" text NOT NULL,
  "rating" bigint NULL DEFAULT 5,
  "status" text NULL DEFAULT 'active',
  "sort_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_testimonials_deleted_at" to table: "testimonials"
CREATE INDEX "idx_testimonials_deleted_at" ON "testimonials" ("deleted_at");
