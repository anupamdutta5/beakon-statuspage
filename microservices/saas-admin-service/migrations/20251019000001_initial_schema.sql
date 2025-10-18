--
-- PostgreSQL database dump
--

-- Dumped from database version 14.18 (Homebrew)
-- Dumped by pg_dump version 14.18 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: uuid-ossp; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA public;


--
-- Name: EXTENSION "uuid-ossp"; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION "uuid-ossp" IS 'generate universally unique identifiers (UUIDs)';


--
-- Name: monitor_type; Type: TYPE; Schema: public; Owner: -
--

CREATE TYPE public.monitor_type AS ENUM (
    'http',
    'https',
    'tcp',
    'udp',
    'ping',
    'dns',
    'docker',
    'kubernetes',
    'api',
    'custom'
);


--
-- Name: update_updated_at_column(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.update_updated_at_column() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: analytics_data; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.analytics_data (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid,
    date date NOT NULL,
    page_views integer DEFAULT 0,
    unique_visitors integer DEFAULT 0,
    uptime_percentage numeric(5,2) DEFAULT 100.00,
    avg_response_time integer DEFAULT 0,
    incident_count integer DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: audit_logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.audit_logs (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    user_id bigint NOT NULL,
    action character varying(100) NOT NULL,
    resource character varying(100) NOT NULL,
    resource_id bigint,
    details text,
    ip_address character varying(45),
    user_agent text,
    success boolean DEFAULT true,
    error_message text,
    created_at timestamp with time zone
);


--
-- Name: audit_logs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.audit_logs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: audit_logs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.audit_logs_id_seq OWNED BY public.audit_logs.id;


--
-- Name: component_alerts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_alerts (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    component_id bigint NOT NULL,
    alert_type text NOT NULL,
    status text NOT NULL,
    message text,
    threshold numeric,
    current_value numeric,
    resolved_at timestamp with time zone,
    resolved_by bigint,
    metadata text
);


--
-- Name: component_alerts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_alerts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_alerts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_alerts_id_seq OWNED BY public.component_alerts.id;


--
-- Name: component_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_groups (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    "position" bigint DEFAULT 0,
    is_visible boolean DEFAULT true
);


--
-- Name: component_groups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_groups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_groups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_groups_id_seq OWNED BY public.component_groups.id;


--
-- Name: component_history; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_history (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    component_id bigint NOT NULL,
    old_status text,
    new_status text NOT NULL,
    message text,
    updated_by bigint,
    metadata text
);


--
-- Name: component_history_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_history_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_history_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_history_id_seq OWNED BY public.component_history.id;


--
-- Name: component_metrics; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_metrics (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    component_id bigint NOT NULL,
    metric_type text NOT NULL,
    value numeric NOT NULL,
    unit text,
    "timestamp" timestamp with time zone NOT NULL,
    metadata text
);


--
-- Name: component_metrics_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_metrics_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_metrics_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_metrics_id_seq OWNED BY public.component_metrics.id;


--
-- Name: component_statuses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_statuses (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    component_id bigint NOT NULL,
    status text NOT NULL,
    message text,
    updated_by bigint,
    metadata text
);


--
-- Name: component_statuses_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_statuses_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_statuses_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_statuses_id_seq OWNED BY public.component_statuses.id;


--
-- Name: component_webhooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.component_webhooks (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    component_id bigint NOT NULL,
    url text NOT NULL,
    events text,
    secret text,
    is_active boolean DEFAULT true,
    last_triggered timestamp with time zone,
    metadata text
);


--
-- Name: component_webhooks_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.component_webhooks_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: component_webhooks_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.component_webhooks_id_seq OWNED BY public.component_webhooks.id;


--
-- Name: components; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.components (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    status text DEFAULT 'operational'::text,
    "position" bigint DEFAULT 0,
    is_visible boolean DEFAULT true,
    group_id bigint,
    metadata text
);


--
-- Name: components_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.components_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: components_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.components_id_seq OWNED BY public.components.id;


--
-- Name: domain_verification_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.domain_verification_records (
    type text,
    name text,
    value text
);


--
-- Name: email_verifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.email_verifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    user_id bigint NOT NULL,
    token text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    is_used boolean DEFAULT false
);


--
-- Name: email_verifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.email_verifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: email_verifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.email_verifications_id_seq OWNED BY public.email_verifications.id;


--
-- Name: incident_updates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incident_updates (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    incident_id integer NOT NULL,
    status character varying(50) NOT NULL,
    message text NOT NULL,
    is_visible boolean DEFAULT true,
    created_by integer,
    metadata text
);


--
-- Name: incident_updates_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.incident_updates_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: incident_updates_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.incident_updates_id_seq OWNED BY public.incident_updates.id;


--
-- Name: incidents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.incidents (
    id integer NOT NULL,
    created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    updated_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    deleted_at timestamp without time zone,
    tenant_id integer NOT NULL,
    title character varying(255) NOT NULL,
    description text,
    status character varying(50) DEFAULT 'investigating'::character varying,
    impact character varying(50) DEFAULT 'minor'::character varying,
    severity character varying(50) DEFAULT 'low'::character varying,
    is_visible boolean DEFAULT true,
    started_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resolved_at timestamp without time zone,
    created_by integer,
    updated_by integer,
    metadata text
);


--
-- Name: incidents_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.incidents_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: incidents_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.incidents_id_seq OWNED BY public.incidents.id;


--
-- Name: monitors; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.monitors (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid,
    name character varying(255) NOT NULL,
    description text,
    type character varying(50) DEFAULT 'http'::character varying NOT NULL,
    target text NOT NULL,
    interval_seconds integer DEFAULT 60,
    timeout_seconds integer DEFAULT 30,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: password_resets; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_resets (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    user_id bigint NOT NULL,
    token text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    is_used boolean DEFAULT false
);


--
-- Name: password_resets_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.password_resets_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: password_resets_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.password_resets_id_seq OWNED BY public.password_resets.id;


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    name character varying(100) NOT NULL,
    display_name character varying(255) NOT NULL,
    description text,
    resource character varying(100) NOT NULL,
    action character varying(50) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: permissions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq OWNED BY public.permissions.id;


--
-- Name: plan_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.plan_features (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    plan_id uuid NOT NULL,
    feature_id integer NOT NULL,
    is_enabled boolean DEFAULT true,
    "order" integer DEFAULT 0,
    metadata text
);


--
-- Name: plan_features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.plan_features_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: plan_features_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.plan_features_id_seq OWNED BY public.plan_features.id;


--
-- Name: platforms; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.platforms (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    url text NOT NULL,
    description text,
    version text NOT NULL,
    status text DEFAULT 'active'::text,
    admin_email text NOT NULL,
    support_email text NOT NULL,
    metadata text
);


--
-- Name: pricing_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pricing_features (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    name character varying(255) NOT NULL,
    description text,
    category character varying(100) NOT NULL,
    icon text,
    is_active boolean DEFAULT true,
    "order" integer DEFAULT 0,
    metadata text
);


--
-- Name: pricing_features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pricing_features_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pricing_features_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pricing_features_id_seq OWNED BY public.pricing_features.id;


--
-- Name: pricing_tiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pricing_tiers (
    id integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    plan_id uuid NOT NULL,
    billing_interval character varying(20) NOT NULL,
    price numeric(10,2) NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying,
    discount_percent numeric(5,2) DEFAULT 0,
    is_active boolean DEFAULT true,
    metadata text
);


--
-- Name: pricing_tiers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pricing_tiers_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pricing_tiers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pricing_tiers_id_seq OWNED BY public.pricing_tiers.id;


--
-- Name: role_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_permissions (
    permission_id bigint NOT NULL,
    role_id bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    display_name character varying(255) NOT NULL,
    description text,
    is_default boolean DEFAULT false,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq OWNED BY public.roles.id;


--
-- Name: saas_admin_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_admin_users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    role text DEFAULT 'admin'::text,
    status text DEFAULT 'active'::text,
    last_login_at timestamp with time zone,
    metadata text
);


--
-- Name: saas_admin_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_admin_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_admin_users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_admin_users_id_seq OWNED BY public.saas_admin_users.id;


--
-- Name: saas_backups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_backups (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'pending'::text,
    size bigint,
    location text NOT NULL,
    checksum text,
    expires_at timestamp with time zone,
    metadata text
);


--
-- Name: saas_backups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_backups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_backups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_backups_id_seq OWNED BY public.saas_backups.id;


--
-- Name: saas_components; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_components (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id text NOT NULL,
    name text NOT NULL,
    description text,
    status text DEFAULT 'operational'::text,
    group_name text,
    "order" bigint DEFAULT 0,
    visible boolean DEFAULT true,
    show_uptime boolean DEFAULT true,
    uptime numeric DEFAULT 100,
    link text
);


--
-- Name: saas_feature_flags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_feature_flags (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    is_enabled boolean DEFAULT false,
    config text,
    metadata text
);


--
-- Name: saas_feature_flags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_feature_flags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_feature_flags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_feature_flags_id_seq OWNED BY public.saas_feature_flags.id;


--
-- Name: saas_features; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_features (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    slug text NOT NULL,
    description text,
    category text NOT NULL,
    is_enabled boolean DEFAULT true,
    is_public boolean DEFAULT true,
    config text,
    metadata text
);


--
-- Name: saas_features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_features_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_features_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_features_id_seq OWNED BY public.saas_features.id;


--
-- Name: saas_incidents; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_incidents (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    description text,
    status text NOT NULL,
    impact text NOT NULL,
    component_status text DEFAULT 'operational'::text,
    started_at timestamp with time zone NOT NULL,
    resolved_at timestamp with time zone,
    tenant_id text NOT NULL,
    created_by text,
    updated_by text,
    is_visible boolean DEFAULT true,
    external_id text,
    metadata text
);


--
-- Name: saas_integrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_integrations (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id text NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'active'::text,
    config text,
    last_used timestamp with time zone,
    last_error text,
    event_types text,
    enabled boolean DEFAULT true,
    description text
);


--
-- Name: saas_maintenance_windows; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_maintenance_windows (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    description text,
    status text NOT NULL,
    impact text NOT NULL,
    scheduled_for timestamp with time zone NOT NULL,
    estimated_end timestamp with time zone,
    actual_start timestamp with time zone,
    actual_end timestamp with time zone,
    tenant_id text NOT NULL,
    created_by text,
    updated_by text,
    is_visible boolean DEFAULT true,
    notify_subscribers boolean DEFAULT true,
    metadata text
);


--
-- Name: saas_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_notifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    title text NOT NULL,
    message text NOT NULL,
    type text NOT NULL,
    priority text DEFAULT 'normal'::text,
    status text DEFAULT 'active'::text,
    is_global boolean DEFAULT false,
    target_roles text,
    expires_at timestamp with time zone,
    metadata text
);


--
-- Name: saas_notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_notifications_id_seq OWNED BY public.saas_notifications.id;


--
-- Name: saas_plans; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_plans (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    name character varying(255) NOT NULL,
    slug character varying(255) NOT NULL,
    description text,
    price numeric(10,2) NOT NULL,
    currency character varying(3) DEFAULT 'USD'::character varying,
    billing_interval character varying(20) DEFAULT 'monthly'::character varying,
    max_tenants integer DEFAULT 1,
    max_users integer DEFAULT 5,
    max_services integer DEFAULT 10,
    max_monitors integer DEFAULT 50,
    max_subscribers integer DEFAULT 1000,
    max_incidents integer DEFAULT 100,
    max_maintenance integer DEFAULT 50,
    custom_domain boolean DEFAULT false,
    white_label boolean DEFAULT false,
    api boolean DEFAULT false,
    integrations boolean DEFAULT false,
    analytics boolean DEFAULT false,
    support character varying(50) DEFAULT 'email'::character varying,
    is_active boolean DEFAULT true,
    is_public boolean DEFAULT true,
    is_popular boolean DEFAULT false,
    button_text character varying(100) DEFAULT 'Get Started'::character varying,
    button_url character varying(255) DEFAULT '/signup'::character varying,
    display_order integer DEFAULT 0,
    features text,
    limits text,
    metadata text,
    max_status_pages integer DEFAULT 3,
    max_custom_domains integer DEFAULT 1,
    max_monitors_per_page integer DEFAULT 20,
    custom_status_pages boolean DEFAULT true,
    custom_domains boolean DEFAULT false,
    ssl_certificates boolean DEFAULT false,
    advanced_analytics boolean DEFAULT false,
    team_collaboration boolean DEFAULT false
);


--
-- Name: saas_services; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_services (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    status text NOT NULL,
    category text,
    "position" bigint DEFAULT 0,
    tenant_id text NOT NULL,
    is_visible boolean DEFAULT true,
    is_active boolean DEFAULT true,
    metadata text
);


--
-- Name: saas_stats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_stats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    total_tenants bigint,
    active_tenants bigint,
    total_users bigint,
    active_users bigint,
    total_revenue numeric,
    monthly_revenue numeric,
    total_plans bigint,
    active_plans bigint,
    total_features bigint,
    active_features bigint,
    last_updated timestamp with time zone
);


