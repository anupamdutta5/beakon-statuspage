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


SET default_tablespace = '';

SET default_table_access_method = heap;

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
-- Name: domain_verification_records; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.domain_verification_records (
    type text,
    name text,
    value text
);


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
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    plan_id uuid NOT NULL,
    feature_id bigint NOT NULL,
    is_enabled boolean DEFAULT true,
    "order" bigint DEFAULT 0,
    metadata text
);


--
-- Name: plan_features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.plan_features_id_seq
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
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    description text,
    category text NOT NULL,
    icon text,
    is_active boolean DEFAULT true,
    "order" bigint DEFAULT 0,
    metadata text
);


--
-- Name: pricing_features_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pricing_features_id_seq
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
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    plan_id uuid NOT NULL,
    billing_interval text NOT NULL,
    price numeric NOT NULL,
    currency text DEFAULT 'USD'::text,
    discount_percent numeric DEFAULT 0,
    is_active boolean DEFAULT true,
    metadata text
);


--
-- Name: pricing_tiers_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pricing_tiers_id_seq
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
-- Name: saas_activities; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_activities (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
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
-- Name: saas_activities_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.saas_activities_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: saas_activities_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.saas_activities_id_seq OWNED BY public.saas_activities.id;


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
-- Name: saas_incident_updates; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.saas_incident_updates (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    incident_id uuid NOT NULL,
    status text NOT NULL,
    message text NOT NULL,
    created_by text,
    is_public boolean DEFAULT true,
    update_type text DEFAULT 'status_update'::text,
    metadata text
);


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
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    slug text NOT NULL,
    description text,
    price numeric NOT NULL,
    currency text DEFAULT 'USD'::text,
    billing_interval text DEFAULT 'monthly'::text,
    max_tenants bigint DEFAULT 1,
    max_users bigint DEFAULT 5,
    max_services bigint DEFAULT 10,
    max_monitors bigint DEFAULT 50,
    max_subscribers bigint DEFAULT 1000,
    max_incidents bigint DEFAULT 100,
    max_maintenance bigint DEFAULT 50,
    custom_domain boolean DEFAULT false,
    white_label boolean DEFAULT false,
    api boolean DEFAULT false,
    integrations boolean DEFAULT false,
    analytics boolean DEFAULT false,
    support text DEFAULT 'email'::text,
    is_active boolean DEFAULT true,
    is_public boolean DEFAULT true,
    is_popular boolean DEFAULT false,
    button_text text DEFAULT 'Get Started'::text,
    button_url text DEFAULT '/signup'::text,
    display_order bigint DEFAULT 0,
    features text,
    limits text,
    metadata text
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
    ip_address character varying(45),
    user_agent text,
    is_active boolean DEFAULT true,
    last_seen timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    tenant_id uuid NOT NULL
);


--
-- Name: status_page_configs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.status_page_configs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    status_page_id bigint NOT NULL,
    site_name text NOT NULL,
    site_description text,
    logo_url text,
    favicon text,
    og_image text,
    site_url text,
    subscribe_url text,
    history_url text,
    api_url text,
    web_socket_url text,
    powered_by text,
    powered_by_url text,
    theme text DEFAULT 'default'::text,
    custom_css text,
    metadata text
);


--
-- Name: status_page_configs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.status_page_configs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: status_page_configs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.status_page_configs_id_seq OWNED BY public.status_page_configs.id;


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
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    name text NOT NULL,
    slug text NOT NULL,
    title text NOT NULL,
    description text,
    domain text,
    is_public boolean DEFAULT true,
    is_active boolean DEFAULT true,
    metadata text
);


--
-- Name: status_pages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.status_pages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: status_pages_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.status_pages_id_seq OWNED BY public.status_pages.id;


