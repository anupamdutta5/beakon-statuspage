-- Create "event_store_stats" table
CREATE TABLE "event_store_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "total_streams" bigint NULL,
  "total_events" bigint NULL,
  "total_projections" bigint NULL,
  "total_snapshots" bigint NULL,
  "last_updated" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create "projections" table
CREATE TABLE "projections" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "description" text NULL,
  "query" text NOT NULL,
  "status" text NULL DEFAULT 'stopped',
  "last_event" text NULL,
  "last_updated" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_projections_deleted_at" to table: "projections"
CREATE INDEX "idx_projections_deleted_at" ON "projections" ("deleted_at");
-- Create index "idx_projections_last_event" to table: "projections"
CREATE INDEX "idx_projections_last_event" ON "projections" ("last_event");
-- Create index "idx_projections_name" to table: "projections"
CREATE UNIQUE INDEX "idx_projections_name" ON "projections" ("name");
-- Create "stream_stats" table
CREATE TABLE "stream_stats" (
  "id" bigserial NOT NULL,
  "created_at" timestamptz NULL,
  "stream_id" text NOT NULL,
  "event_count" bigint NULL,
  "snapshot_count" bigint NULL,
  "last_event_at" timestamptz NULL,
  "last_updated" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_stream_stats_stream_id" to table: "stream_stats"
CREATE INDEX "idx_stream_stats_stream_id" ON "stream_stats" ("stream_id");
-- Create "streams" table
CREATE TABLE "streams" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "type" text NOT NULL,
  "version" bigint NULL DEFAULT 0,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_streams_deleted_at" to table: "streams"
CREATE INDEX "idx_streams_deleted_at" ON "streams" ("deleted_at");
-- Create index "idx_streams_type" to table: "streams"
CREATE INDEX "idx_streams_type" ON "streams" ("type");
-- Create "subscriptions" table
CREATE TABLE "subscriptions" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "name" text NOT NULL,
  "stream_id" text NULL,
  "event_types" text NULL,
  "endpoint" text NOT NULL,
  "status" text NULL DEFAULT 'active',
  "last_event" text NULL,
  "last_updated" timestamptz NULL,
  "error" text NULL,
  "metadata" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_subscriptions_deleted_at" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_deleted_at" ON "subscriptions" ("deleted_at");
-- Create index "idx_subscriptions_last_event" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_last_event" ON "subscriptions" ("last_event");
-- Create index "idx_subscriptions_name" to table: "subscriptions"
CREATE UNIQUE INDEX "idx_subscriptions_name" ON "subscriptions" ("name");
-- Create index "idx_subscriptions_stream_id" to table: "subscriptions"
CREATE INDEX "idx_subscriptions_stream_id" ON "subscriptions" ("stream_id");
-- Create "events" table
CREATE TABLE "events" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "stream_id" text NOT NULL,
  "type" text NOT NULL,
  "version" bigint NOT NULL,
  "data" text NOT NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_events_stream" FOREIGN KEY ("stream_id") REFERENCES "streams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_events_deleted_at" to table: "events"
CREATE INDEX "idx_events_deleted_at" ON "events" ("deleted_at");
-- Create index "idx_events_stream_id" to table: "events"
CREATE INDEX "idx_events_stream_id" ON "events" ("stream_id");
-- Create index "idx_events_timestamp" to table: "events"
CREATE INDEX "idx_events_timestamp" ON "events" ("timestamp");
-- Create index "idx_events_type" to table: "events"
CREATE INDEX "idx_events_type" ON "events" ("type");
-- Create index "idx_events_version" to table: "events"
CREATE INDEX "idx_events_version" ON "events" ("version");
-- Create "snapshots" table
CREATE TABLE "snapshots" (
  "id" text NOT NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  "stream_id" text NOT NULL,
  "version" bigint NOT NULL,
  "data" text NOT NULL,
  "metadata" text NULL,
  "timestamp" timestamptz NOT NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_snapshots_stream" FOREIGN KEY ("stream_id") REFERENCES "streams" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_snapshots_deleted_at" to table: "snapshots"
CREATE INDEX "idx_snapshots_deleted_at" ON "snapshots" ("deleted_at");
-- Create index "idx_snapshots_stream_id" to table: "snapshots"
CREATE INDEX "idx_snapshots_stream_id" ON "snapshots" ("stream_id");
-- Create index "idx_snapshots_timestamp" to table: "snapshots"
CREATE INDEX "idx_snapshots_timestamp" ON "snapshots" ("timestamp");
-- Create index "idx_snapshots_version" to table: "snapshots"
CREATE INDEX "idx_snapshots_version" ON "snapshots" ("version");