--
-- Name: saas_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_stats_id_seq OWNED BY public.saas_stats.id;


--
-- Name: saas_subscribers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_subscribers (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id text NOT NULL,
    email text NOT NULL,
    phone text,
    status text DEFAULT 'active'::text,
    event_types text,
    components text,
    verified_at timestamp with time zone,
    unsubscribe_token text,
    preferences text
);


--
-- Name: saas_webhooks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_webhooks (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id text NOT NULL,
    name text NOT NULL,
    url text NOT NULL,
    secret text,
    event_types text,
    headers text,
    method text DEFAULT 'POST'::text,
    status text DEFAULT 'active'::text,
    last_status bigint,
    last_error text,
    last_sent timestamp with time zone,
    enabled boolean DEFAULT true,
    retries bigint DEFAULT 3,
    timeout bigint DEFAULT 30
);


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id character varying(128) NOT NULL,
    user_id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    ip_address character varying(45),
    user_agent text,
    is_active boolean DEFAULT true,
    last_seen timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: status_page_domains; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.status_page_domains (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    status_page_id bigint NOT NULL,
    domain text NOT NULL,
    status bigint DEFAULT 1 NOT NULL,
    verification_method text NOT NULL,
    verification_token text NOT NULL,
    verified_at timestamp with time zone,
    is_primary boolean DEFAULT false,
    ssl_enabled boolean DEFAULT false,
    ssl_cert_issued_at timestamp with time zone,
    ssl_cert_expires_at timestamp with time zone,
    ssl_cert_issuer text,
    ssl_cert_common_name text,
    ssl_cert_raw text,
    metadata text
);