--
-- Name: team_members; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_members (
    id bigint NOT NULL,
    team_id bigint NOT NULL,
    user_id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    role character varying(50) DEFAULT 'member'::character varying NOT NULL,
    added_by bigint NOT NULL,
    added_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: team_members_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.team_members_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: team_members_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.team_members_id_seq OWNED BY public.team_members.id;


--
-- Name: team_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.team_roles (
    id bigint NOT NULL,
    team_id bigint NOT NULL,
    role_id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    assigned_by bigint NOT NULL,
    assigned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: team_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.team_roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: team_roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.team_roles_id_seq OWNED BY public.team_roles.id;


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
    tenant_id uuid NOT NULL,
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
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    user_id bigint NOT NULL,
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
    tenant_id uuid NOT NULL,
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
-- Name: tenant_branding; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_branding (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
    logo_url text,
    favicon_url text,
    primary_color text,
    secondary_color text,
    accent_color text,
    background_color text,
    text_color text,
    font_family text,
    font_size text,
    font_weight text,
    layout text,
    show_logo boolean,
    show_footer boolean,
    footer_text text,
    custom_header text,
    custom_footer text,
    meta_title text,
    meta_description text,
    meta_keywords text
);


--
-- Name: tenant_branding_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tenant_branding_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tenant_branding_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tenant_branding_id_seq OWNED BY public.tenant_branding.id;


--
-- Name: tenant_feature_flags; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tenant_feature_flags (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    tenant_id uuid NOT NULL,
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
    tenant_id uuid NOT NULL,
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
    tenant_id uuid NOT NULL,
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
    tenant_id uuid NOT NULL,
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
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    name text NOT NULL,
    slug text NOT NULL,
    domain text,
    subdomain text,
    contact_email text NOT NULL,
    billing_email text,
    status text DEFAULT 'active'::text,
    is_active boolean DEFAULT true,
    max_users bigint,
    settings text,
    branding text,
    features text,
    plan_id uuid,
    metadata text,
    CONSTRAINT chk_tenants_max_users CHECK (((max_users IS NULL) OR (max_users > 0)))
);


--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.user_roles (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    role_id bigint NOT NULL,
    tenant_id bigint NOT NULL,
    assigned_by bigint NOT NULL,
    assigned_at timestamp with time zone DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp with time zone,
    is_active boolean DEFAULT true,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


--
-- Name: user_roles_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.user_roles_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: user_roles_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.user_roles_id_seq OWNED BY public.user_roles.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    email text NOT NULL,
    password_hash text NOT NULL,
    first_name text,
    last_name text,
    is_active boolean DEFAULT true,
    last_login_at timestamp with time zone,
    tenant_id uuid,
    role character varying(50) DEFAULT 'admin'::character varying NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
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
-- Name: saas_activities id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_activities ALTER COLUMN id SET DEFAULT nextval('public.saas_activities_id_seq'::regclass);


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
-- Name: status_page_configs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_configs ALTER COLUMN id SET DEFAULT nextval('public.status_page_configs_id_seq'::regclass);


--
-- Name: status_page_domains id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_domains ALTER COLUMN id SET DEFAULT nextval('public.status_page_domains_id_seq'::regclass);


--
-- Name: status_pages id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_pages ALTER COLUMN id SET DEFAULT nextval('public.status_pages_id_seq'::regclass);


--
-- Name: team_members id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members ALTER COLUMN id SET DEFAULT nextval('public.team_members_id_seq'::regclass);


--
-- Name: team_roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_roles ALTER COLUMN id SET DEFAULT nextval('public.team_roles_id_seq'::regclass);


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
-- Name: tenant_branding id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_branding ALTER COLUMN id SET DEFAULT nextval('public.tenant_branding_id_seq'::regclass);


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
-- Name: user_roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles ALTER COLUMN id SET DEFAULT nextval('public.user_roles_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Name: audit_logs audit_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.audit_logs
    ADD CONSTRAINT audit_logs_pkey PRIMARY KEY (id);


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
-- Name: saas_activities saas_activities_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_activities
    ADD CONSTRAINT saas_activities_pkey PRIMARY KEY (id);


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
-- Name: saas_incident_updates saas_incident_updates_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_incident_updates
    ADD CONSTRAINT saas_incident_updates_pkey PRIMARY KEY (id);


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
-- Name: saas_plans saas_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_plans
    ADD CONSTRAINT saas_plans_pkey PRIMARY KEY (id);


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
-- Name: status_page_configs status_page_configs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_configs
    ADD CONSTRAINT status_page_configs_pkey PRIMARY KEY (id);


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
-- Name: team_members team_members_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT team_members_pkey PRIMARY KEY (id);


--
-- Name: team_roles team_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_roles
    ADD CONSTRAINT team_roles_pkey PRIMARY KEY (id);


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
-- Name: tenant_branding tenant_branding_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_branding
    ADD CONSTRAINT tenant_branding_pkey PRIMARY KEY (id);


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
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


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
-- Name: idx_permissions_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_permissions_name ON public.permissions USING btree (name);


--
-- Name: idx_plan_features_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_deleted_at ON public.plan_features USING btree (deleted_at);


--
-- Name: idx_plan_features_feature_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_feature_id ON public.plan_features USING btree (feature_id);


--
-- Name: idx_plan_features_is_enabled; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_plan_features_is_enabled ON public.plan_features USING btree (is_enabled);


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
-- Name: idx_pricing_features_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_features_deleted_at ON public.pricing_features USING btree (deleted_at);


--
-- Name: idx_pricing_features_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_features_is_active ON public.pricing_features USING btree (is_active);


--
-- Name: idx_pricing_tiers_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_tiers_deleted_at ON public.pricing_tiers USING btree (deleted_at);


--
-- Name: idx_pricing_tiers_is_active; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_tiers_is_active ON public.pricing_tiers USING btree (is_active);


--
-- Name: idx_pricing_tiers_plan_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_pricing_tiers_plan_id ON public.pricing_tiers USING btree (plan_id);


--
-- Name: idx_roles_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_roles_tenant_id ON public.roles USING btree (tenant_id);


--
-- Name: idx_saas_activities_action; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_activities_action ON public.saas_activities USING btree (action);


--
-- Name: idx_saas_activities_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_activities_deleted_at ON public.saas_activities USING btree (deleted_at);


--
-- Name: idx_saas_activities_resource; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_activities_resource ON public.saas_activities USING btree (resource);


--
-- Name: idx_saas_activities_resource_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_activities_resource_id ON public.saas_activities USING btree (resource_id);


--
-- Name: idx_saas_activities_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_activities_user_id ON public.saas_activities USING btree (user_id);


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
-- Name: idx_saas_incident_updates_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incident_updates_deleted_at ON public.saas_incident_updates USING btree (deleted_at);


--
-- Name: idx_saas_incident_updates_incident_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_incident_updates_incident_id ON public.saas_incident_updates USING btree (incident_id);


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
-- Name: idx_saas_plans_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_saas_plans_deleted_at ON public.saas_plans USING btree (deleted_at);


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
-- Name: idx_saas_plans_name; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_plans_name ON public.saas_plans USING btree (name);


--
-- Name: idx_saas_plans_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_saas_plans_slug ON public.saas_plans USING btree (slug);


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
-- Name: idx_status_page_configs_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_configs_deleted_at ON public.status_page_configs USING btree (deleted_at);


--
-- Name: idx_status_page_configs_status_page_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_page_configs_status_page_id ON public.status_page_configs USING btree (status_page_id);


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
-- Name: idx_status_pages_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_pages_deleted_at ON public.status_pages USING btree (deleted_at);


--
-- Name: idx_status_pages_slug; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_status_pages_slug ON public.status_pages USING btree (slug);


--
-- Name: idx_status_pages_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_status_pages_tenant_id ON public.status_pages USING btree (tenant_id);


--
-- Name: idx_team_members_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_team_id ON public.team_members USING btree (team_id);


--
-- Name: idx_team_members_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_tenant_id ON public.team_members USING btree (tenant_id);


--
-- Name: idx_team_members_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_members_user_id ON public.team_members USING btree (user_id);


--
-- Name: idx_team_roles_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_roles_role_id ON public.team_roles USING btree (role_id);


--
-- Name: idx_team_roles_team_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_roles_team_id ON public.team_roles USING btree (team_id);


--
-- Name: idx_team_roles_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_team_roles_tenant_id ON public.team_roles USING btree (tenant_id);


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
-- Name: idx_tenant_admins_role; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_admins_role ON public.tenant_admins USING btree (role);


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
-- Name: idx_tenant_branding_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_tenant_branding_deleted_at ON public.tenant_branding USING btree (deleted_at);


--
-- Name: idx_tenant_branding_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_branding_tenant_id ON public.tenant_branding USING btree (tenant_id);


--
-- Name: idx_tenant_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenant_email ON public.users USING btree (tenant_id, email) WHERE (deleted_at IS NULL);


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
-- Name: idx_tenants_domain; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenants_domain ON public.tenants USING btree (domain);


--
-- Name: idx_tenants_slug_active; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenants_slug_active ON public.tenants USING btree (slug) WHERE (deleted_at IS NULL);


--
-- Name: idx_tenants_subdomain; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_tenants_subdomain ON public.tenants USING btree (subdomain);


--
-- Name: idx_user_roles_role_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_roles_role_id ON public.user_roles USING btree (role_id);


--
-- Name: idx_user_roles_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_roles_tenant_id ON public.user_roles USING btree (tenant_id);


--
-- Name: idx_user_roles_user_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_user_roles_user_id ON public.user_roles USING btree (user_id);


--
-- Name: idx_users_deleted_at; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_deleted_at ON public.users USING btree (deleted_at);


--
-- Name: idx_users_email; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX idx_users_email ON public.users USING btree (email);


--
-- Name: idx_users_tenant_id; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX idx_users_tenant_id ON public.users USING btree (tenant_id);


--
-- Name: plan_features fk_plan_features_feature; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT fk_plan_features_feature FOREIGN KEY (feature_id) REFERENCES public.pricing_features(id);


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
-- Name: user_roles fk_roles_user_roles; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT fk_roles_user_roles FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: saas_activities fk_saas_activities_user; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_activities
    ADD CONSTRAINT fk_saas_activities_user FOREIGN KEY (user_id) REFERENCES public.saas_admin_users(id);


--
-- Name: saas_incident_updates fk_saas_incidents_updates; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.saas_incident_updates
    ADD CONSTRAINT fk_saas_incidents_updates FOREIGN KEY (incident_id) REFERENCES public.saas_incidents(id);


--
-- Name: plan_features fk_saas_plans_plan_features; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.plan_features
    ADD CONSTRAINT fk_saas_plans_plan_features FOREIGN KEY (plan_id) REFERENCES public.saas_plans(id);


--
-- Name: pricing_tiers fk_saas_plans_pricing_tiers; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pricing_tiers
    ADD CONSTRAINT fk_saas_plans_pricing_tiers FOREIGN KEY (plan_id) REFERENCES public.saas_plans(id);


--
-- Name: status_page_configs fk_status_page_configs_status_page; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.status_page_configs
    ADD CONSTRAINT fk_status_page_configs_status_page FOREIGN KEY (status_page_id) REFERENCES public.status_pages(id);


--
-- Name: team_roles fk_team_roles_role; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_roles
    ADD CONSTRAINT fk_team_roles_role FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: team_members fk_teams_team_members; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_members
    ADD CONSTRAINT fk_teams_team_members FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: team_roles fk_teams_team_roles; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.team_roles
    ADD CONSTRAINT fk_teams_team_roles FOREIGN KEY (team_id) REFERENCES public.teams(id);


--
-- Name: tenant_branding fk_tenant_branding_tenant; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tenant_branding
    ADD CONSTRAINT fk_tenant_branding_tenant FOREIGN KEY (tenant_id) REFERENCES public.tenants(id);


--
-- PostgreSQL database dump complete
--

