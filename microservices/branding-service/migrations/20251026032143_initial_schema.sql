-- Create "branding_stats" table
CREATE TABLE "branding_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "total_brands" bigint NULL,
  "total_themes" bigint NULL,
  "total_assets" bigint NULL,
  "total_custom_css" bigint NULL,
  "total_custom_js" bigint NULL,
  "total_layouts" bigint NULL,
  "total_components" bigint NULL,
  "last_updated" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "brands" table
CREATE TABLE "brands" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "tenant_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NULL,
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_brands_deleted_at" to table: "brands"
CREATE INDEX "idx_brands_deleted_at" ON "brands" ("deleted_at");
-- Create index "idx_brands_name" to table: "brands"
CREATE INDEX "idx_brands_name" ON "brands" ("name");
-- Create index "idx_brands_slug" to table: "brands"
CREATE UNIQUE INDEX "idx_brands_slug" ON "brands" ("slug");
-- Create index "idx_brands_tenant_id" to table: "brands"
CREATE INDEX "idx_brands_tenant_id" ON "brands" ("tenant_id");
-- Create "assets" table
CREATE TABLE "assets" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "brand_id" bigint NOT NULL,
  "name" text NOT NULL,
  "type" text NOT NULL,
  "category" text NOT NULL,
  "filename" text NOT NULL,
  "original_name" text NOT NULL,
  "mime_type" text NOT NULL,
  "size" bigint NOT NULL,
  "width" bigint NULL,
  "height" bigint NULL,
  "url" text NOT NULL,
  "thumbnail_url" text NULL,
  "alt_text" text NULL,
  "description" text NULL,
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_assets_brand" FOREIGN KEY ("brand_id") REFERENCES "brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_assets_brand_id" to table: "assets"
CREATE INDEX "idx_assets_brand_id" ON "assets" ("brand_id");
-- Create index "idx_assets_category" to table: "assets"
CREATE INDEX "idx_assets_category" ON "assets" ("category");
-- Create index "idx_assets_deleted_at" to table: "assets"
CREATE INDEX "idx_assets_deleted_at" ON "assets" ("deleted_at");
-- Create index "idx_assets_name" to table: "assets"
CREATE INDEX "idx_assets_name" ON "assets" ("name");
-- Create index "idx_assets_type" to table: "assets"
CREATE INDEX "idx_assets_type" ON "assets" ("type");
-- Create "themes" table
CREATE TABLE "themes" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "brand_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "description" text NULL,
  "version" text NULL DEFAULT '1.0.0',
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "is_public" boolean NULL DEFAULT true,
  "preview_url" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_brands_themes" FOREIGN KEY ("brand_id") REFERENCES "brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_themes_brand_id" to table: "themes"
CREATE INDEX "idx_themes_brand_id" ON "themes" ("brand_id");
-- Create index "idx_themes_deleted_at" to table: "themes"
CREATE INDEX "idx_themes_deleted_at" ON "themes" ("deleted_at");
-- Create index "idx_themes_name" to table: "themes"
CREATE INDEX "idx_themes_name" ON "themes" ("name");
-- Create index "idx_themes_slug" to table: "themes"
CREATE INDEX "idx_themes_slug" ON "themes" ("slug");
-- Create "color_schemes" table
CREATE TABLE "color_schemes" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "theme_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "primary" text NOT NULL,
  "secondary" text NOT NULL,
  "accent" text NULL,
  "background" text NOT NULL,
  "surface" text NOT NULL,
  "text" text NOT NULL,
  "text_secondary" text NULL,
  "success" text NULL,
  "warning" text NULL,
  "error" text NULL,
  "info" text NULL,
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_color_schemes_theme" FOREIGN KEY ("theme_id") REFERENCES "themes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_color_schemes_deleted_at" to table: "color_schemes"
CREATE INDEX "idx_color_schemes_deleted_at" ON "color_schemes" ("deleted_at");
-- Create index "idx_color_schemes_name" to table: "color_schemes"
CREATE INDEX "idx_color_schemes_name" ON "color_schemes" ("name");
-- Create index "idx_color_schemes_slug" to table: "color_schemes"
CREATE INDEX "idx_color_schemes_slug" ON "color_schemes" ("slug");
-- Create index "idx_color_schemes_theme_id" to table: "color_schemes"
CREATE INDEX "idx_color_schemes_theme_id" ON "color_schemes" ("theme_id");
-- Create "components" table
CREATE TABLE "components" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "theme_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "type" text NOT NULL,
  "category" text NOT NULL,
  "config" text NOT NULL,
  "css" text NULL,
  "html" text NULL,
  "java_script" text NULL,
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "is_public" boolean NULL DEFAULT true,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_components_theme" FOREIGN KEY ("theme_id") REFERENCES "themes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_components_category" to table: "components"
CREATE INDEX "idx_components_category" ON "components" ("category");
-- Create index "idx_components_deleted_at" to table: "components"
CREATE INDEX "idx_components_deleted_at" ON "components" ("deleted_at");
-- Create index "idx_components_name" to table: "components"
CREATE INDEX "idx_components_name" ON "components" ("name");
-- Create index "idx_components_slug" to table: "components"
CREATE INDEX "idx_components_slug" ON "components" ("slug");
-- Create index "idx_components_theme_id" to table: "components"
CREATE INDEX "idx_components_theme_id" ON "components" ("theme_id");
-- Create index "idx_components_type" to table: "components"
CREATE INDEX "idx_components_type" ON "components" ("type");
-- Create "custom_css" table
CREATE TABLE "custom_css" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "brand_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "css" text NOT NULL,
  "version" text NULL DEFAULT '1.0.0',
  "status" text NULL DEFAULT 'active',
  "is_minified" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_custom_css_brand" FOREIGN KEY ("brand_id") REFERENCES "brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_custom_css_brand_id" to table: "custom_css"