--
-- Name: status_page_domains_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.status_page_domains_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: status_page_domains_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.status_page_domains_id_seq OWNED BY public.status_page_domains.id;


--
-- Name: status_pages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.status_pages (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid,
    name character varying(255) NOT NULL,
    slug character varying(100) NOT NULL,
    description text,
    timezone character varying(50) DEFAULT 'UTC'::character varying,
    custom_css text,
    custom_js text,
    meta_title character varying(255),
    meta_description text,
    google_analytics_id character varying(50),
    is_public boolean DEFAULT true,
    is_default boolean DEFAULT false,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone
);


--
-- Name: teams; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.teams (
    id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    name character varying(100) NOT NULL,
    description text,
    is_active boolean DEFAULT true,
    created_by bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: teams_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.teams_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: teams_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.teams_id_seq OWNED BY public.teams.id;


--
-- Name: tenant_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    user_id bigint NOT NULL,
    action text NOT NULL,
    resource text NOT NULL,
    resource_id text,
    description text,
    ip_address text,
    user_agent text,
    metadata text
);


--
-- Name: tenant_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_activities_id_seq OWNED BY public.tenant_activities.id;


--
-- Name: tenant_admins; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_admins (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    user_id integer NOT NULL,
    role text NOT NULL,
    status text DEFAULT 'active'::text,
    permissions text,
    last_login_at timestamp with time zone,
    metadata text
);


--
-- Name: tenant_admins_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_admins_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_admins_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_admins_id_seq OWNED BY public.tenant_admins.id;


--
-- Name: tenant_audit; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_audit (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    action character varying(50) NOT NULL,
    slug character varying(255) NOT NULL,
    tenant_name character varying(255),
    changed_by character varying(255),
    changed_at timestamp without time zone DEFAULT now() NOT NULL,
    previous_state jsonb,
    new_state jsonb,
    metadata jsonb,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT tenant_audit_action_check CHECK (((action)::text = ANY ((ARRAY['created'::character varying, 'updated'::character varying, 'deleted'::character varying, 'restored'::character varying])::text[])))
);


--
-- Name: tenant_backups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_backups (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    name text NOT NULL,
    type text NOT NULL,
    status text DEFAULT 'pending'::text,
    size bigint,
    location text NOT NULL,
    checksum text,
    expires_at timestamp with time zone,
    metadata text
);


--
-- Name: tenant_backups_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_backups_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_backups_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_backups_id_seq OWNED BY public.tenant_backups.id;


--
-- Name: tenant_billing; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_billing (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    plan_id bigint NOT NULL,
    plan_name text NOT NULL,
    billing_cycle text DEFAULT 'monthly'::text,
    status text DEFAULT 'active'::text,
    current_period_start timestamp with time zone,
    current_period_end timestamp with time zone,
    trial_start timestamp with time zone,
    trial_end timestamp with time zone,
    is_trial_active boolean DEFAULT false,
    next_billing_date timestamp with time zone,
    amount numeric DEFAULT 0,
    currency text DEFAULT 'USD'::text,
    payment_method text,
    metadata text
);


--
-- Name: tenant_billing_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_billing_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_billing_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_billing_id_seq OWNED BY public.tenant_billing.id;


--
-- Name: tenant_feature_flags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_feature_flags (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    name text NOT NULL,
    description text,
    is_enabled boolean DEFAULT false,
    config text,
    metadata text
);


--
-- Name: tenant_feature_flags_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_feature_flags_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_feature_flags_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_feature_flags_id_seq OWNED BY public.tenant_feature_flags.id;


--
-- Name: tenant_notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_notifications (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    title text NOT NULL,
    message text NOT NULL,
    type text NOT NULL,
    priority text DEFAULT 'normal'::text,
    status text DEFAULT 'active'::text,
    is_read boolean DEFAULT false,
    read_at timestamp with time zone,
    expires_at timestamp with time zone,
    metadata text
);


--
-- Name: tenant_notifications_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_notifications_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_notifications_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_notifications_id_seq OWNED BY public.tenant_notifications.id;


--
-- Name: tenant_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_settings (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    settings text NOT NULL,
    version text DEFAULT '1.0.0'::text,
    status text DEFAULT 'active'::text,
    metadata text
);


--
-- Name: tenant_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_settings_id_seq OWNED BY public.tenant_settings.id;


--
-- Name: tenant_stats; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_stats (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    total_users bigint,
    active_users bigint,
    total_services bigint,
    active_services bigint,
    total_monitors bigint,
    active_monitors bigint,
    total_subscribers bigint,
    active_subscribers bigint,
    total_incidents bigint,
    open_incidents bigint,
    total_maintenance bigint,
    active_maintenance bigint,
    total_api_requests bigint,
    total_page_views bigint,
    storage_used bigint,
    bandwidth_used bigint,
    last_updated timestamp with time zone
);


--
-- Name: tenant_stats_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_stats_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_stats_id_seq OWNED BY public.tenant_stats.id;


--
-- Name: tenant_usage; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_usage (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id bigint NOT NULL,
    date timestamp with time zone NOT NULL,
    users_count bigint DEFAULT 0,
    services_count bigint DEFAULT 0,
    monitors_count bigint DEFAULT 0,
    subscribers_count bigint DEFAULT 0,
    incidents_count bigint DEFAULT 0,
    maintenance_count bigint DEFAULT 0,
    api_requests bigint DEFAULT 0,
    page_views bigint DEFAULT 0,
    storage_used bigint DEFAULT 0,
    bandwidth_used bigint DEFAULT 0,
    metadata text
);


--
-- Name: tenant_usage_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_usage_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_usage_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_usage_id_seq OWNED BY public.tenant_usage.id;


--
-- Name: tenants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(255) NOT NULL,
    slug character varying(100) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    domain character varying(255),
    subdomain character varying(255),
    contact_email character varying(255),
    billing_email character varying(255),
    plan_id uuid,
    status character varying(50) DEFAULT 'active'::character varying,
    is_active boolean DEFAULT true,
    settings text,
    branding text,
    features text,
    metadata text,
    deleted_at timestamp with time zone,
    max_users integer
);


--
-- Name: user_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    user_id bigint NOT NULL,
    action text NOT NULL,
    resource text,
    resource_id text,
    ip_address text,
    user_agent text,
    metadata text
);


--
-- Name: user_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_activities_id_seq OWNED BY public.user_activities.id;


--
-- Name: user_profiles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_profiles (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    user_id bigint NOT NULL,
    bio text,
    website text,
    location text,
    company text,
    job_title text,
    skills text,
    interests text
);


--
-- Name: user_profiles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_profiles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_profiles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_profiles_id_seq OWNED BY public.user_profiles.id;


--
-- Name: user_sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_sessions (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    user_id bigint NOT NULL,
    token text NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    ip_address text,
    user_agent text,
    is_active boolean DEFAULT true
);


--
-- Name: user_sessions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_sessions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_sessions_id_seq OWNED BY public.user_sessions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text NOT NULL,
    password_hash text NOT NULL,
    first_name text,
    last_name text,
    is_active boolean DEFAULT true,
    last_login_at timestamp with time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: audit_logs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs ALTER COLUMN id SET DEFAULT nextval('public.audit_logs_id_seq'::regclass);


--
-- Name: component_alerts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_alerts ALTER COLUMN id SET DEFAULT nextval('public.component_alerts_id_seq'::regclass);


--
-- Name: component_groups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_groups ALTER COLUMN id SET DEFAULT nextval('public.component_groups_id_seq'::regclass);


--
-- Name: component_history id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_history ALTER COLUMN id SET DEFAULT nextval('public.component_history_id_seq'::regclass);


--
-- Name: component_metrics id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_metrics ALTER COLUMN id SET DEFAULT nextval('public.component_metrics_id_seq'::regclass);


--
-- Name: component_statuses id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_statuses ALTER COLUMN id SET DEFAULT nextval('public.component_statuses_id_seq'::regclass);


--
-- Name: component_webhooks id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_webhooks ALTER COLUMN id SET DEFAULT nextval('public.component_webhooks_id_seq'::regclass);


--
-- Name: components id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.components ALTER COLUMN id SET DEFAULT nextval('public.components_id_seq'::regclass);


--
-- Name: email_verifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verifications ALTER COLUMN id SET DEFAULT nextval('public.email_verifications_id_seq'::regclass);


--
-- Name: incident_updates id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_updates ALTER COLUMN id SET DEFAULT nextval('public.incident_updates_id_seq'::regclass);


--
-- Name: incidents id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents ALTER COLUMN id SET DEFAULT nextval('public.incidents_id_seq'::regclass);


--
-- Name: password_resets id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_resets ALTER COLUMN id SET DEFAULT nextval('public.password_resets_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq'::regclass);


--
-- Name: plan_features id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features ALTER COLUMN id SET DEFAULT nextval('public.plan_features_id_seq'::regclass);


--
-- Name: pricing_features id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_features ALTER COLUMN id SET DEFAULT nextval('public.pricing_features_id_seq'::regclass);


--
-- Name: pricing_tiers id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_tiers ALTER COLUMN id SET DEFAULT nextval('public.pricing_tiers_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq'::regclass);


--
-- Name: saas_admin_users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_admin_users ALTER COLUMN id SET DEFAULT nextval('public.saas_admin_users_id_seq'::regclass);


--
-- Name: saas_backups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_backups ALTER COLUMN id SET DEFAULT nextval('public.saas_backups_id_seq'::regclass);


--
-- Name: saas_feature_flags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_feature_flags ALTER COLUMN id SET DEFAULT nextval('public.saas_feature_flags_id_seq'::regclass);


--
-- Name: saas_features id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_features ALTER COLUMN id SET DEFAULT nextval('public.saas_features_id_seq'::regclass);


--
-- Name: saas_notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_notifications ALTER COLUMN id SET DEFAULT nextval('public.saas_notifications_id_seq'::regclass);


--
-- Name: saas_stats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_stats ALTER COLUMN id SET DEFAULT nextval('public.saas_stats_id_seq'::regclass);


--
-- Name: status_page_domains id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_domains ALTER COLUMN id SET DEFAULT nextval('public.status_page_domains_id_seq'::regclass);


--
-- Name: teams id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams ALTER COLUMN id SET DEFAULT nextval('public.teams_id_seq'::regclass);


--
-- Name: tenant_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_activities ALTER COLUMN id SET DEFAULT nextval('public.tenant_activities_id_seq'::regclass);