CREATE INDEX "idx_custom_css_brand_id" ON "custom_css" ("brand_id");
-- Create index "idx_custom_css_deleted_at" to table: "custom_css"
CREATE INDEX "idx_custom_css_deleted_at" ON "custom_css" ("deleted_at");
-- Create index "idx_custom_css_name" to table: "custom_css"
CREATE INDEX "idx_custom_css_name" ON "custom_css" ("name");
-- Create "custom_js" table
CREATE TABLE "custom_js" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "brand_id" bigint NOT NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "java_script" text NOT NULL,
  "version" text NULL DEFAULT '1.0.0',
  "status" text NULL DEFAULT 'active',
  "is_minified" boolean NULL DEFAULT false,
  "load_order" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_custom_js_brand" FOREIGN KEY ("brand_id") REFERENCES "brands" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_custom_js_brand_id" to table: "custom_js"
CREATE INDEX "idx_custom_js_brand_id" ON "custom_js" ("brand_id");
-- Create index "idx_custom_js_deleted_at" to table: "custom_js"
CREATE INDEX "idx_custom_js_deleted_at" ON "custom_js" ("deleted_at");
-- Create index "idx_custom_js_name" to table: "custom_js"
CREATE INDEX "idx_custom_js_name" ON "custom_js" ("name");
-- Create "layouts" table
CREATE TABLE "layouts" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "theme_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "type" text NOT NULL,
  "config" text NOT NULL,
  "css" text NULL,
  "html" text NULL,
  "status" text NULL DEFAULT 'active',
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_layouts_theme" FOREIGN KEY ("theme_id") REFERENCES "themes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_layouts_deleted_at" to table: "layouts"
CREATE INDEX "idx_layouts_deleted_at" ON "layouts" ("deleted_at");
-- Create index "idx_layouts_name" to table: "layouts"
CREATE INDEX "idx_layouts_name" ON "layouts" ("name");
-- Create index "idx_layouts_slug" to table: "layouts"
CREATE INDEX "idx_layouts_slug" ON "layouts" ("slug");
-- Create index "idx_layouts_theme_id" to table: "layouts"
CREATE INDEX "idx_layouts_theme_id" ON "layouts" ("theme_id");
-- Create index "idx_layouts_type" to table: "layouts"
CREATE INDEX "idx_layouts_type" ON "layouts" ("type");
-- Create "typographies" table
CREATE TABLE "typographies" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "theme_id" bigint NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL,
  "font_family" text NOT NULL,
  "font_size" text NOT NULL,
  "line_height" text NOT NULL,
  "font_weight" text NOT NULL,
  "font_style" text NULL,
  "letter_spacing" text NULL,
  "text_transform" text NULL,
  "is_default" boolean NULL DEFAULT false,
  "metadata" text NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_typographies_theme" FOREIGN KEY ("theme_id") REFERENCES "themes" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_typographies_deleted_at" to table: "typographies"
CREATE INDEX "idx_typographies_deleted_at" ON "typographies" ("deleted_at");
-- Create index "idx_typographies_name" to table: "typographies"
CREATE INDEX "idx_typographies_name" ON "typographies" ("name");
-- Create index "idx_typographies_slug" to table: "typographies"
CREATE INDEX "idx_typographies_slug" ON "typographies" ("slug");
-- Create index "idx_typographies_theme_id" to table: "typographies"
CREATE INDEX "idx_typographies_theme_id" ON "typographies" ("theme_id");