--
-- Name: tenant_admins id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admins ALTER COLUMN id SET DEFAULT nextval('public.tenant_admins_id_seq'::regclass);


--
-- Name: tenant_backups id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_backups ALTER COLUMN id SET DEFAULT nextval('public.tenant_backups_id_seq'::regclass);


--
-- Name: tenant_billing id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_billing ALTER COLUMN id SET DEFAULT nextval('public.tenant_billing_id_seq'::regclass);


--
-- Name: tenant_feature_flags id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_feature_flags ALTER COLUMN id SET DEFAULT nextval('public.tenant_feature_flags_id_seq'::regclass);


--
-- Name: tenant_notifications id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_notifications ALTER COLUMN id SET DEFAULT nextval('public.tenant_notifications_id_seq'::regclass);


--
-- Name: tenant_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings ALTER COLUMN id SET DEFAULT nextval('public.tenant_settings_id_seq'::regclass);


--
-- Name: tenant_stats id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_stats ALTER COLUMN id SET DEFAULT nextval('public.tenant_stats_id_seq'::regclass);


--
-- Name: tenant_usage id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_usage ALTER COLUMN id SET DEFAULT nextval('public.tenant_usage_id_seq'::regclass);


--
-- Name: user_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_activities ALTER COLUMN id SET DEFAULT nextval('public.user_activities_id_seq'::regclass);


--
-- Name: user_profiles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profiles ALTER COLUMN id SET DEFAULT nextval('public.user_profiles_id_seq'::regclass);


--
-- Name: user_sessions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions ALTER COLUMN id SET DEFAULT nextval('public.user_sessions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: analytics_data analytics_data_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analytics_data
    ADD CONSTRAINT analytics_data_pkey PRIMARY KEY (id);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


--
-- Name: component_alerts component_alerts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_alerts
    ADD CONSTRAINT component_alerts_pkey PRIMARY KEY (id);


--
-- Name: component_groups component_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_groups
    ADD CONSTRAINT component_groups_pkey PRIMARY KEY (id);


--
-- Name: component_history component_history_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_history
    ADD CONSTRAINT component_history_pkey PRIMARY KEY (id);


--
-- Name: component_metrics component_metrics_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_metrics
    ADD CONSTRAINT component_metrics_pkey PRIMARY KEY (id);


--
-- Name: component_statuses component_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_statuses
    ADD CONSTRAINT component_statuses_pkey PRIMARY KEY (id);


--
-- Name: component_webhooks component_webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_webhooks
    ADD CONSTRAINT component_webhooks_pkey PRIMARY KEY (id);


--
-- Name: components components_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.components
    ADD CONSTRAINT components_pkey PRIMARY KEY (id);


--
-- Name: email_verifications email_verifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.email_verifications
    ADD CONSTRAINT email_verifications_pkey PRIMARY KEY (id);


--
-- Name: incident_updates incident_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incident_updates
    ADD CONSTRAINT incident_updates_pkey PRIMARY KEY (id);


--
-- Name: incidents incidents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.incidents
    ADD CONSTRAINT incidents_pkey PRIMARY KEY (id);


--
-- Name: monitors monitors_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monitors
    ADD CONSTRAINT monitors_pkey PRIMARY KEY (id);


--
-- Name: password_resets password_resets_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_resets
    ADD CONSTRAINT password_resets_pkey PRIMARY KEY (id);


--
-- Name: permissions permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey PRIMARY KEY (id);


--
-- Name: plan_features plan_features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT plan_features_pkey PRIMARY KEY (id);


--
-- Name: plan_features plan_features_plan_id_feature_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT plan_features_plan_id_feature_id_key UNIQUE (plan_id, feature_id);


--
-- Name: platforms platforms_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.platforms
    ADD CONSTRAINT platforms_pkey PRIMARY KEY (id);


--
-- Name: pricing_features pricing_features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_features
    ADD CONSTRAINT pricing_features_pkey PRIMARY KEY (id);


--
-- Name: pricing_tiers pricing_tiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_tiers
    ADD CONSTRAINT pricing_tiers_pkey PRIMARY KEY (id);


--
-- Name: role_permissions role_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT role_permissions_pkey PRIMARY KEY (permission_id, role_id);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id);


--
-- Name: saas_admin_users saas_admin_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_admin_users
    ADD CONSTRAINT saas_admin_users_pkey PRIMARY KEY (id);


--
-- Name: saas_backups saas_backups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_backups
    ADD CONSTRAINT saas_backups_pkey PRIMARY KEY (id);


--
-- Name: saas_components saas_components_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_components
    ADD CONSTRAINT saas_components_pkey PRIMARY KEY (id);


--
-- Name: saas_feature_flags saas_feature_flags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_feature_flags
    ADD CONSTRAINT saas_feature_flags_pkey PRIMARY KEY (id);


--
-- Name: saas_features saas_features_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_features
    ADD CONSTRAINT saas_features_pkey PRIMARY KEY (id);


--
-- Name: saas_incidents saas_incidents_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_incidents
    ADD CONSTRAINT saas_incidents_pkey PRIMARY KEY (id);


--
-- Name: saas_integrations saas_integrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_integrations
    ADD CONSTRAINT saas_integrations_pkey PRIMARY KEY (id);


--
-- Name: saas_maintenance_windows saas_maintenance_windows_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_maintenance_windows
    ADD CONSTRAINT saas_maintenance_windows_pkey PRIMARY KEY (id);


--
-- Name: saas_notifications saas_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_notifications
    ADD CONSTRAINT saas_notifications_pkey PRIMARY KEY (id);


--
-- Name: saas_plans saas_plans_name_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plans
    ADD CONSTRAINT saas_plans_name_key UNIQUE (name);


--
-- Name: saas_plans saas_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plans
    ADD CONSTRAINT saas_plans_pkey PRIMARY KEY (id);


--
-- Name: saas_plans saas_plans_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plans
    ADD CONSTRAINT saas_plans_slug_key UNIQUE (slug);


--
-- Name: saas_services saas_services_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_services
    ADD CONSTRAINT saas_services_pkey PRIMARY KEY (id);


--
-- Name: saas_stats saas_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_stats
    ADD CONSTRAINT saas_stats_pkey PRIMARY KEY (id);


--
-- Name: saas_subscribers saas_subscribers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_subscribers
    ADD CONSTRAINT saas_subscribers_pkey PRIMARY KEY (id);


--
-- Name: saas_webhooks saas_webhooks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_webhooks
    ADD CONSTRAINT saas_webhooks_pkey PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: status_page_domains status_page_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_domains
    ADD CONSTRAINT status_page_domains_pkey PRIMARY KEY (id);


--
-- Name: status_pages status_pages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_pages
    ADD CONSTRAINT status_pages_pkey PRIMARY KEY (id);


--
-- Name: teams teams_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.teams
    ADD CONSTRAINT teams_pkey PRIMARY KEY (id);


--
-- Name: tenant_activities tenant_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_activities
    ADD CONSTRAINT tenant_activities_pkey PRIMARY KEY (id);


--
-- Name: tenant_admins tenant_admins_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_admins
    ADD CONSTRAINT tenant_admins_pkey PRIMARY KEY (id);


--
-- Name: tenant_audit tenant_audit_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_audit
    ADD CONSTRAINT tenant_audit_pkey PRIMARY KEY (id);


--
-- Name: tenant_backups tenant_backups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_backups
    ADD CONSTRAINT tenant_backups_pkey PRIMARY KEY (id);


--
-- Name: tenant_billing tenant_billing_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_billing
    ADD CONSTRAINT tenant_billing_pkey PRIMARY KEY (id);


--
-- Name: tenant_feature_flags tenant_feature_flags_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_feature_flags
    ADD CONSTRAINT tenant_feature_flags_pkey PRIMARY KEY (id);


--
-- Name: tenant_notifications tenant_notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_notifications
    ADD CONSTRAINT tenant_notifications_pkey PRIMARY KEY (id);


--
-- Name: tenant_settings tenant_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_settings
    ADD CONSTRAINT tenant_settings_pkey PRIMARY KEY (id);


--
-- Name: tenant_stats tenant_stats_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_stats
    ADD CONSTRAINT tenant_stats_pkey PRIMARY KEY (id);


--
-- Name: tenant_usage tenant_usage_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_usage
    ADD CONSTRAINT tenant_usage_pkey PRIMARY KEY (id);


--
-- Name: tenants tenants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenants
    ADD CONSTRAINT tenants_pkey PRIMARY KEY (id);


--
-- Name: user_activities user_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_activities
    ADD CONSTRAINT user_activities_pkey PRIMARY KEY (id);


--
-- Name: user_profiles user_profiles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_profiles
    ADD CONSTRAINT user_profiles_pkey PRIMARY KEY (id);


--
-- Name: user_sessions user_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_sessions
    ADD CONSTRAINT user_sessions_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: idx_audit_logs_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_resource_id ON public.audit_logs USING btree (resource_id);


--
-- Name: idx_audit_logs_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_tenant_id ON public.audit_logs USING btree (tenant_id);


--
-- Name: idx_audit_logs_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_audit_logs_user_id ON public.audit_logs USING btree (user_id);


--
-- Name: idx_component_alerts_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_alerts_component_id ON public.component_alerts USING btree (component_id);


--
-- Name: idx_component_alerts_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_alerts_deleted_at ON public.component_alerts USING btree (deleted_at);


--
-- Name: idx_component_groups_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_groups_deleted_at ON public.component_groups USING btree (deleted_at);


--
-- Name: idx_component_groups_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_groups_tenant_id ON public.component_groups USING btree (tenant_id);


--
-- Name: idx_component_history_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_history_component_id ON public.component_history USING btree (component_id);


--
-- Name: idx_component_history_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_history_deleted_at ON public.component_history USING btree (deleted_at);


--
-- Name: idx_component_metrics_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_metrics_component_id ON public.component_metrics USING btree (component_id);


--
-- Name: idx_component_metrics_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_metrics_deleted_at ON public.component_metrics USING btree (deleted_at);


--
-- Name: idx_component_metrics_timestamp; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_metrics_timestamp ON public.component_metrics USING btree ("timestamp");


--
-- Name: idx_component_statuses_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_statuses_component_id ON public.component_statuses USING btree (component_id);


--
-- Name: idx_component_statuses_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_statuses_deleted_at ON public.component_statuses USING btree (deleted_at);


--
-- Name: idx_component_webhooks_component_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_webhooks_component_id ON public.component_webhooks USING btree (component_id);


--
-- Name: idx_component_webhooks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_component_webhooks_deleted_at ON public.component_webhooks USING btree (deleted_at);


--
-- Name: idx_components_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_components_deleted_at ON public.components USING btree (deleted_at);


--
-- Name: idx_components_group_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_components_group_id ON public.components USING btree (group_id);


--
-- Name: idx_components_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_components_tenant_id ON public.components USING btree (tenant_id);


--
-- Name: idx_email_verifications_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_email_verifications_token ON public.email_verifications USING btree (token);


--
-- Name: idx_email_verifications_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_email_verifications_user_id ON public.email_verifications USING btree (user_id);


--
-- Name: idx_incident_updates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_updates_deleted_at ON public.incident_updates USING btree (deleted_at);


--
-- Name: idx_incident_updates_incident_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incident_updates_incident_id ON public.incident_updates USING btree (incident_id);


--
-- Name: idx_incidents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_deleted_at ON public.incidents USING btree (deleted_at);


--
-- Name: idx_incidents_resolved_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_resolved_at ON public.incidents USING btree (resolved_at);


--
-- Name: idx_incidents_started_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_started_at ON public.incidents USING btree (started_at);


--
-- Name: idx_incidents_status_impact; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_status_impact ON public.incidents USING btree (status, impact);


--
-- Name: idx_incidents_tenant_visible; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_incidents_tenant_visible ON public.incidents USING btree (tenant_id, is_visible);


--
-- Name: idx_password_resets_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_password_resets_token ON public.password_resets USING btree (token);


--
-- Name: idx_password_resets_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_password_resets_user_id ON public.password_resets USING btree (user_id);


--
-- Name: idx_permissions_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_permissions_name ON public.permissions USING btree (name);


--
-- Name: idx_plan_features_feature_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_feature_id ON public.plan_features USING btree (feature_id);


--
-- Name: idx_plan_features_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_plan_id ON public.plan_features USING btree (plan_id);


--
-- Name: idx_platforms_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_platforms_deleted_at ON public.platforms USING btree (deleted_at);


--
-- Name: idx_platforms_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_platforms_name ON public.platforms USING btree (name);


--
-- Name: idx_pricing_features_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_features_category ON public.pricing_features USING btree (category);


--
-- Name: idx_pricing_features_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_features_is_active ON public.pricing_features USING btree (is_active);


--
-- Name: idx_pricing_tiers_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_tiers_plan_id ON public.pricing_tiers USING btree (plan_id);


--
-- Name: idx_roles_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_tenant_id ON public.roles USING btree (tenant_id);


--
-- Name: idx_saas_admin_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_admin_users_deleted_at ON public.saas_admin_users USING btree (deleted_at);


--
-- Name: idx_saas_admin_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_admin_users_email ON public.saas_admin_users USING btree (email);


--
-- Name: idx_saas_admin_users_username; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_admin_users_username ON public.saas_admin_users USING btree (username);


--
-- Name: idx_saas_backups_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_backups_deleted_at ON public.saas_backups USING btree (deleted_at);


--
-- Name: idx_saas_backups_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_backups_type ON public.saas_backups USING btree (type);


--
-- Name: idx_saas_components_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_components_deleted_at ON public.saas_components USING btree (deleted_at);


--
-- Name: idx_saas_components_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_components_tenant_id ON public.saas_components USING btree (tenant_id);


--
-- Name: idx_saas_feature_flags_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_feature_flags_deleted_at ON public.saas_feature_flags USING btree (deleted_at);


--
-- Name: idx_saas_feature_flags_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_feature_flags_name ON public.saas_feature_flags USING btree (name);


--
-- Name: idx_saas_features_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_features_category ON public.saas_features USING btree (category);


--
-- Name: idx_saas_features_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_features_deleted_at ON public.saas_features USING btree (deleted_at);


--
-- Name: idx_saas_features_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_features_name ON public.saas_features USING btree (name);


--
-- Name: idx_saas_features_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_features_slug ON public.saas_features USING btree (slug);


--
-- Name: idx_saas_incidents_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incidents_deleted_at ON public.saas_incidents USING btree (deleted_at);


--
-- Name: idx_saas_incidents_external_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incidents_external_id ON public.saas_incidents USING btree (external_id);


--
-- Name: idx_saas_incidents_impact; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incidents_impact ON public.saas_incidents USING btree (impact);


--
-- Name: idx_saas_incidents_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incidents_status ON public.saas_incidents USING btree (status);


--
-- Name: idx_saas_incidents_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incidents_tenant_id ON public.saas_incidents USING btree (tenant_id);


--
-- Name: idx_saas_integrations_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_integrations_deleted_at ON public.saas_integrations USING btree (deleted_at);


--
-- Name: idx_saas_integrations_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_integrations_tenant_id ON public.saas_integrations USING btree (tenant_id);


--
-- Name: idx_saas_integrations_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_integrations_type ON public.saas_integrations USING btree (type);


--
-- Name: idx_saas_maintenance_windows_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_maintenance_windows_deleted_at ON public.saas_maintenance_windows USING btree (deleted_at);


--
-- Name: idx_saas_maintenance_windows_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_maintenance_windows_status ON public.saas_maintenance_windows USING btree (status);


--
-- Name: idx_saas_maintenance_windows_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_maintenance_windows_tenant_id ON public.saas_maintenance_windows USING btree (tenant_id);


--
-- Name: idx_saas_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_notifications_deleted_at ON public.saas_notifications USING btree (deleted_at);


--
-- Name: idx_saas_notifications_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_notifications_type ON public.saas_notifications USING btree (type);


--
-- Name: idx_saas_plans_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_plans_is_active ON public.saas_plans USING btree (is_active);


--
-- Name: idx_saas_plans_is_popular; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_plans_is_popular ON public.saas_plans USING btree (is_popular);


--
-- Name: idx_saas_plans_is_public; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_plans_is_public ON public.saas_plans USING btree (is_public);


--
-- Name: idx_saas_services_category; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_services_category ON public.saas_services USING btree (category);


--
-- Name: idx_saas_services_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_services_deleted_at ON public.saas_services USING btree (deleted_at);


--
-- Name: idx_saas_services_status; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_services_status ON public.saas_services USING btree (status);


--
-- Name: idx_saas_services_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_services_tenant_id ON public.saas_services USING btree (tenant_id);


--
-- Name: idx_saas_subscribers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_subscribers_deleted_at ON public.saas_subscribers USING btree (deleted_at);


--
-- Name: idx_saas_subscribers_email; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_subscribers_email ON public.saas_subscribers USING btree (email);


--
-- Name: idx_saas_subscribers_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_subscribers_tenant_id ON public.saas_subscribers USING btree (tenant_id);


--
-- Name: idx_saas_subscribers_unsubscribe_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_subscribers_unsubscribe_token ON public.saas_subscribers USING btree (unsubscribe_token);


--
-- Name: idx_saas_webhooks_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_webhooks_deleted_at ON public.saas_webhooks USING btree (deleted_at);


--
-- Name: idx_saas_webhooks_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_webhooks_tenant_id ON public.saas_webhooks USING btree (tenant_id);


--
-- Name: idx_sessions_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_tenant_id ON public.sessions USING btree (tenant_id);


--
-- Name: idx_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_sessions_user_id ON public.sessions USING btree (user_id);


--
-- Name: idx_status_page_domains_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_domains_deleted_at ON public.status_page_domains USING btree (deleted_at);


--
-- Name: idx_status_page_domains_domain; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_status_page_domains_domain ON public.status_page_domains USING btree (domain);


--
-- Name: idx_status_page_domains_status_page_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_domains_status_page_id ON public.status_page_domains USING btree (status_page_id);


--
-- Name: idx_status_page_domains_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_domains_tenant_id ON public.status_page_domains USING btree (tenant_id);


--
-- Name: idx_teams_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_teams_tenant_id ON public.teams USING btree (tenant_id);


--
-- Name: idx_tenant_activities_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_action ON public.tenant_activities USING btree (action);


--
-- Name: idx_tenant_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_deleted_at ON public.tenant_activities USING btree (deleted_at);


--
-- Name: idx_tenant_activities_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_resource ON public.tenant_activities USING btree (resource);


--
-- Name: idx_tenant_activities_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_resource_id ON public.tenant_activities USING btree (resource_id);


--
-- Name: idx_tenant_activities_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_tenant_id ON public.tenant_activities USING btree (tenant_id);


--
-- Name: idx_tenant_activities_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_activities_user_id ON public.tenant_activities USING btree (user_id);


--
-- Name: idx_tenant_admins_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_admins_deleted_at ON public.tenant_admins USING btree (deleted_at);


--
-- Name: idx_tenant_admins_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_admins_tenant_id ON public.tenant_admins USING btree (tenant_id);


--
-- Name: idx_tenant_admins_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_admins_user_id ON public.tenant_admins USING btree (user_id);


--
-- Name: idx_tenant_audit_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_audit_action ON public.tenant_audit USING btree (action);


--
-- Name: idx_tenant_audit_changed_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_audit_changed_at ON public.tenant_audit USING btree (changed_at DESC);


--
-- Name: idx_tenant_audit_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_audit_slug ON public.tenant_audit USING btree (slug);


--
-- Name: idx_tenant_audit_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_audit_tenant_id ON public.tenant_audit USING btree (tenant_id);


--
-- Name: idx_tenant_backups_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_backups_deleted_at ON public.tenant_backups USING btree (deleted_at);


--
-- Name: idx_tenant_backups_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_backups_tenant_id ON public.tenant_backups USING btree (tenant_id);


--
-- Name: idx_tenant_backups_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_backups_type ON public.tenant_backups USING btree (type);


--
-- Name: idx_tenant_billing_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_billing_deleted_at ON public.tenant_billing USING btree (deleted_at);


--
-- Name: idx_tenant_billing_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_billing_plan_id ON public.tenant_billing USING btree (plan_id);


--
-- Name: idx_tenant_billing_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_billing_tenant_id ON public.tenant_billing USING btree (tenant_id);


--
-- Name: idx_tenant_feature_flags_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_feature_flags_deleted_at ON public.tenant_feature_flags USING btree (deleted_at);


--
-- Name: idx_tenant_feature_flags_name; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_feature_flags_name ON public.tenant_feature_flags USING btree (name);


--
-- Name: idx_tenant_feature_flags_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_feature_flags_tenant_id ON public.tenant_feature_flags USING btree (tenant_id);


--
-- Name: idx_tenant_notifications_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_notifications_deleted_at ON public.tenant_notifications USING btree (deleted_at);


--
-- Name: idx_tenant_notifications_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_notifications_tenant_id ON public.tenant_notifications USING btree (tenant_id);


--
-- Name: idx_tenant_notifications_type; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_notifications_type ON public.tenant_notifications USING btree (type);


--
-- Name: idx_tenant_settings_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_settings_deleted_at ON public.tenant_settings USING btree (deleted_at);


--
-- Name: idx_tenant_settings_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_settings_tenant_id ON public.tenant_settings USING btree (tenant_id);


--
-- Name: idx_tenant_stats_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_stats_tenant_id ON public.tenant_stats USING btree (tenant_id);


--
-- Name: idx_tenant_usage_date; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_usage_date ON public.tenant_usage USING btree (date);


--
-- Name: idx_tenant_usage_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_usage_deleted_at ON public.tenant_usage USING btree (deleted_at);


--
-- Name: idx_tenant_usage_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_usage_tenant_id ON public.tenant_usage USING btree (tenant_id);


--
-- Name: idx_tenants_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenants_deleted_at ON public.tenants USING btree (deleted_at);


--
-- Name: idx_tenants_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenants_plan_id ON public.tenants USING btree (plan_id);


--
-- Name: idx_tenants_slug_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenants_slug_active ON public.tenants USING btree (slug) WHERE (deleted_at IS NULL);


--
-- Name: idx_user_activities_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_activities_user_id ON public.user_activities USING btree (user_id);


--
-- Name: idx_user_profiles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_profiles_user_id ON public.user_profiles USING btree (user_id);


--
-- Name: idx_user_sessions_token; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_user_sessions_token ON public.user_sessions USING btree (token);


--
-- Name: idx_user_sessions_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_sessions_user_id ON public.user_sessions USING btree (user_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: analytics_data analytics_data_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.analytics_data
    ADD CONSTRAINT analytics_data_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: component_alerts fk_component_alerts_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_alerts
    ADD CONSTRAINT fk_component_alerts_component FOREIGN KEY (component_id) REFERENCES public.components(id);


--
-- Name: components fk_component_groups_components; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.components
    ADD CONSTRAINT fk_component_groups_components FOREIGN KEY (group_id) REFERENCES public.component_groups(id);


--
-- Name: component_history fk_component_history_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_history
    ADD CONSTRAINT fk_component_history_component FOREIGN KEY (component_id) REFERENCES public.components(id);


--
-- Name: component_metrics fk_component_metrics_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_metrics
    ADD CONSTRAINT fk_component_metrics_component FOREIGN KEY (component_id) REFERENCES public.components(id);


--
-- Name: component_statuses fk_component_statuses_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_statuses
    ADD CONSTRAINT fk_component_statuses_component FOREIGN KEY (component_id) REFERENCES public.components(id);


--
-- Name: component_webhooks fk_component_webhooks_component; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.component_webhooks
    ADD CONSTRAINT fk_component_webhooks_component FOREIGN KEY (component_id) REFERENCES public.components(id);


--
-- Name: role_permissions fk_role_permissions_permission; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_permission FOREIGN KEY (permission_id) REFERENCES public.permissions(id);


--
-- Name: role_permissions fk_role_permissions_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_permissions
    ADD CONSTRAINT fk_role_permissions_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: monitors monitors_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.monitors
    ADD CONSTRAINT monitors_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- Name: plan_features plan_features_feature_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT plan_features_feature_id_fkey FOREIGN KEY (feature_id) REFERENCES public.pricing_features(id) ON DELETE CASCADE;


--
-- Name: plan_features plan_features_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT plan_features_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.saas_plans(id) ON DELETE CASCADE;


--
-- Name: pricing_tiers pricing_tiers_plan_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_tiers
    ADD CONSTRAINT pricing_tiers_plan_id_fkey FOREIGN KEY (plan_id) REFERENCES public.saas_plans(id) ON DELETE CASCADE;


--
-- Name: status_pages status_pages_tenant_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_pages
    ADD CONSTRAINT status_pages_tenant_id_fkey FOREIGN KEY (tenant_id) REFERENCES public.tenants(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

