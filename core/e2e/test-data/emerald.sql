--
-- PostgreSQL database dump
--

\restrict lwgGQZ4PFEziDh5zTfrRYXaeWD4ZEyHDnIMPXT3GIqJci3Uia2Fi2N6QmWlmb6L

-- Dumped from database version 17.10
-- Dumped by pg_dump version 17.10

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: timescaledb; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;


--
-- Name: EXTENSION timescaledb; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON EXTENSION timescaledb IS 'Enables scalable inserts and complex queries for time-series data (Community Edition)';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: allergens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.allergens (
    id uuid NOT NULL,
    name character varying(64) NOT NULL,
    slug character varying(64) NOT NULL,
    icon_emoji character varying(8)
);


--
-- Name: cafe_locations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cafe_locations (
    id uuid NOT NULL,
    name character varying(160) NOT NULL,
    slug character varying(64) NOT NULL,
    address character varying(255) NOT NULL,
    latitude numeric(9,6) NOT NULL,
    longitude numeric(9,6) NOT NULL,
    phone character varying(32),
    opens_at time without time zone,
    closes_at time without time zone,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: cart_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cart_items (
    id uuid NOT NULL,
    cart_id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    variant_id uuid,
    quantity integer DEFAULT 1 NOT NULL,
    line_total numeric(10,2) NOT NULL,
    notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: carts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.carts (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    subtotal numeric(10,2) DEFAULT 0 NOT NULL,
    expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: customer_addresses; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customer_addresses (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    label character varying(64),
    line1 character varying(255) NOT NULL,
    line2 character varying(255),
    area character varying(128),
    city character varying(128) NOT NULL,
    postal_code character varying(16),
    latitude numeric(9,6),
    longitude numeric(9,6),
    instructions text,
    is_default boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: customers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.customers (
    id uuid NOT NULL,
    full_name character varying(160) NOT NULL,
    email character varying(255) NOT NULL,
    phone character varying(32),
    password_hash character varying(255) NOT NULL,
    default_address_id uuid,
    marketing_opt_in boolean DEFAULT false NOT NULL,
    status character varying(32) DEFAULT 'active'::character varying NOT NULL,
    last_login_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: deliveries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.deliveries (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    driver_id uuid NOT NULL,
    assigned_at timestamp with time zone DEFAULT now() NOT NULL,
    picked_up_at timestamp with time zone,
    delivered_at timestamp with time zone,
    expected_delivery_at timestamp with time zone,
    distance_km numeric(6,2),
    delivery_route jsonb DEFAULT '[]'::jsonb NOT NULL,
    proof_of_delivery_url character varying(512),
    customer_signature_url character varying(512),
    status character varying(32) DEFAULT 'assigned'::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: drivers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.drivers (
    id uuid NOT NULL,
    full_name character varying(160) NOT NULL,
    phone character varying(32) NOT NULL,
    email character varying(255),
    vehicle_type character varying(32) NOT NULL,
    license_plate character varying(32),
    status character varying(32) DEFAULT 'offline'::character varying NOT NULL,
    current_latitude numeric(9,6),
    current_longitude numeric(9,6),
    rating_average numeric(3,2) DEFAULT 0 NOT NULL,
    total_deliveries integer DEFAULT 0 NOT NULL,
    hired_at date,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    home_location_id uuid
);


--
-- Name: kysely_migration; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kysely_migration (
    name character varying(255) NOT NULL,
    "timestamp" character varying(255) NOT NULL
);


--
-- Name: kysely_migration_lock; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.kysely_migration_lock (
    id character varying(255) NOT NULL,
    is_locked integer DEFAULT 0 NOT NULL
);


--
-- Name: loyalty_accounts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loyalty_accounts (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    balance_points integer DEFAULT 0 NOT NULL,
    lifetime_points integer DEFAULT 0 NOT NULL,
    tier character varying(32) DEFAULT 'bronze'::character varying NOT NULL,
    tier_expires_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: loyalty_transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.loyalty_transactions (
    id uuid NOT NULL,
    account_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    order_id uuid,
    kind character varying(32) NOT NULL,
    points integer NOT NULL,
    description character varying(255),
    balance_after integer NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: menu_categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_categories (
    id uuid NOT NULL,
    name character varying(128) NOT NULL,
    slug character varying(64) NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: menu_item_allergens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_allergens (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    allergen_id uuid NOT NULL
);


--
-- Name: menu_item_modifier_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_modifier_groups (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    modifier_group_id uuid NOT NULL,
    is_required boolean DEFAULT false NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: menu_item_variants; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_item_variants (
    id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    name character varying(64) NOT NULL,
    price_delta numeric(10,2) DEFAULT 0 NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: menu_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.menu_items (
    id uuid NOT NULL,
    category_id uuid NOT NULL,
    name character varying(160) NOT NULL,
    slug character varying(96) NOT NULL,
    description text,
    base_price numeric(10,2) NOT NULL,
    is_vegetarian boolean DEFAULT false NOT NULL,
    is_vegan boolean DEFAULT false NOT NULL,
    is_signature boolean DEFAULT false NOT NULL,
    calories integer,
    prep_time_minutes integer,
    image_url character varying(512),
    is_available boolean DEFAULT true NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: modifier_groups; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modifier_groups (
    id uuid NOT NULL,
    name character varying(128) NOT NULL,
    selection_min integer DEFAULT 0 NOT NULL,
    selection_max integer DEFAULT 1 NOT NULL,
    sort_order integer DEFAULT 0 NOT NULL
);


--
-- Name: modifiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.modifiers (
    id uuid NOT NULL,
    modifier_group_id uuid NOT NULL,
    name character varying(128) NOT NULL,
    price_delta numeric(10,2) DEFAULT 0 NOT NULL,
    is_default boolean DEFAULT false NOT NULL
);


--
-- Name: notifications; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.notifications (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    channel character varying(16) NOT NULL,
    kind character varying(32) NOT NULL,
    title character varying(255) NOT NULL,
    body text,
    payload jsonb DEFAULT '{}'::jsonb NOT NULL,
    read_at timestamp with time zone,
    sent_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: order_item_modifiers; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_item_modifiers (
    id uuid NOT NULL,
    order_item_id uuid NOT NULL,
    modifier_id uuid NOT NULL,
    snapshot_name character varying(128) NOT NULL,
    snapshot_price_delta numeric(10,2) NOT NULL
);


--
-- Name: order_items; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_items (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    menu_item_id uuid NOT NULL,
    variant_id uuid,
    quantity integer NOT NULL,
    unit_price numeric(10,2) NOT NULL,
    line_total numeric(10,2) NOT NULL,
    snapshot_name character varying(160) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: order_promotions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.order_promotions (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    promo_code_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    discount_amount numeric(10,2) NOT NULL,
    applied_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    id uuid NOT NULL,
    customer_id uuid NOT NULL,
    address_id uuid NOT NULL,
    order_number character varying(32) NOT NULL,
    status character varying(32) DEFAULT 'placed'::character varying NOT NULL,
    placed_at timestamp with time zone DEFAULT now() NOT NULL,
    subtotal numeric(10,2) NOT NULL,
    delivery_fee numeric(10,2) DEFAULT 0 NOT NULL,
    discount_total numeric(10,2) DEFAULT 0 NOT NULL,
    tax numeric(10,2) DEFAULT 0 NOT NULL,
    grand_total numeric(10,2) NOT NULL,
    payment_method character varying(32) NOT NULL,
    scheduled_for timestamp with time zone,
    customer_notes text,
    internal_notes text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    location_id uuid
);


--
-- Name: payments; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.payments (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    method character varying(32) NOT NULL,
    provider_reference character varying(128),
    amount numeric(10,2) NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    captured_at timestamp with time zone,
    refunded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: promo_codes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.promo_codes (
    id uuid NOT NULL,
    code character varying(64) NOT NULL,
    description character varying(255),
    discount_type character varying(32) NOT NULL,
    discount_value numeric(10,2) NOT NULL,
    min_order_amount numeric(10,2) DEFAULT 0 NOT NULL,
    max_discount numeric(10,2),
    valid_from timestamp with time zone NOT NULL,
    valid_until timestamp with time zone NOT NULL,
    usage_limit integer,
    used_count integer DEFAULT 0 NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    id uuid NOT NULL,
    order_id uuid NOT NULL,
    customer_id uuid NOT NULL,
    menu_item_id uuid,
    driver_id uuid,
    rating integer NOT NULL,
    title character varying(255),
    body text,
    photos jsonb DEFAULT '[]'::jsonb NOT NULL,
    is_published boolean DEFAULT true NOT NULL,
    response text,
    responded_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


--
-- Data for Name: hypertable; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.hypertable (id, schema_name, table_name, associated_schema_name, associated_table_prefix, num_dimensions, chunk_sizing_func_schema, chunk_sizing_func_name, chunk_target_size, compression_state, compressed_hypertable_id, status) FROM stdin;
\.


--
-- Data for Name: bgw_job; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.bgw_job (id, application_name, schedule_interval, max_runtime, max_retries, retry_period, proc_schema, proc_name, owner, scheduled, fixed_schedule, initial_start, hypertable_id, config, check_schema, check_name, timezone) FROM stdin;
\.


--
-- Data for Name: chunk; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.chunk (id, hypertable_id, schema_name, table_name, compressed_chunk_id, status, osm_chunk, creation_time) FROM stdin;
\.


--
-- Data for Name: chunk_column_stats; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.chunk_column_stats (id, hypertable_id, chunk_id, column_name, range_start, range_end, valid) FROM stdin;
\.


--
-- Data for Name: compression_chunk_size; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.compression_chunk_size (chunk_id, compressed_chunk_id, uncompressed_heap_size, uncompressed_toast_size, uncompressed_index_size, compressed_heap_size, compressed_toast_size, compressed_index_size, numrows_pre_compression, numrows_post_compression, numrows_frozen_immediately) FROM stdin;
\.


--
-- Data for Name: compression_settings; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.compression_settings (relid, compress_relid, segmentby, orderby, orderby_desc, orderby_nullsfirst, index) FROM stdin;
\.


--
-- Data for Name: continuous_agg; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_agg (mat_hypertable_id, raw_hypertable_id, parent_mat_hypertable_id, user_view_schema, user_view_name, partial_view_schema, partial_view_name, direct_view_schema, direct_view_name, materialized_only, schema_change_timestamp) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_bucket_function; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_bucket_function (mat_hypertable_id, bucket_func, bucket_width, bucket_origin, bucket_offset, bucket_timezone, bucket_fixed_width) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_hypertable_invalidation_log; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_hypertable_invalidation_log (hypertable_id, lowest_modified_value, greatest_modified_value) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_invalidation_threshold; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_invalidation_threshold (hypertable_id, watermark) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_jobs_refresh_ranges; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_jobs_refresh_ranges (materialization_id, start_range, end_range, pid, job_id, created_at) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_materialization_invalidation_log; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_materialization_invalidation_log (materialization_id, lowest_modified_value, greatest_modified_value) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_materialization_ranges; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_materialization_ranges (materialization_id, lowest_modified_value, greatest_modified_value) FROM stdin;
\.


--
-- Data for Name: continuous_aggs_watermark; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.continuous_aggs_watermark (mat_hypertable_id, watermark) FROM stdin;
\.


--
-- Data for Name: dimension; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.dimension (id, hypertable_id, column_name, column_type, aligned, num_slices, partitioning_func_schema, partitioning_func, interval_length, compress_interval_length, integer_now_func_schema, integer_now_func) FROM stdin;
\.


--
-- Data for Name: dimension_slice; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.dimension_slice (id, chunk_id, dimension_id, range_start, range_end) FROM stdin;
\.


--
-- Data for Name: metadata; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.metadata (key, value, include_in_telemetry) FROM stdin;
install_timestamp	2026-06-27 16:21:52.237874+00	t
timescaledb_version	2.28.0	f
exported_uuid	3b306fdf-81b3-4f0c-983b-ca58689b096f	t
\.


--
-- Data for Name: tablespace; Type: TABLE DATA; Schema: _timescaledb_catalog; Owner: -
--

COPY _timescaledb_catalog.tablespace (id, hypertable_id, tablespace_name) FROM stdin;
\.


--
-- Data for Name: allergens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.allergens (id, name, slug, icon_emoji) FROM stdin;
019f09e3-547b-718d-83ce-0d51318992e8	Gluten	gluten	🌾
019f09e3-547b-718d-83ce-0d528ced4a5b	Dairy	dairy	🥛
019f09e3-547b-718d-83ce-0d537d0716c4	Eggs	eggs	🥚
019f09e3-547b-718d-83ce-0d540ab6ef89	Nuts	nuts	🥜
019f09e3-547b-718d-83ce-0d55710497d8	Soy	soy	🌱
019f09e3-547b-718d-83ce-0d56d3c60e62	Sesame	sesame	🪴
019f09e3-547b-718d-83ce-0d572de50d3a	Fish	fish	🐟
019f09e3-547b-718d-83ce-0d58776a6d07	Shellfish	shellfish	🦐
\.


--
-- Data for Name: cafe_locations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.cafe_locations (id, name, slug, address, latitude, longitude, phone, opens_at, closes_at, is_active, created_at) FROM stdin;
00000001-0000-7000-8000-000000000001	Emerald Cafe & Bakery — Gulshan	gulshan	House 42, Road 11, Gulshan-1, Dhaka	23.780600	90.415400	+8801800000001	08:00:00	23:00:00	t	2026-06-27 16:22:01.08958+00
00000002-0000-7000-8000-000000000002	Emerald Cafe & Bakery — Dhanmondi	dhanmondi	House 17, Road 7, Dhanmondi, Dhaka	23.746100	90.374200	+8801800000002	08:00:00	23:00:00	t	2026-06-27 16:22:01.08958+00
00000003-0000-7000-8000-000000000003	Emerald Cafe & Bakery — Banani	banani	House 22, Road 11, Banani, Dhaka	23.793700	90.406600	+8801800000003	08:00:00	23:00:00	t	2026-06-27 16:22:01.08958+00
\.


--
-- Data for Name: cart_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.cart_items (id, cart_id, menu_item_id, variant_id, quantity, line_total, notes, created_at) FROM stdin;
019f09e3-5455-754e-9c73-348313bf2430	019f09e3-5455-754e-9c73-347ec2796acc	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	Less ice please.	2026-06-27 16:22:01.045885+00
019f09e3-5455-754e-9c73-3484328c732e	019f09e3-5455-754e-9c73-347f03f580fa	019f09e3-5452-71ba-a465-4cabdf313210	\N	2	480.00	\N	2026-06-27 16:22:01.045885+00
019f09e3-5455-754e-9c73-348515fde0c3	019f09e3-5455-754e-9c73-3480387b7b3b	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	\N	2026-06-27 16:22:01.045885+00
019f09e3-5455-754e-9c73-3486b3a19b1f	019f09e3-5455-754e-9c73-3481a0a9510c	019f09e3-5452-71ba-a465-4cb119d7d975	\N	2	440.00	\N	2026-06-27 16:22:01.045885+00
019f09e3-5455-754e-9c73-3487a91f3c27	019f09e3-5455-754e-9c73-34829257d87c	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	\N	2026-06-27 16:22:01.045885+00
\.


--
-- Data for Name: carts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.carts (id, customer_id, subtotal, expires_at, created_at, updated_at, location_id) FROM stdin;
019f09e3-5455-754e-9c73-347f03f580fa	019f09e3-5442-7d10-91b2-be18ecd018b1	480.00	2025-12-03 12:00:00+00	2026-06-27 16:22:01.045428+00	2026-06-27 16:22:01.045428+00	00000001-0000-7000-8000-000000000001
019f09e3-5455-754e-9c73-3480387b7b3b	019f09e3-5442-7d10-91b2-be195b3c14e2	360.00	2025-12-03 12:00:00+00	2026-06-27 16:22:01.045428+00	2026-06-27 16:22:01.045428+00	00000001-0000-7000-8000-000000000001
019f09e3-5455-754e-9c73-34829257d87c	019f09e3-5442-7d10-91b2-be1b7dad8028	320.00	2025-12-03 12:00:00+00	2026-06-27 16:22:01.045428+00	2026-06-27 16:22:01.045428+00	00000001-0000-7000-8000-000000000001
019f09e3-5455-754e-9c73-3481a0a9510c	019f09e3-5442-7d10-91b2-be1a53eca627	440.00	2025-12-03 12:00:00+00	2026-06-27 16:22:01.045428+00	2026-06-27 16:22:01.045428+00	00000001-0000-7000-8000-000000000001
019f09e3-5455-754e-9c73-347ec2796acc	019f09e3-5442-7d10-91b2-be17f929eb2d	380.00	2025-12-03 12:00:00+00	2026-06-27 16:22:01.045428+00	2026-06-27 16:22:01.045428+00	00000001-0000-7000-8000-000000000001
\.


--
-- Data for Name: customer_addresses; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.customer_addresses (id, customer_id, label, line1, line2, area, city, postal_code, latitude, longitude, instructions, is_default, created_at) FROM stdin;
019f09e3-5443-7fbe-b1a5-b10eaebf83c5	019f09e3-5442-7d10-91b2-be17f929eb2d	Home	Flat 10, House 20	Road 5	Gulshan	Dhaka	1212	23.792500	90.407800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b10f47601892	019f09e3-5442-7d10-91b2-be17f929eb2d	Work	Office 100	\N	Bashundhara	Dhaka	1229	23.813700	90.430800	Reception desk.	f	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	019f09e3-5442-7d10-91b2-be18ecd018b1	Home	Flat 11, House 21	Road 6	Banani	Dhaka	1213	23.793700	90.406600	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b111a8b8d367	019f09e3-5442-7d10-91b2-be195b3c14e2	Home	Flat 12, House 22	Road 7	Bashundhara	Dhaka	1229	23.813700	90.430800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1128f3e4441	019f09e3-5442-7d10-91b2-be1a53eca627	Home	Flat 13, House 23	Road 8	Dhanmondi	Dhaka	1209	23.746100	90.374200	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b113dc7cd3cf	019f09e3-5442-7d10-91b2-be1a53eca627	Work	Office 103	\N	Mohakhali	Dhaka	1212	23.777500	90.405300	Reception desk.	f	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b114d0308581	019f09e3-5442-7d10-91b2-be1b7dad8028	Home	Flat 14, House 24	Road 9	Uttara	Dhaka	1230	23.875900	90.379500	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1152464cd40	019f09e3-5442-7d10-91b2-be1c99a7d8f2	Home	Flat 15, House 25	Road 10	Mohakhali	Dhaka	1212	23.777500	90.405300	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11657ae506c	019f09e3-5442-7d10-91b2-be1d7fed74b3	Home	Flat 16, House 26	Road 11	Gulshan	Dhaka	1212	23.792500	90.407800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1177d7c7e3d	019f09e3-5442-7d10-91b2-be1d7fed74b3	Work	Office 106	\N	Bashundhara	Dhaka	1229	23.813700	90.430800	Reception desk.	f	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b118417956c0	019f09e3-5442-7d10-91b2-be1e19f149d9	Home	Flat 17, House 27	Road 5	Banani	Dhaka	1213	23.793700	90.406600	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11961535e61	019f09e3-5442-7d10-91b2-be1f5114d14b	Home	Flat 18, House 28	Road 6	Bashundhara	Dhaka	1229	23.813700	90.430800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11aff50cbea	019f09e3-5442-7d10-91b2-be2064b523b7	Home	Flat 19, House 29	Road 7	Dhanmondi	Dhaka	1209	23.746100	90.374200	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11b6e5a60b5	019f09e3-5442-7d10-91b2-be2064b523b7	Work	Office 109	\N	Mohakhali	Dhaka	1212	23.777500	90.405300	Reception desk.	f	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11cba8fdf04	019f09e3-5442-7d10-91b2-be2144da3493	Home	Flat 20, House 30	Road 8	Uttara	Dhaka	1230	23.875900	90.379500	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11d86170386	019f09e3-5442-7d10-91b2-be2261956b32	Home	Flat 21, House 31	Road 9	Mohakhali	Dhaka	1212	23.777500	90.405300	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11ea3efbb3d	019f09e3-5442-7d10-91b2-be233ab206bc	Home	Flat 22, House 32	Road 10	Gulshan	Dhaka	1212	23.792500	90.407800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b11fea4f66d5	019f09e3-5442-7d10-91b2-be233ab206bc	Work	Office 112	\N	Bashundhara	Dhaka	1229	23.813700	90.430800	Reception desk.	f	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1205cad0de2	019f09e3-5442-7d10-91b2-be24727af8e6	Home	Flat 23, House 33	Road 11	Banani	Dhaka	1213	23.793700	90.406600	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
019f09e3-5443-7fbe-b1a5-b1210ab70e3b	019f09e3-5442-7d10-91b2-be25179006d0	Home	Flat 24, House 34	Road 5	Bashundhara	Dhaka	1229	23.813700	90.430800	Ring the bell twice.	t	2026-06-27 16:22:01.028296+00
\.


--
-- Data for Name: customers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.customers (id, full_name, email, phone, password_hash, default_address_id, marketing_opt_in, status, last_login_at, created_at, updated_at) FROM stdin;
019f09e3-5442-7d10-91b2-be17f929eb2d	Aisha Rahman	aisha.rahman@example.com	+8801711000001	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b10eaebf83c5	t	active	2025-12-01 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be2144da3493	Tasnim Ferdous	tasnim.ferdous@example.com	+8801711000011	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11cba8fdf04	t	active	2025-11-28 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be2261956b32	Mahbub Alam	mahbub.alam@example.com	+8801711000012	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11d86170386	f	active	2025-11-27 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be233ab206bc	Sumaya Khan	sumaya.khan@example.com	+8801711000013	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11ea3efbb3d	t	active	2025-11-26 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be24727af8e6	Ashraf Uddin	ashraf.uddin@example.com	+8801711000014	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b1205cad0de2	f	active	2025-11-25 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be25179006d0	Nabila Tabassum	nabila.tabassum@example.com	+8801711000015	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b1210ab70e3b	t	active	2025-12-01 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be195b3c14e2	Nusrat Jahan	nusrat.jahan@example.com	+8801711000003	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b111a8b8d367	t	active	2025-11-29 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1b7dad8028	Farzana Akter	farzana.akter@example.com	+8801711000005	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b114d0308581	t	active	2025-11-27 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1d7fed74b3	Mehnaz Sultana	mehnaz.sultana@example.com	+8801711000007	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11657ae506c	t	active	2025-11-25 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1c99a7d8f2	Sajid Hossain	sajid.hossain@example.com	+8801711000006	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b1152464cd40	f	active	2025-11-26 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be18ecd018b1	Tanvir Ahmed	tanvir.ahmed@example.com	+8801711000002	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	f	active	2025-11-30 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be2064b523b7	Rifat Haque	rifat.haque@example.com	+8801711000010	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11aff50cbea	f	active	2025-11-29 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1e19f149d9	Imran Chowdhury	imran.chowdhury@example.com	+8801711000008	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b118417956c0	f	active	2025-12-01 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1a53eca627	Rezaul Karim	rezaul.karim@example.com	+8801711000004	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b1128f3e4441	f	active	2025-11-28 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
019f09e3-5442-7d10-91b2-be1f5114d14b	Sadia Islam	sadia.islam@example.com	+8801711000009	$argon2id$v=19$placeholder	019f09e3-5443-7fbe-b1a5-b11961535e61	t	active	2025-11-30 12:00:00+00	2026-06-27 16:22:01.027166+00	2026-06-27 16:22:01.027166+00
\.


--
-- Data for Name: deliveries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.deliveries (id, order_id, driver_id, assigned_at, picked_up_at, delivered_at, expected_delivery_at, distance_km, delivery_route, proof_of_delivery_url, customer_signature_url, status, created_at, updated_at) FROM stdin;
019f09e3-545c-7d9e-be43-2d66c9bbfbd7	019f09e3-5457-70fb-820c-b7eb090f2dc5	019f09e3-5456-7446-ba30-bd8860c1a192	2025-12-01 12:05:00+00	2025-12-01 12:25:00+00	2025-12-01 12:55:00+00	2025-12-01 12:50:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7eb090f2dc5.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d67c86dfd7a	019f09e3-5457-70fb-820c-b7ecd578786e	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-11-30 07:17:00+00	2025-11-30 07:37:00+00	2025-11-30 08:07:00+00	2025-11-30 08:02:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7ecd578786e.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6836033eb0	019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5456-7446-ba30-bd8a9394b803	2025-11-29 02:29:00+00	2025-11-29 02:49:00+00	2025-11-29 03:19:00+00	2025-11-29 03:14:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7ed59eb2dd4.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d69aea865b7	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-11-27 21:41:00+00	2025-11-27 22:01:00+00	2025-11-27 22:31:00+00	2025-11-27 22:26:00+00	3.60	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7eecd8ef115.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6adff760f0	019f09e3-5457-70fb-820c-b7efd97a1689	019f09e3-5456-7446-ba30-bd8c67a34936	2025-11-26 16:53:00+00	2025-11-26 17:13:00+00	2025-11-26 17:43:00+00	2025-11-26 17:38:00+00	4.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7efd97a1689.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6b19ed32a8	019f09e3-5457-70fb-820c-b7f060508c6b	019f09e3-5456-7446-ba30-bd8d92b58338	2025-11-25 12:05:00+00	2025-11-25 12:25:00+00	2025-11-25 12:55:00+00	2025-11-25 12:50:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f060508c6b.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6c43d3c06f	019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5456-7446-ba30-bd8860c1a192	2025-11-24 07:17:00+00	2025-11-24 07:37:00+00	2025-11-24 08:07:00+00	2025-11-24 08:02:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f140dd23d3.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6d87f4fc08	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-11-23 02:29:00+00	2025-11-23 02:49:00+00	2025-11-23 03:19:00+00	2025-11-23 03:14:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f26644452d.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6e33c4cf54	019f09e3-5457-70fb-820c-b7f3b082e1c7	019f09e3-5456-7446-ba30-bd8a9394b803	2025-11-21 21:41:00+00	2025-11-21 22:01:00+00	2025-11-21 22:31:00+00	2025-11-21 22:26:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f3b082e1c7.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d6fa1ea05c4	019f09e3-5457-70fb-820c-b7f4248ea60f	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-11-20 16:53:00+00	2025-11-20 17:13:00+00	2025-11-20 17:43:00+00	2025-11-20 17:38:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f4248ea60f.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7092e73645	019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5456-7446-ba30-bd8c67a34936	2025-11-19 12:05:00+00	2025-11-19 12:25:00+00	2025-11-19 12:55:00+00	2025-11-19 12:50:00+00	3.60	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f58cde71d0.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d710aa74154	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5456-7446-ba30-bd8d92b58338	2025-11-18 07:17:00+00	2025-11-18 07:37:00+00	2025-11-18 08:07:00+00	2025-11-18 08:02:00+00	4.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f60c423fb4.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d724e4af8a4	019f09e3-5457-70fb-820c-b7f7f360daf9	019f09e3-5456-7446-ba30-bd8860c1a192	2025-11-17 02:29:00+00	2025-11-17 02:49:00+00	2025-11-17 03:19:00+00	2025-11-17 03:14:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f7f360daf9.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d73bb17a3d6	019f09e3-5457-70fb-820c-b7f85bf8f694	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-11-15 21:41:00+00	2025-11-15 22:01:00+00	2025-11-15 22:31:00+00	2025-11-15 22:26:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f85bf8f694.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7458eb16b7	019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5456-7446-ba30-bd8a9394b803	2025-11-14 16:53:00+00	2025-11-14 17:13:00+00	2025-11-14 17:43:00+00	2025-11-14 17:38:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7f99c6a23a9.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d75c27e9f57	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-11-13 12:05:00+00	2025-11-13 12:25:00+00	2025-11-13 12:55:00+00	2025-11-13 12:50:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7fab9e7a570.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d76d1aad9b3	019f09e3-5457-70fb-820c-b7fb9a3a04fb	019f09e3-5456-7446-ba30-bd8c67a34936	2025-11-12 07:17:00+00	2025-11-12 07:37:00+00	2025-11-12 08:07:00+00	2025-11-12 08:02:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7fb9a3a04fb.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d77beb21c62	019f09e3-5457-70fb-820c-b7fc7dd1ca8a	019f09e3-5456-7446-ba30-bd8d92b58338	2025-11-11 02:29:00+00	2025-11-11 02:49:00+00	2025-11-11 03:19:00+00	2025-11-11 03:14:00+00	3.60	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7fc7dd1ca8a.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7863af0514	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5456-7446-ba30-bd8860c1a192	2025-11-09 21:41:00+00	2025-11-09 22:01:00+00	2025-11-09 22:31:00+00	2025-11-09 22:26:00+00	4.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7fd427a8c56.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d79fceef9a0	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-11-08 16:53:00+00	2025-11-08 17:13:00+00	2025-11-08 17:43:00+00	2025-11-08 17:38:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7fe32fb109f.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7a187c0944	019f09e3-5457-70fb-820c-b7ff0900f2e9	019f09e3-5456-7446-ba30-bd8a9394b803	2025-11-07 12:05:00+00	2025-11-07 12:25:00+00	2025-11-07 12:55:00+00	2025-11-07 12:50:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b7ff0900f2e9.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7bf54f107e	019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-11-06 07:17:00+00	2025-11-06 07:37:00+00	2025-11-06 08:07:00+00	2025-11-06 08:02:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b8007301815b.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7c552c5f0d	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5456-7446-ba30-bd8c67a34936	2025-11-05 02:29:00+00	2025-11-05 02:49:00+00	2025-11-05 03:19:00+00	2025-11-05 03:14:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b801d964b111.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7d0c2c0802	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5456-7446-ba30-bd8d92b58338	2025-11-03 21:41:00+00	2025-11-03 22:01:00+00	2025-11-03 22:31:00+00	2025-11-03 22:26:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b802ec565457.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7e3cd6103d	019f09e3-5457-70fb-820c-b8035a6701f6	019f09e3-5456-7446-ba30-bd8860c1a192	2025-11-02 16:53:00+00	2025-11-02 17:13:00+00	2025-11-02 17:43:00+00	2025-11-02 17:38:00+00	3.60	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b8035a6701f6.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d7ffba35d1b	019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-11-01 12:05:00+00	2025-11-01 12:25:00+00	2025-11-01 12:55:00+00	2025-11-01 12:50:00+00	4.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b804031f3f4c.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8067b72772	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5456-7446-ba30-bd8a9394b803	2025-10-31 07:17:00+00	2025-10-31 07:37:00+00	2025-10-31 08:07:00+00	2025-10-31 08:02:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80576bf5eb5.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d81c06a6aea	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-10-30 02:29:00+00	2025-10-30 02:49:00+00	2025-10-30 03:19:00+00	2025-10-30 03:14:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80628f307ca.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d82ead92e71	019f09e3-5457-70fb-820c-b80788ff8136	019f09e3-5456-7446-ba30-bd8c67a34936	2025-10-28 21:41:00+00	2025-10-28 22:01:00+00	2025-10-28 22:31:00+00	2025-10-28 22:26:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80788ff8136.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d834ef2054e	019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5456-7446-ba30-bd8d92b58338	2025-10-27 16:53:00+00	2025-10-27 17:13:00+00	2025-10-27 17:43:00+00	2025-10-27 17:38:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b808a56cb42c.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d84b41c07e9	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5456-7446-ba30-bd8860c1a192	2025-10-26 12:05:00+00	2025-10-26 12:25:00+00	2025-10-26 12:55:00+00	2025-10-26 12:50:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b809709e73e2.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d85e684df3d	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-10-25 07:17:00+00	2025-10-25 07:37:00+00	2025-10-25 08:07:00+00	2025-10-25 08:02:00+00	3.60	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80ae19741d1.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8650534133	019f09e3-5457-70fb-820c-b80baa631ff0	019f09e3-5456-7446-ba30-bd8a9394b803	2025-10-24 02:29:00+00	2025-10-24 02:49:00+00	2025-10-24 03:19:00+00	2025-10-24 03:14:00+00	4.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80baa631ff0.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d873467b5cc	019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5456-7446-ba30-bd8b16ac507f	2025-10-22 21:41:00+00	2025-10-22 22:01:00+00	2025-10-22 22:31:00+00	2025-10-22 22:26:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80cd473ce94.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d88e5a8584e	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5456-7446-ba30-bd8c67a34936	2025-10-21 16:53:00+00	2025-10-21 17:13:00+00	2025-10-21 17:43:00+00	2025-10-21 17:38:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80d12f1d6de.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d89a0a9eda9	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5456-7446-ba30-bd8d92b58338	2025-10-20 12:05:00+00	2025-10-20 12:25:00+00	2025-10-20 12:55:00+00	2025-10-20 12:50:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80e0158a4d7.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8a7af782e2	019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5456-7446-ba30-bd8860c1a192	2025-10-19 07:17:00+00	2025-10-19 07:37:00+00	2025-10-19 08:07:00+00	2025-10-19 08:02:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b80f0832fbf9.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8bdc9f92a6	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-10-18 02:29:00+00	2025-10-18 02:49:00+00	2025-10-18 03:19:00+00	2025-10-18 03:14:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b8101276b16a.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8ce7b482a0	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-10-10 21:41:00+00	2025-10-10 22:01:00+00	\N	2025-10-10 22:26:00+00	4.00	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	\N	\N	in_transit	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8d2790112e	019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5456-7446-ba30-bd8a9394b803	2025-10-09 16:53:00+00	2025-10-09 17:13:00+00	\N	2025-10-09 17:38:00+00	2.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	\N	\N	in_transit	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8e98c06327	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5456-7446-ba30-bd8d92b58338	2025-10-06 02:29:00+00	2025-10-06 02:49:00+00	2025-10-06 03:19:00+00	2025-10-06 03:14:00+00	3.20	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b81aa71679ba.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d8f6a21ddea	019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5456-7446-ba30-bd8860c1a192	2025-10-04 21:41:00+00	2025-10-04 22:01:00+00	2025-10-04 22:31:00+00	2025-10-04 22:26:00+00	4.40	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b81b4b9dc427.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
019f09e3-545c-7d9e-be43-2d90a3f88dac	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5456-7446-ba30-bd89b5b9f431	2025-10-03 16:53:00+00	2025-10-03 17:13:00+00	2025-10-03 17:43:00+00	2025-10-03 17:38:00+00	2.80	[{"lat": 23.7806, "lng": 90.4154, "label": "cafe"}, {"lat": 23.79, "lng": 90.42, "label": "midway"}, {"lat": 23.7925, "lng": 90.4078, "label": "customer"}]	https://cdn.example.com/pod/019f09e3-5457-70fb-820c-b81c37a0871d.jpg	\N	delivered	2026-06-27 16:22:01.053172+00	2026-06-27 16:22:01.053172+00
\.


--
-- Data for Name: drivers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.drivers (id, full_name, phone, email, vehicle_type, license_plate, status, current_latitude, current_longitude, rating_average, total_deliveries, hired_at, created_at, updated_at, home_location_id) FROM stdin;
019f09e3-5456-7446-ba30-bd8c67a34936	Habibur Sheikh	+8801911000005	habibur.sheikh@drivers.example.com	bike	DHA-METRO-K-12-3405	busy	23.820000	90.410000	4.40	168	2025-02-04	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000002-0000-7000-8000-000000000002
019f09e3-5456-7446-ba30-bd8b16ac507f	Mizanur Rahman	+8801911000004	mizanur.rahman@drivers.example.com	bicycle	\N	busy	23.810000	90.440000	4.30	151	2025-03-06	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000001-0000-7000-8000-000000000001
019f09e3-5456-7446-ba30-bd89b5b9f431	Jubair Ali	+8801911000002	jubair.ali@drivers.example.com	bike	DHA-METRO-K-12-3402	available	23.790000	90.420000	4.10	117	2025-05-05	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000002-0000-7000-8000-000000000002
019f09e3-5456-7446-ba30-bd8a9394b803	Shoaib Khan	+8801911000003	shoaib.khan@drivers.example.com	bike	DHA-METRO-K-12-3403	available	23.800000	90.430000	4.20	134	2025-04-05	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000003-0000-7000-8000-000000000003
019f09e3-5456-7446-ba30-bd8d92b58338	Sajjad Hossen	+8801911000006	sajjad.hossen@drivers.example.com	car	DHA-METRO-GA-21-1199	offline	23.780000	90.420000	4.50	185	2025-01-05	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000003-0000-7000-8000-000000000003
019f09e3-5456-7446-ba30-bd8860c1a192	Rakib Hasan	+8801911000001	rakib.hasan@drivers.example.com	bike	DHA-METRO-K-12-3401	available	23.780000	90.410000	4.00	100	2025-06-04	2026-06-27 16:22:01.046987+00	2026-06-27 16:22:01.046987+00	00000001-0000-7000-8000-000000000001
\.


--
-- Data for Name: kysely_migration; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.kysely_migration (name, "timestamp") FROM stdin;
s1_001_customers	2026-06-27T16:22:00.028Z
s1_002_menu	2026-06-27T16:22:00.032Z
s1_003_carts_and_orders	2026-06-27T16:22:00.036Z
s1_004_drivers_and_deliveries	2026-06-27T16:22:00.038Z
s2_001_promotions	2026-06-27T16:22:00.040Z
s3_001_reviews	2026-06-27T16:22:00.040Z
s3_002_loyalty	2026-06-27T16:22:00.042Z
s3_003_notifications	2026-06-27T16:22:00.043Z
s4_001_allergens	2026-06-27T16:22:00.044Z
s5_001_multi_location	2026-06-27T16:22:00.046Z
\.


--
-- Data for Name: kysely_migration_lock; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.kysely_migration_lock (id, is_locked) FROM stdin;
migration_lock	0
\.


--
-- Data for Name: loyalty_accounts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.loyalty_accounts (id, customer_id, balance_points, lifetime_points, tier, tier_expires_at, created_at, updated_at) FROM stdin;
019f09e3-5472-7ab1-adbd-28263ef85c71	019f09e3-5442-7d10-91b2-be17f929eb2d	50	200	bronze	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-2827d79f3b20	019f09e3-5442-7d10-91b2-be24727af8e6	80	270	bronze	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	110	340	bronze	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-2829be145a74	019f09e3-5442-7d10-91b2-be1e19f149d9	140	410	bronze	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282aa2236cda	019f09e3-5442-7d10-91b2-be2261956b32	170	480	silver	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	200	550	silver	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282ce3967243	019f09e3-5442-7d10-91b2-be25179006d0	230	620	silver	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282db54c8d7c	019f09e3-5442-7d10-91b2-be195b3c14e2	260	690	silver	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	290	760	silver	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-282f1bcb7c9b	019f09e3-5442-7d10-91b2-be2064b523b7	320	830	gold	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-2830e7bf5347	019f09e3-5442-7d10-91b2-be1f5114d14b	350	900	gold	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	380	970	gold	2026-05-30 12:00:00+00	2026-06-27 16:22:01.074967+00	2026-06-27 16:22:01.074967+00
\.


--
-- Data for Name: loyalty_transactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.loyalty_transactions (id, account_id, customer_id, order_id, kind, points, description, balance_after, created_at) FROM stdin;
019f09e3-5473-787b-a80e-f2ab1f2ce22c	019f09e3-5472-7ab1-adbd-28263ef85c71	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5457-70fb-820c-b81c37a0871d	earn	10	Order ED-01049 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2aceef78576	019f09e3-5472-7ab1-adbd-28263ef85c71	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5457-70fb-820c-b81b4b9dc427	earn	15	Order ED-01048 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2ada94dd8ef	019f09e3-5472-7ab1-adbd-28263ef85c71	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5457-70fb-820c-b81aa71679ba	earn	20	Order ED-01047 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2ae6e6569d9	019f09e3-5472-7ab1-adbd-28263ef85c71	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5457-70fb-820c-b81c37a0871d	redeem	-50	Redeemed for 50 BDT discount	100	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2afbb418911	019f09e3-5472-7ab1-adbd-2827d79f3b20	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5457-70fb-820c-b8171cace121	earn	10	Order ED-01044 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b0daa7001e	019f09e3-5472-7ab1-adbd-2827d79f3b20	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5457-70fb-820c-b816bd1873be	earn	15	Order ED-01043 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b137b20783	019f09e3-5472-7ab1-adbd-2827d79f3b20	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5457-70fb-820c-b815aa76d847	earn	20	Order ED-01042 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b242a850a1	019f09e3-5472-7ab1-adbd-2827d79f3b20	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5457-70fb-820c-b814d7d9057f	earn	25	Order ED-01041 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b3339c9eb9	019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5457-70fb-820c-b812214b2b56	earn	10	Order ED-01039 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b409ff48ae	019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5457-70fb-820c-b811f2b5171b	earn	15	Order ED-01038 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b580a24371	019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5457-70fb-820c-b8101276b16a	earn	20	Order ED-01037 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b618821b00	019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5457-70fb-820c-b80f0832fbf9	earn	25	Order ED-01036 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b74e526339	019f09e3-5472-7ab1-adbd-282827b697eb	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5457-70fb-820c-b80e0158a4d7	earn	30	Order ED-01035 earned points	110	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b86c7d2be4	019f09e3-5472-7ab1-adbd-2829be145a74	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5457-70fb-820c-b80d12f1d6de	earn	10	Order ED-01034 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2b9e8dea25f	019f09e3-5472-7ab1-adbd-2829be145a74	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5457-70fb-820c-b80cd473ce94	earn	15	Order ED-01033 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2ba7c9511ab	019f09e3-5472-7ab1-adbd-2829be145a74	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5457-70fb-820c-b80baa631ff0	earn	20	Order ED-01032 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2bb3fb8b187	019f09e3-5472-7ab1-adbd-2829be145a74	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5457-70fb-820c-b8199c81e149	redeem	-50	Redeemed for 50 BDT discount	100	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2bc7a75b229	019f09e3-5472-7ab1-adbd-282aa2236cda	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5457-70fb-820c-b808a56cb42c	earn	10	Order ED-01029 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2bdc5a874dd	019f09e3-5472-7ab1-adbd-282aa2236cda	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5457-70fb-820c-b80788ff8136	earn	15	Order ED-01028 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2bee28002c2	019f09e3-5472-7ab1-adbd-282aa2236cda	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5457-70fb-820c-b80628f307ca	earn	20	Order ED-01027 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2bf5074e4ee	019f09e3-5472-7ab1-adbd-282aa2236cda	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5457-70fb-820c-b80576bf5eb5	earn	25	Order ED-01026 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c08479612f	019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5457-70fb-820c-b8035a6701f6	earn	10	Order ED-01024 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c11f8a2e6f	019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5457-70fb-820c-b802ec565457	earn	15	Order ED-01023 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c2796dce36	019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5457-70fb-820c-b801d964b111	earn	20	Order ED-01022 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c3a3f1f1ed	019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5457-70fb-820c-b8007301815b	earn	25	Order ED-01021 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c480549c5e	019f09e3-5472-7ab1-adbd-282bc5db3911	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5457-70fb-820c-b7ff0900f2e9	earn	30	Order ED-01020 earned points	110	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c58cb40ec0	019f09e3-5472-7ab1-adbd-282ce3967243	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5457-70fb-820c-b7fe32fb109f	earn	10	Order ED-01019 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c66223849b	019f09e3-5472-7ab1-adbd-282ce3967243	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5457-70fb-820c-b7fd427a8c56	earn	15	Order ED-01018 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c737aadbbf	019f09e3-5472-7ab1-adbd-282ce3967243	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5457-70fb-820c-b7fc7dd1ca8a	earn	20	Order ED-01017 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c8a863fc8b	019f09e3-5472-7ab1-adbd-282ce3967243	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5457-70fb-820c-b816bd1873be	redeem	-50	Redeemed for 50 BDT discount	100	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2c9a60c701e	019f09e3-5472-7ab1-adbd-282db54c8d7c	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5457-70fb-820c-b7f99c6a23a9	earn	10	Order ED-01014 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2ca29e9ec97	019f09e3-5472-7ab1-adbd-282db54c8d7c	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5457-70fb-820c-b7f85bf8f694	earn	15	Order ED-01013 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2cb96661f03	019f09e3-5472-7ab1-adbd-282db54c8d7c	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5457-70fb-820c-b7f7f360daf9	earn	20	Order ED-01012 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2ccb027aaa8	019f09e3-5472-7ab1-adbd-282db54c8d7c	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5457-70fb-820c-b7f60c423fb4	earn	25	Order ED-01011 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2cd85db5687	019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5457-70fb-820c-b7f4248ea60f	earn	10	Order ED-01009 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2cec08ccd96	019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5457-70fb-820c-b7f3b082e1c7	earn	15	Order ED-01008 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2cf44e11a11	019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5457-70fb-820c-b7f26644452d	earn	20	Order ED-01007 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d0ccfdd75e	019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5457-70fb-820c-b7f140dd23d3	earn	25	Order ED-01006 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d1028d71a6	019f09e3-5472-7ab1-adbd-282eee77bd89	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5457-70fb-820c-b7f060508c6b	earn	30	Order ED-01005 earned points	110	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d21ca5ff7e	019f09e3-5472-7ab1-adbd-282f1bcb7c9b	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5457-70fb-820c-b7efd97a1689	earn	10	Order ED-01004 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d345a45b65	019f09e3-5472-7ab1-adbd-282f1bcb7c9b	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5457-70fb-820c-b7eecd8ef115	earn	15	Order ED-01003 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d4899a3326	019f09e3-5472-7ab1-adbd-282f1bcb7c9b	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5457-70fb-820c-b7ed59eb2dd4	earn	20	Order ED-01002 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d51ba8185a	019f09e3-5472-7ab1-adbd-282f1bcb7c9b	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5457-70fb-820c-b813f82db22f	redeem	-50	Redeemed for 50 BDT discount	100	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d6ca18d0d1	019f09e3-5472-7ab1-adbd-2830e7bf5347	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5457-70fb-820c-b81c37a0871d	earn	10	Order ED-01049 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d7e211eb4e	019f09e3-5472-7ab1-adbd-2830e7bf5347	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5457-70fb-820c-b81b4b9dc427	earn	15	Order ED-01048 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d803795d5a	019f09e3-5472-7ab1-adbd-2830e7bf5347	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5457-70fb-820c-b81aa71679ba	earn	20	Order ED-01047 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2d900de3dfd	019f09e3-5472-7ab1-adbd-2830e7bf5347	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5457-70fb-820c-b8199c81e149	earn	25	Order ED-01046 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2daeec92399	019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5457-70fb-820c-b8171cace121	earn	10	Order ED-01044 earned points	50	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2dbee491ec0	019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5457-70fb-820c-b816bd1873be	earn	15	Order ED-01043 earned points	65	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2dc73a3f9e9	019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5457-70fb-820c-b815aa76d847	earn	20	Order ED-01042 earned points	80	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2dd42a6e07a	019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5457-70fb-820c-b814d7d9057f	earn	25	Order ED-01041 earned points	95	2026-06-27 16:22:01.075575+00
019f09e3-5473-787b-a80e-f2decdc0ace0	019f09e3-5472-7ab1-adbd-283184bf870a	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5457-70fb-820c-b813f82db22f	earn	30	Order ED-01040 earned points	110	2026-06-27 16:22:01.075575+00
\.


--
-- Data for Name: menu_categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.menu_categories (id, name, slug, sort_order, is_active, created_at, location_id) FROM stdin;
019f09e3-5451-78e5-8697-6b5fef33aa8e	Breakfast	breakfast	1	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
019f09e3-5451-78e5-8697-6b60fe4455a9	Cakes	cakes	2	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
019f09e3-5451-78e5-8697-6b61c642cfbe	Coffee	coffee	3	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
019f09e3-5451-78e5-8697-6b6233c3e284	Cold Drinks	cold-drinks	4	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
019f09e3-5451-78e5-8697-6b6355f074b0	Pastries	pastries	5	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
019f09e3-5451-78e5-8697-6b64759768ec	Sandwiches	sandwiches	6	t	2026-06-27 16:22:01.041863+00	00000001-0000-7000-8000-000000000001
\.


--
-- Data for Name: menu_item_allergens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.menu_item_allergens (id, menu_item_id, allergen_id) FROM stdin;
019f09e3-547c-7f93-876c-212caaf43dad	019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-212dc8ec6c6d	019f09e3-5452-71ba-a465-4ca93d66db4f	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-212ed9f4ad94	019f09e3-5452-71ba-a465-4caa0b18f629	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-212ff3678e06	019f09e3-5452-71ba-a465-4cacba4db301	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21300387cc6f	019f09e3-5452-71ba-a465-4caf98b840ae	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21319f08e46b	019f09e3-5452-71ba-a465-4cb00757b39f	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2132add52d53	019f09e3-5452-71ba-a465-4cb00757b39f	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-213388564d6c	019f09e3-5452-71ba-a465-4cb00757b39f	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-2134d95f763b	019f09e3-5452-71ba-a465-4cb119d7d975	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21350bcff270	019f09e3-5452-71ba-a465-4cb119d7d975	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2136c23997c9	019f09e3-5452-71ba-a465-4cb119d7d975	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-2137e9422a51	019f09e3-5452-71ba-a465-4cb29c358ed7	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-213806f8da90	019f09e3-5452-71ba-a465-4cb29c358ed7	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2139500b3414	019f09e3-5452-71ba-a465-4cb29c358ed7	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-213af7c3a52e	019f09e3-5452-71ba-a465-4cb29c358ed7	019f09e3-547b-718d-83ce-0d540ab6ef89
019f09e3-547c-7f93-876c-213b921b6758	019f09e3-5452-71ba-a465-4cb31033def5	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-213cb5dcca6e	019f09e3-5452-71ba-a465-4cb31033def5	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-213d561ccb46	019f09e3-5452-71ba-a465-4cb31033def5	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-213e29a72731	019f09e3-5452-71ba-a465-4cb45101ddbf	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-213f8ed3e15d	019f09e3-5452-71ba-a465-4cb45101ddbf	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2140e19b9ebc	019f09e3-5452-71ba-a465-4cb45101ddbf	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-214114ffbba8	019f09e3-5452-71ba-a465-4cb5e0a557d4	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-21423389499d	019f09e3-5452-71ba-a465-4cb671a7e228	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-2143bef4ca4f	019f09e3-5452-71ba-a465-4cb7aa36e904	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2144e0bdd7c8	019f09e3-5452-71ba-a465-4cb7aa36e904	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-214581949751	019f09e3-5452-71ba-a465-4cb898c932f7	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21462a99d0ae	019f09e3-5452-71ba-a465-4cb898c932f7	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-21472649a94f	019f09e3-5452-71ba-a465-4cb898c932f7	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-2148e230abe4	019f09e3-5452-71ba-a465-4cb975a98552	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2149b0a4d11a	019f09e3-5452-71ba-a465-4cb975a98552	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-214ad9dac23d	019f09e3-5452-71ba-a465-4cba035441e9	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-214bf6e60199	019f09e3-5452-71ba-a465-4cba035441e9	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-214c0da17736	019f09e3-5452-71ba-a465-4cbbad56b0d2	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-214d0a4df235	019f09e3-5452-71ba-a465-4cbbad56b0d2	019f09e3-547b-718d-83ce-0d572de50d3a
019f09e3-547c-7f93-876c-214e5d419765	019f09e3-5452-71ba-a465-4cbbad56b0d2	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-214fb70d68db	019f09e3-5452-71ba-a465-4cbca84dbaf8	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-21509013c529	019f09e3-5452-71ba-a465-4cbca84dbaf8	019f09e3-547b-718d-83ce-0d56d3c60e62
019f09e3-547c-7f93-876c-21519fcfc119	019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2152ad866e8e	019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-547b-718d-83ce-0d572de50d3a
019f09e3-547c-7f93-876c-2153f70ceaaf	019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-215487c9f57c	019f09e3-5452-71ba-a465-4cbe4b6bb815	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2155f004da82	019f09e3-5452-71ba-a465-4cbe4b6bb815	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2156070f9336	019f09e3-5452-71ba-a465-4cbe4b6bb815	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-215759383fd0	019f09e3-5452-71ba-a465-4cbf63802fc8	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21587f43b5cd	019f09e3-5452-71ba-a465-4cbf63802fc8	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-215944649dd6	019f09e3-5452-71ba-a465-4cbf63802fc8	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-215a12feadf3	019f09e3-5452-71ba-a465-4cc0a14f2663	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-215b8061a24e	019f09e3-5452-71ba-a465-4cc0a14f2663	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-215c4f4acd38	019f09e3-5452-71ba-a465-4cc0a14f2663	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-215dea6f046a	019f09e3-5452-71ba-a465-4cc10d44dadd	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-215e1c5277b5	019f09e3-5452-71ba-a465-4cc10d44dadd	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-215f41233028	019f09e3-5452-71ba-a465-4cc10d44dadd	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-2160b0c557f2	019f09e3-5452-71ba-a465-4cc259ec69c6	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2161af2e87fe	019f09e3-5452-71ba-a465-4cc259ec69c6	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-2162c920a376	019f09e3-5452-71ba-a465-4cc259ec69c6	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-216323cc11db	019f09e3-5452-71ba-a465-4cc259ec69c6	019f09e3-547b-718d-83ce-0d540ab6ef89
019f09e3-547c-7f93-876c-2164659e22d5	019f09e3-5452-71ba-a465-4cc3f23c86a5	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-2165913dc5f5	019f09e3-5452-71ba-a465-4cc3f23c86a5	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-216632c87788	019f09e3-5452-71ba-a465-4cc3f23c86a5	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-21679b112c8d	019f09e3-5452-71ba-a465-4cc4d9541029	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-21681ee8c94d	019f09e3-5452-71ba-a465-4cc4d9541029	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-21693ef5a16c	019f09e3-5452-71ba-a465-4cc4d9541029	019f09e3-547b-718d-83ce-0d537d0716c4
019f09e3-547c-7f93-876c-216ac88a849d	019f09e3-5452-71ba-a465-4cc4d9541029	019f09e3-547b-718d-83ce-0d55710497d8
019f09e3-547c-7f93-876c-216bf7b39b07	019f09e3-5452-71ba-a465-4cc5e71a257a	019f09e3-547b-718d-83ce-0d528ced4a5b
019f09e3-547c-7f93-876c-216cc6abfd66	019f09e3-5452-71ba-a465-4cc5e71a257a	019f09e3-547b-718d-83ce-0d51318992e8
019f09e3-547c-7f93-876c-216d41142d71	019f09e3-5452-71ba-a465-4cc5e71a257a	019f09e3-547b-718d-83ce-0d537d0716c4
\.


--
-- Data for Name: menu_item_modifier_groups; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.menu_item_modifier_groups (id, menu_item_id, modifier_group_id, is_required, sort_order) FROM stdin;
019f09e3-5454-7561-b9ec-e2546a0fc4af	019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-5454-7561-b9ec-e23a59310200	t	1
019f09e3-5454-7561-b9ec-e2557dfb6b32	019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-5454-7561-b9ec-e23b70b0296f	t	2
019f09e3-5454-7561-b9ec-e2561825ed11	019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-5454-7561-b9ec-e23c02b9bcd1	f	3
019f09e3-5454-7561-b9ec-e25735e34b25	019f09e3-5452-71ba-a465-4ca93d66db4f	019f09e3-5454-7561-b9ec-e23a59310200	t	1
019f09e3-5454-7561-b9ec-e2580172171d	019f09e3-5452-71ba-a465-4ca93d66db4f	019f09e3-5454-7561-b9ec-e23b70b0296f	t	2
019f09e3-5454-7561-b9ec-e2596e7a301e	019f09e3-5452-71ba-a465-4caa0b18f629	019f09e3-5454-7561-b9ec-e23a59310200	t	1
019f09e3-5454-7561-b9ec-e25a2df6db80	019f09e3-5452-71ba-a465-4caa0b18f629	019f09e3-5454-7561-b9ec-e23b70b0296f	t	2
019f09e3-5454-7561-b9ec-e25bcfe10c96	019f09e3-5452-71ba-a465-4cacba4db301	019f09e3-5454-7561-b9ec-e23a59310200	t	1
019f09e3-5454-7561-b9ec-e25c2cde7b42	019f09e3-5452-71ba-a465-4cacba4db301	019f09e3-5454-7561-b9ec-e23b70b0296f	t	2
019f09e3-5454-7561-b9ec-e25db1a77723	019f09e3-5452-71ba-a465-4cacba4db301	019f09e3-5454-7561-b9ec-e23c02b9bcd1	f	3
019f09e3-5454-7561-b9ec-e25ea0636b07	019f09e3-5452-71ba-a465-4caf98b840ae	019f09e3-5454-7561-b9ec-e23c02b9bcd1	f	1
019f09e3-5454-7561-b9ec-e25f55034142	019f09e3-5452-71ba-a465-4cb975a98552	019f09e3-5454-7561-b9ec-e23d32b6401c	t	1
019f09e3-5454-7561-b9ec-e260979e66a7	019f09e3-5452-71ba-a465-4cb975a98552	019f09e3-5454-7561-b9ec-e23e6fad945f	f	2
019f09e3-5454-7561-b9ec-e26130fa18e4	019f09e3-5452-71ba-a465-4cba035441e9	019f09e3-5454-7561-b9ec-e23d32b6401c	t	1
019f09e3-5454-7561-b9ec-e2626362a0b5	019f09e3-5452-71ba-a465-4cba035441e9	019f09e3-5454-7561-b9ec-e23e6fad945f	f	2
019f09e3-5454-7561-b9ec-e2637546a848	019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-5454-7561-b9ec-e23d32b6401c	t	1
019f09e3-5454-7561-b9ec-e26400122470	019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-5454-7561-b9ec-e23e6fad945f	f	2
019f09e3-5454-7561-b9ec-e2654018353e	019f09e3-5452-71ba-a465-4cb5e0a557d4	019f09e3-5454-7561-b9ec-e23e6fad945f	f	1
019f09e3-5454-7561-b9ec-e2665f765389	019f09e3-5452-71ba-a465-4cb7aa36e904	019f09e3-5454-7561-b9ec-e23e6fad945f	f	1
\.


--
-- Data for Name: menu_item_variants; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.menu_item_variants (id, menu_item_id, name, price_delta, is_default, sort_order) FROM stdin;
019f09e3-5453-7883-87e8-c1d6b0a70bd0	019f09e3-5452-71ba-a465-4ca8c71c9b3d	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1d7d664f7e7	019f09e3-5452-71ba-a465-4ca8c71c9b3d	Large	60.00	f	2
019f09e3-5453-7883-87e8-c1d87e6ac04a	019f09e3-5452-71ba-a465-4ca93d66db4f	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1d98d0acfc5	019f09e3-5452-71ba-a465-4ca93d66db4f	Large	50.00	f	2
019f09e3-5453-7883-87e8-c1daa2f69a1b	019f09e3-5452-71ba-a465-4caa0b18f629	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1db061810f1	019f09e3-5452-71ba-a465-4caa0b18f629	Large	50.00	f	2
019f09e3-5453-7883-87e8-c1dc2bf1e05b	019f09e3-5452-71ba-a465-4cacba4db301	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1dd88cb1729	019f09e3-5452-71ba-a465-4cacba4db301	Large	60.00	f	2
019f09e3-5453-7883-87e8-c1def14d6e12	019f09e3-5452-71ba-a465-4cae707ee024	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1dfb4716043	019f09e3-5452-71ba-a465-4caf98b840ae	Regular	0.00	t	1
019f09e3-5453-7883-87e8-c1e07cdc3759	019f09e3-5452-71ba-a465-4caf98b840ae	Large	80.00	f	2
019f09e3-5453-7883-87e8-c1e1f78f488c	019f09e3-5452-71ba-a465-4cbe4b6bb815	Slice	0.00	t	1
019f09e3-5453-7883-87e8-c1e25cb59e0a	019f09e3-5452-71ba-a465-4cbe4b6bb815	Whole Cake	1800.00	f	2
019f09e3-5453-7883-87e8-c1e3720dd043	019f09e3-5452-71ba-a465-4cc3f23c86a5	Slice	0.00	t	1
019f09e3-5453-7883-87e8-c1e4c060f3b4	019f09e3-5452-71ba-a465-4cc3f23c86a5	Whole Cake	2200.00	f	2
\.


--
-- Data for Name: menu_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.menu_items (id, category_id, name, slug, description, base_price, is_vegetarian, is_vegan, is_signature, calories, prep_time_minutes, image_url, is_available, tags, created_at, updated_at) FROM stdin;
019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-5451-78e5-8697-6b61c642cfbe	Emerald Mocha	emerald-mocha	House signature mocha with dark Belgian chocolate.	380.00	t	f	t	320	6	https://cdn.example.com/menu/emerald-mocha.jpg	t	["hot", "chocolate"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4ca93d66db4f	019f09e3-5451-78e5-8697-6b61c642cfbe	Flat White	flat-white	Double shot espresso with steamed milk.	290.00	t	f	f	180	5	https://cdn.example.com/menu/flat-white.jpg	t	["hot"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4caa0b18f629	019f09e3-5451-78e5-8697-6b61c642cfbe	Cappuccino	cappuccino	Classic cappuccino with thick foam.	280.00	t	f	f	160	5	https://cdn.example.com/menu/cappuccino.jpg	t	["hot"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cabdf313210	019f09e3-5451-78e5-8697-6b61c642cfbe	Long Black	long-black	Double espresso over hot water.	240.00	t	t	f	5	4	https://cdn.example.com/menu/long-black.jpg	t	["hot", "strong"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cacba4db301	019f09e3-5451-78e5-8697-6b6233c3e284	Iced Latte	iced-latte	Cold espresso latte over ice.	320.00	t	f	f	200	4	https://cdn.example.com/menu/iced-latte.jpg	t	["cold"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cad92ca4061	019f09e3-5451-78e5-8697-6b6233c3e284	Lemon Mint Fizz	lemon-mint-fizz	Sparkling lemon and mint refresher.	220.00	t	t	f	90	3	https://cdn.example.com/menu/lemon-mint-fizz.jpg	t	["cold", "refreshing"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cae707ee024	019f09e3-5451-78e5-8697-6b6233c3e284	Matcha Tonic	matcha-tonic	Ceremonial matcha with tonic and citrus.	360.00	t	t	t	110	4	https://cdn.example.com/menu/matcha-tonic.jpg	t	["cold", "matcha"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4caf98b840ae	019f09e3-5451-78e5-8697-6b6233c3e284	Mango Smoothie	mango-smoothie	Alphonso mango blended with yogurt.	340.00	t	f	f	280	4	https://cdn.example.com/menu/mango-smoothie.jpg	t	["cold", "fruit"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb00757b39f	019f09e3-5451-78e5-8697-6b6355f074b0	Bakery Special Croissant	bakery-special-croissant	Buttery 27-layer croissant.	180.00	t	f	t	380	2	https://cdn.example.com/menu/bakery-special-croissant.jpg	t	["baked"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb119d7d975	019f09e3-5451-78e5-8697-6b6355f074b0	Pain au Chocolat	pain-au-chocolat	Buttery pastry with dark chocolate batons.	220.00	t	f	f	410	2	https://cdn.example.com/menu/pain-au-chocolat.jpg	t	["baked", "chocolate"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb29c358ed7	019f09e3-5451-78e5-8697-6b6355f074b0	Almond Danish	almond-danish	Flaky danish topped with toasted almonds.	240.00	t	f	f	420	2	https://cdn.example.com/menu/almond-danish.jpg	t	["baked"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb31033def5	019f09e3-5451-78e5-8697-6b6355f074b0	Cheese Twist	cheese-twist	Savory puff pastry twist with cheddar.	160.00	t	f	f	280	2	https://cdn.example.com/menu/cheese-twist.jpg	t	["baked", "savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb45101ddbf	019f09e3-5451-78e5-8697-6b5fef33aa8e	Spinach Quiche	spinach-quiche	Classic spinach and feta quiche.	320.00	t	f	t	450	8	https://cdn.example.com/menu/spinach-quiche.jpg	t	["savory", "baked"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb5e0a557d4	019f09e3-5451-78e5-8697-6b5fef33aa8e	Avocado Toast	avocado-toast	Sourdough toast with smashed avocado and chilli flakes.	380.00	t	t	f	360	7	https://cdn.example.com/menu/avocado-toast.jpg	t	["savory", "healthy"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb671a7e228	019f09e3-5451-78e5-8697-6b5fef33aa8e	Shakshuka	shakshuka	Eggs poached in spiced tomato sauce.	460.00	t	f	t	480	12	https://cdn.example.com/menu/shakshuka.jpg	t	["savory", "spicy"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb7aa36e904	019f09e3-5451-78e5-8697-6b5fef33aa8e	Classic Omelette	classic-omelette	Three-egg omelette with cheese and herbs.	320.00	t	f	f	380	8	https://cdn.example.com/menu/classic-omelette.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb898c932f7	019f09e3-5451-78e5-8697-6b5fef33aa8e	Banana Pancakes	banana-pancakes	Stack of banana pancakes with maple syrup.	360.00	t	f	f	520	10	https://cdn.example.com/menu/banana-pancakes.jpg	t	["sweet"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cb975a98552	019f09e3-5451-78e5-8697-6b64759768ec	Emerald Club	club-sandwich	Triple-decker chicken club with bacon.	480.00	f	f	t	620	9	https://cdn.example.com/menu/club-sandwich.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cba035441e9	019f09e3-5451-78e5-8697-6b64759768ec	Caprese Panini	caprese-panini	Mozzarella, tomato, basil pesto on ciabatta.	420.00	t	f	f	520	8	https://cdn.example.com/menu/caprese-panini.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cbbad56b0d2	019f09e3-5451-78e5-8697-6b64759768ec	Smoked Salmon Bagel	smoked-salmon-bagel	Cream cheese, smoked salmon, capers, red onion.	520.00	f	f	f	560	6	https://cdn.example.com/menu/smoked-salmon-bagel.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cbca84dbaf8	019f09e3-5451-78e5-8697-6b64759768ec	Falafel Wrap	falafel-wrap	Crispy falafel with tahini sauce and pickled veg.	380.00	t	t	f	480	7	https://cdn.example.com/menu/falafel-wrap.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cbd5bbe94ff	019f09e3-5451-78e5-8697-6b64759768ec	Tuna Melt	tuna-melt	Tuna salad and melted cheddar on sourdough.	440.00	f	f	f	580	8	https://cdn.example.com/menu/tuna-melt.jpg	t	["savory"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cbe4b6bb815	019f09e3-5451-78e5-8697-6b60fe4455a9	Red Velvet Slice	red-velvet-slice	Classic red velvet with cream cheese frosting.	280.00	t	f	f	460	1	https://cdn.example.com/menu/red-velvet-slice.jpg	t	["sweet"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cbf63802fc8	019f09e3-5451-78e5-8697-6b60fe4455a9	Tiramisu Cup	tiramisu-cup	Espresso-soaked ladyfingers with mascarpone.	320.00	t	f	t	410	1	https://cdn.example.com/menu/tiramisu-cup.jpg	t	["sweet", "coffee"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc0a14f2663	019f09e3-5451-78e5-8697-6b60fe4455a9	Lemon Tart	lemon-tart	Tangy lemon curd in a buttery shortcrust shell.	300.00	t	f	f	380	1	https://cdn.example.com/menu/lemon-tart.jpg	t	["sweet", "citrus"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc10d44dadd	019f09e3-5451-78e5-8697-6b60fe4455a9	Chocolate Fudge Slice	chocolate-fudge-slice	Dense fudgy chocolate slice with ganache.	320.00	t	f	f	510	1	https://cdn.example.com/menu/chocolate-fudge-slice.jpg	t	["sweet", "chocolate"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc259ec69c6	019f09e3-5451-78e5-8697-6b60fe4455a9	Carrot Walnut Cake	carrot-walnut-cake	Spiced carrot cake with walnuts and cream cheese frosting.	290.00	t	f	f	470	1	https://cdn.example.com/menu/carrot-walnut-cake.jpg	t	["sweet"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc3f23c86a5	019f09e3-5451-78e5-8697-6b60fe4455a9	New York Cheesecake	cheesecake-newyork	Dense New York-style cheesecake with graham crust.	340.00	t	f	t	540	1	https://cdn.example.com/menu/cheesecake-newyork.jpg	t	["sweet"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc4d9541029	019f09e3-5451-78e5-8697-6b6355f074b0	Chocolate Chip Muffin	chocolate-chip-muffin	Classic muffin loaded with chocolate chips.	180.00	t	f	f	380	1	https://cdn.example.com/menu/chocolate-chip-muffin.jpg	t	["baked", "chocolate"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
019f09e3-5452-71ba-a465-4cc5e71a257a	019f09e3-5451-78e5-8697-6b6355f074b0	Blueberry Scone	blueberry-scone	Crumbly scone studded with blueberries.	200.00	t	f	f	340	1	https://cdn.example.com/menu/blueberry-scone.jpg	t	["baked"]	2026-06-27 16:22:01.042908+00	2026-06-27 16:22:01.042908+00
\.


--
-- Data for Name: modifier_groups; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.modifier_groups (id, name, selection_min, selection_max, sort_order) FROM stdin;
019f09e3-5454-7561-b9ec-e23a59310200	Milk Choice	1	1	1
019f09e3-5454-7561-b9ec-e23b70b0296f	Sugar Level	1	1	2
019f09e3-5454-7561-b9ec-e23c02b9bcd1	Add-ons	0	3	3
019f09e3-5454-7561-b9ec-e23d32b6401c	Sandwich Bread	1	1	4
019f09e3-5454-7561-b9ec-e23e6fad945f	Toppings	0	4	5
\.


--
-- Data for Name: modifiers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.modifiers (id, modifier_group_id, name, price_delta, is_default) FROM stdin;
019f09e3-5454-7561-b9ec-e23f09bcd4da	019f09e3-5454-7561-b9ec-e23a59310200	Whole Milk	0.00	t
019f09e3-5454-7561-b9ec-e240fdf1bb9c	019f09e3-5454-7561-b9ec-e23a59310200	Skim Milk	0.00	f
019f09e3-5454-7561-b9ec-e241435c627a	019f09e3-5454-7561-b9ec-e23a59310200	Oat Milk	40.00	f
019f09e3-5454-7561-b9ec-e242c6ba66cf	019f09e3-5454-7561-b9ec-e23a59310200	Almond Milk	40.00	f
019f09e3-5454-7561-b9ec-e24339a724a1	019f09e3-5454-7561-b9ec-e23a59310200	Soy Milk	30.00	f
019f09e3-5454-7561-b9ec-e244d080c56a	019f09e3-5454-7561-b9ec-e23b70b0296f	No Sugar	0.00	f
019f09e3-5454-7561-b9ec-e245f863c8a3	019f09e3-5454-7561-b9ec-e23b70b0296f	Less Sugar	0.00	f
019f09e3-5454-7561-b9ec-e246746b0790	019f09e3-5454-7561-b9ec-e23b70b0296f	Normal	0.00	t
019f09e3-5454-7561-b9ec-e247a763ef02	019f09e3-5454-7561-b9ec-e23b70b0296f	Extra Sugar	0.00	f
019f09e3-5454-7561-b9ec-e2487464db13	019f09e3-5454-7561-b9ec-e23c02b9bcd1	Extra Shot	60.00	f
019f09e3-5454-7561-b9ec-e249e118fb36	019f09e3-5454-7561-b9ec-e23c02b9bcd1	Whipped Cream	30.00	f
019f09e3-5454-7561-b9ec-e24a193a7d3b	019f09e3-5454-7561-b9ec-e23c02b9bcd1	Vanilla Syrup	25.00	f
019f09e3-5454-7561-b9ec-e24ba9643512	019f09e3-5454-7561-b9ec-e23c02b9bcd1	Caramel Drizzle	30.00	f
019f09e3-5454-7561-b9ec-e24cb71f9431	019f09e3-5454-7561-b9ec-e23d32b6401c	Sourdough	0.00	t
019f09e3-5454-7561-b9ec-e24de77071b7	019f09e3-5454-7561-b9ec-e23d32b6401c	Whole Wheat	0.00	f
019f09e3-5454-7561-b9ec-e24ef4ee1b1b	019f09e3-5454-7561-b9ec-e23d32b6401c	Ciabatta	20.00	f
019f09e3-5454-7561-b9ec-e24f6431082e	019f09e3-5454-7561-b9ec-e23d32b6401c	Gluten-Free	60.00	f
019f09e3-5454-7561-b9ec-e2508b900b27	019f09e3-5454-7561-b9ec-e23e6fad945f	Avocado	80.00	f
019f09e3-5454-7561-b9ec-e25154f8222c	019f09e3-5454-7561-b9ec-e23e6fad945f	Bacon	70.00	f
019f09e3-5454-7561-b9ec-e252faa8690e	019f09e3-5454-7561-b9ec-e23e6fad945f	Extra Cheese	50.00	f
019f09e3-5454-7561-b9ec-e25304864f87	019f09e3-5454-7561-b9ec-e23e6fad945f	Sundried Tomato	40.00	f
\.


--
-- Data for Name: notifications; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.notifications (id, customer_id, channel, kind, title, body, payload, read_at, sent_at, created_at) FROM stdin;
019f09e3-5474-7537-ac81-54322ae30a16	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01049 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b81c37a0871d"}	\N	2025-10-03 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5433ae0b731b	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01049 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b81c37a0871d"}	\N	2025-10-03 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5434eb0ac33b	019f09e3-5442-7d10-91b2-be1a53eca627	sms	order_status	Order ED-01048 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b81b4b9dc427"}	2025-10-04 21:37:00+00	2025-10-04 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54359f00a99d	019f09e3-5442-7d10-91b2-be1a53eca627	push	order_status	Order ED-01048 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b81b4b9dc427"}	\N	2025-10-04 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5436ca1e5941	019f09e3-5442-7d10-91b2-be195b3c14e2	email	order_status	Order ED-01047 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b81aa71679ba"}	2025-10-06 02:25:00+00	2025-10-06 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5437f4227456	019f09e3-5442-7d10-91b2-be195b3c14e2	push	order_status	Order ED-01047 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b81aa71679ba"}	\N	2025-10-06 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5438d5c60cf9	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01046 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b8199c81e149"}	2025-10-07 07:13:00+00	2025-10-07 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543914f9c1e4	019f09e3-5442-7d10-91b2-be17f929eb2d	sms	order_status	Order ED-01045 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b818f71c48da"}	\N	2025-10-08 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543ae69a5c39	019f09e3-5442-7d10-91b2-be25179006d0	email	order_status	Order ED-01044 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b8171cace121"}	2025-10-09 16:49:00+00	2025-10-09 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543b4808a5dd	019f09e3-5442-7d10-91b2-be24727af8e6	push	order_status	Order ED-01043 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b816bd1873be"}	2025-10-10 21:37:00+00	2025-10-10 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543c3a97cf2e	019f09e3-5442-7d10-91b2-be233ab206bc	sms	order_status	Order ED-01042 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b815aa76d847"}	2025-10-12 02:25:00+00	2025-10-12 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543d84d63df0	019f09e3-5442-7d10-91b2-be2261956b32	email	order_status	Order ED-01041 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b814d7d9057f"}	\N	2025-10-13 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543ec816e701	019f09e3-5442-7d10-91b2-be2144da3493	push	order_status	Order ED-01040 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b813f82db22f"}	2025-10-14 12:01:00+00	2025-10-14 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-543f34906ebe	019f09e3-5442-7d10-91b2-be2064b523b7	sms	order_status	Order ED-01039 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b812214b2b56"}	2025-10-15 16:49:00+00	2025-10-15 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5440430f42b1	019f09e3-5442-7d10-91b2-be1f5114d14b	email	order_status	Order ED-01038 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b811f2b5171b"}	2025-10-16 21:37:00+00	2025-10-16 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5441fc4cf388	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01037 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b8101276b16a"}	\N	2025-10-18 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5442cd0596c4	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01037 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b8101276b16a"}	\N	2025-10-18 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544353ef986f	019f09e3-5442-7d10-91b2-be1d7fed74b3	sms	order_status	Order ED-01036 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80f0832fbf9"}	2025-10-19 07:13:00+00	2025-10-19 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5444c832813c	019f09e3-5442-7d10-91b2-be1d7fed74b3	push	order_status	Order ED-01036 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80f0832fbf9"}	\N	2025-10-19 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544599456d39	019f09e3-5442-7d10-91b2-be1c99a7d8f2	email	order_status	Order ED-01035 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80e0158a4d7"}	2025-10-20 12:01:00+00	2025-10-20 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5446818e60a9	019f09e3-5442-7d10-91b2-be1c99a7d8f2	push	order_status	Order ED-01035 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80e0158a4d7"}	\N	2025-10-20 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544711922b1c	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01034 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80d12f1d6de"}	2025-10-21 16:49:00+00	2025-10-21 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5448fd263a0e	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01034 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80d12f1d6de"}	\N	2025-10-21 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5449448e7dd7	019f09e3-5442-7d10-91b2-be1a53eca627	sms	order_status	Order ED-01033 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80cd473ce94"}	\N	2025-10-22 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544ab6c49ec3	019f09e3-5442-7d10-91b2-be1a53eca627	push	order_status	Order ED-01033 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80cd473ce94"}	\N	2025-10-22 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544b85d1df6f	019f09e3-5442-7d10-91b2-be195b3c14e2	email	order_status	Order ED-01032 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80baa631ff0"}	2025-10-24 02:25:00+00	2025-10-24 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544cfcc20fa8	019f09e3-5442-7d10-91b2-be195b3c14e2	push	order_status	Order ED-01032 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80baa631ff0"}	\N	2025-10-24 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544d4de0f91e	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01031 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80ae19741d1"}	2025-10-25 07:13:00+00	2025-10-25 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544ef7e5a833	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01031 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80ae19741d1"}	\N	2025-10-25 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-544f87e7a463	019f09e3-5442-7d10-91b2-be17f929eb2d	sms	order_status	Order ED-01030 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b809709e73e2"}	2025-10-26 12:01:00+00	2025-10-26 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5450b2326c28	019f09e3-5442-7d10-91b2-be17f929eb2d	push	order_status	Order ED-01030 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b809709e73e2"}	\N	2025-10-26 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54519f610688	019f09e3-5442-7d10-91b2-be25179006d0	email	order_status	Order ED-01029 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b808a56cb42c"}	\N	2025-10-27 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54529e1efe43	019f09e3-5442-7d10-91b2-be25179006d0	push	order_status	Order ED-01029 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b808a56cb42c"}	\N	2025-10-27 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5453d9ae09af	019f09e3-5442-7d10-91b2-be24727af8e6	push	order_status	Order ED-01028 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80788ff8136"}	2025-10-28 21:37:00+00	2025-10-28 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545481916f78	019f09e3-5442-7d10-91b2-be24727af8e6	push	order_status	Order ED-01028 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80788ff8136"}	\N	2025-10-28 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54558f4cb019	019f09e3-5442-7d10-91b2-be233ab206bc	sms	order_status	Order ED-01027 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80628f307ca"}	2025-10-30 02:25:00+00	2025-10-30 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5456d20645d4	019f09e3-5442-7d10-91b2-be233ab206bc	push	order_status	Order ED-01027 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80628f307ca"}	\N	2025-10-30 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5457449c11c3	019f09e3-5442-7d10-91b2-be2261956b32	email	order_status	Order ED-01026 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b80576bf5eb5"}	2025-10-31 07:13:00+00	2025-10-31 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5458c6b46eb8	019f09e3-5442-7d10-91b2-be2261956b32	push	order_status	Order ED-01026 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b80576bf5eb5"}	\N	2025-10-31 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5459d02aa2d9	019f09e3-5442-7d10-91b2-be2144da3493	push	order_status	Order ED-01025 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b804031f3f4c"}	\N	2025-11-01 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545aa0fbb945	019f09e3-5442-7d10-91b2-be2144da3493	push	order_status	Order ED-01025 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b804031f3f4c"}	\N	2025-11-01 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545bb710e900	019f09e3-5442-7d10-91b2-be2064b523b7	sms	order_status	Order ED-01024 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b8035a6701f6"}	2025-11-02 16:49:00+00	2025-11-02 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545c6ac5608e	019f09e3-5442-7d10-91b2-be2064b523b7	push	order_status	Order ED-01024 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b8035a6701f6"}	\N	2025-11-02 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545daa1511f6	019f09e3-5442-7d10-91b2-be1f5114d14b	email	order_status	Order ED-01023 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b802ec565457"}	2025-11-03 21:37:00+00	2025-11-03 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545e9dc9894e	019f09e3-5442-7d10-91b2-be1f5114d14b	push	order_status	Order ED-01023 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b802ec565457"}	\N	2025-11-03 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-545f7cf41e78	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01022 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b801d964b111"}	2025-11-05 02:25:00+00	2025-11-05 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5460eb69deae	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01022 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b801d964b111"}	\N	2025-11-05 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5461b5bc3d7d	019f09e3-5442-7d10-91b2-be1d7fed74b3	sms	order_status	Order ED-01021 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b8007301815b"}	\N	2025-11-06 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5462847fb5b2	019f09e3-5442-7d10-91b2-be1d7fed74b3	push	order_status	Order ED-01021 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b8007301815b"}	\N	2025-11-06 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5463a12910b3	019f09e3-5442-7d10-91b2-be1c99a7d8f2	email	order_status	Order ED-01020 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7ff0900f2e9"}	2025-11-07 12:01:00+00	2025-11-07 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546408ffa708	019f09e3-5442-7d10-91b2-be1c99a7d8f2	push	order_status	Order ED-01020 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7ff0900f2e9"}	\N	2025-11-07 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546555f6ed41	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01019 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7fe32fb109f"}	2025-11-08 16:49:00+00	2025-11-08 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5466795c037c	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01019 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7fe32fb109f"}	\N	2025-11-08 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54674e5d0dea	019f09e3-5442-7d10-91b2-be1a53eca627	sms	order_status	Order ED-01018 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7fd427a8c56"}	2025-11-09 21:37:00+00	2025-11-09 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546875b50eee	019f09e3-5442-7d10-91b2-be1a53eca627	push	order_status	Order ED-01018 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7fd427a8c56"}	\N	2025-11-09 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546976c29bca	019f09e3-5442-7d10-91b2-be195b3c14e2	email	order_status	Order ED-01017 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7fc7dd1ca8a"}	\N	2025-11-11 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546ad0720cb7	019f09e3-5442-7d10-91b2-be195b3c14e2	push	order_status	Order ED-01017 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7fc7dd1ca8a"}	\N	2025-11-11 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546bc16955f3	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01016 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7fb9a3a04fb"}	2025-11-12 07:13:00+00	2025-11-12 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546cc1d5f35c	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01016 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7fb9a3a04fb"}	\N	2025-11-12 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546d895eea31	019f09e3-5442-7d10-91b2-be17f929eb2d	sms	order_status	Order ED-01015 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7fab9e7a570"}	2025-11-13 12:01:00+00	2025-11-13 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546e02042fbb	019f09e3-5442-7d10-91b2-be17f929eb2d	push	order_status	Order ED-01015 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7fab9e7a570"}	\N	2025-11-13 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-546f435adc66	019f09e3-5442-7d10-91b2-be25179006d0	email	order_status	Order ED-01014 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f99c6a23a9"}	2025-11-14 16:49:00+00	2025-11-14 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547086123d81	019f09e3-5442-7d10-91b2-be25179006d0	push	order_status	Order ED-01014 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f99c6a23a9"}	\N	2025-11-14 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54716be6a29d	019f09e3-5442-7d10-91b2-be24727af8e6	push	order_status	Order ED-01013 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f85bf8f694"}	\N	2025-11-15 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5472903f0efc	019f09e3-5442-7d10-91b2-be24727af8e6	push	order_status	Order ED-01013 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f85bf8f694"}	\N	2025-11-15 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547320cd7cb0	019f09e3-5442-7d10-91b2-be233ab206bc	sms	order_status	Order ED-01012 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f7f360daf9"}	2025-11-17 02:25:00+00	2025-11-17 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547419e4930f	019f09e3-5442-7d10-91b2-be233ab206bc	push	order_status	Order ED-01012 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f7f360daf9"}	\N	2025-11-17 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547591c67a6a	019f09e3-5442-7d10-91b2-be2261956b32	email	order_status	Order ED-01011 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f60c423fb4"}	2025-11-18 07:13:00+00	2025-11-18 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54766e0150aa	019f09e3-5442-7d10-91b2-be2261956b32	push	order_status	Order ED-01011 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f60c423fb4"}	\N	2025-11-18 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54778bf6787d	019f09e3-5442-7d10-91b2-be2144da3493	push	order_status	Order ED-01010 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f58cde71d0"}	2025-11-19 12:01:00+00	2025-11-19 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5478ef4a191a	019f09e3-5442-7d10-91b2-be2144da3493	push	order_status	Order ED-01010 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f58cde71d0"}	\N	2025-11-19 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547971686710	019f09e3-5442-7d10-91b2-be2064b523b7	sms	order_status	Order ED-01009 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f4248ea60f"}	\N	2025-11-20 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547a4ae1f1da	019f09e3-5442-7d10-91b2-be2064b523b7	push	order_status	Order ED-01009 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f4248ea60f"}	\N	2025-11-20 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547b479b2273	019f09e3-5442-7d10-91b2-be1f5114d14b	email	order_status	Order ED-01008 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f3b082e1c7"}	2025-11-21 21:37:00+00	2025-11-21 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547c15d6a513	019f09e3-5442-7d10-91b2-be1f5114d14b	push	order_status	Order ED-01008 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f3b082e1c7"}	\N	2025-11-21 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547d26f58a3d	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01007 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f26644452d"}	2025-11-23 02:25:00+00	2025-11-23 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547e85bce603	019f09e3-5442-7d10-91b2-be1e19f149d9	push	order_status	Order ED-01007 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f26644452d"}	\N	2025-11-23 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-547f66b9e70e	019f09e3-5442-7d10-91b2-be1d7fed74b3	sms	order_status	Order ED-01006 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f140dd23d3"}	2025-11-24 07:13:00+00	2025-11-24 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-548027bba2d0	019f09e3-5442-7d10-91b2-be1d7fed74b3	push	order_status	Order ED-01006 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f140dd23d3"}	\N	2025-11-24 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5481669bc55c	019f09e3-5442-7d10-91b2-be1c99a7d8f2	email	order_status	Order ED-01005 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7f060508c6b"}	\N	2025-11-25 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54829d407113	019f09e3-5442-7d10-91b2-be1c99a7d8f2	push	order_status	Order ED-01005 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7f060508c6b"}	\N	2025-11-25 12:55:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54832f72c97f	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01004 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7efd97a1689"}	2025-11-26 16:49:00+00	2025-11-26 16:48:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54843805b31b	019f09e3-5442-7d10-91b2-be1b7dad8028	push	order_status	Order ED-01004 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7efd97a1689"}	\N	2025-11-26 17:43:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54858089640f	019f09e3-5442-7d10-91b2-be1a53eca627	sms	order_status	Order ED-01003 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7eecd8ef115"}	2025-11-27 21:37:00+00	2025-11-27 21:36:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5486b55c58e3	019f09e3-5442-7d10-91b2-be1a53eca627	push	order_status	Order ED-01003 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7eecd8ef115"}	\N	2025-11-27 22:31:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5487fdf18816	019f09e3-5442-7d10-91b2-be195b3c14e2	email	order_status	Order ED-01002 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7ed59eb2dd4"}	2025-11-29 02:25:00+00	2025-11-29 02:24:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-54884ee2b21e	019f09e3-5442-7d10-91b2-be195b3c14e2	push	order_status	Order ED-01002 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7ed59eb2dd4"}	\N	2025-11-29 03:19:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-5489e9db659e	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01001 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7ecd578786e"}	\N	2025-11-30 07:12:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-548a03633a71	019f09e3-5442-7d10-91b2-be18ecd018b1	push	order_status	Order ED-01001 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7ecd578786e"}	\N	2025-11-30 08:07:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-548b8c15d267	019f09e3-5442-7d10-91b2-be17f929eb2d	sms	order_status	Order ED-01000 placed	Your order has been placed.	{"status": "placed", "order_id": "019f09e3-5457-70fb-820c-b7eb090f2dc5"}	2025-12-01 12:01:00+00	2025-12-01 12:00:00+00	2026-06-27 16:22:01.076807+00
019f09e3-5474-7537-ac81-548cbcb52860	019f09e3-5442-7d10-91b2-be17f929eb2d	push	order_status	Order ED-01000 delivered	Enjoy your meal!	{"status": "delivered", "order_id": "019f09e3-5457-70fb-820c-b7eb090f2dc5"}	\N	2025-12-01 12:55:00+00	2026-06-27 16:22:01.076807+00
\.


--
-- Data for Name: order_item_modifiers; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.order_item_modifiers (id, order_item_id, modifier_id, snapshot_name, snapshot_price_delta) FROM stdin;
019f09e3-545a-7b49-8131-bb2566890eb2	019f09e3-5459-7040-b27c-fa8d2368147c	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545a-7b49-8131-bb26327b84b7	019f09e3-5459-7040-b27c-fa8d2368147c	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545a-7b49-8131-bb27cc96a397	019f09e3-5459-7040-b27c-fa8e8a101fdb	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f034eef558e0	019f09e3-5459-7040-b27c-fa8e8a101fdb	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f035da4db97e	019f09e3-5459-7040-b27c-fa929b78a0e8	019f09e3-5454-7561-b9ec-e24ef4ee1b1b	Ciabatta	20.00
019f09e3-545b-7fae-9450-f036c05d45b4	019f09e3-5459-7040-b27c-fa96df3ccfda	019f09e3-5454-7561-b9ec-e24339a724a1	Soy Milk	30.00
019f09e3-545b-7fae-9450-f0372ed9cb17	019f09e3-5459-7040-b27c-fa96df3ccfda	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f038c3915204	019f09e3-5459-7040-b27c-fa9a9e4215cb	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f039d268b6db	019f09e3-5459-7040-b27c-fa9c6c99f5ce	019f09e3-5454-7561-b9ec-e242c6ba66cf	Almond Milk	40.00
019f09e3-545b-7fae-9450-f03a3801b5d6	019f09e3-5459-7040-b27c-fa9c6c99f5ce	019f09e3-5454-7561-b9ec-e244d080c56a	No Sugar	0.00
019f09e3-545b-7fae-9450-f03b38113cf9	019f09e3-5459-7040-b27c-fa9d825c3fb1	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f03c1dcc33d8	019f09e3-5459-7040-b27c-faa4e0748a9e	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f03d575a378a	019f09e3-5459-7040-b27c-faa4e0748a9e	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f03ef019ec57	019f09e3-5459-7040-b27c-faa7cb24297d	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f03f4a7c4688	019f09e3-5459-7040-b27c-faa7cb24297d	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f040899fdc16	019f09e3-5459-7040-b27c-faa966cc1858	019f09e3-5454-7561-b9ec-e24ef4ee1b1b	Ciabatta	20.00
019f09e3-545b-7fae-9450-f0416cfff37c	019f09e3-5459-7040-b27c-faaf3bf80c8d	019f09e3-5454-7561-b9ec-e24de77071b7	Whole Wheat	0.00
019f09e3-545b-7fae-9450-f042e0e6dd0a	019f09e3-5459-7040-b27c-fab56f7106b1	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f0430c2814b5	019f09e3-5459-7040-b27c-fab63b4da1fb	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f044e90486cd	019f09e3-5459-7040-b27c-fab9d8e2ae9c	019f09e3-5454-7561-b9ec-e241435c627a	Oat Milk	40.00
019f09e3-545b-7fae-9450-f045dde9e7a5	019f09e3-5459-7040-b27c-fab9d8e2ae9c	019f09e3-5454-7561-b9ec-e245f863c8a3	Less Sugar	0.00
019f09e3-545b-7fae-9450-f046bae5404c	019f09e3-5459-7040-b27c-fabe37e22d9d	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f04793964780	019f09e3-5459-7040-b27c-fabf6cb1fcb8	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f048d2f04bfa	019f09e3-5459-7040-b27c-fabf6cb1fcb8	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f049141587fb	019f09e3-5459-7040-b27c-fac005ecf0c2	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f04adcde3a19	019f09e3-5459-7040-b27c-fac005ecf0c2	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f04bdfc13a6f	019f09e3-5459-7040-b27c-fac4224d329a	019f09e3-5454-7561-b9ec-e24ef4ee1b1b	Ciabatta	20.00
019f09e3-545b-7fae-9450-f04cbc0cf179	019f09e3-5459-7040-b27c-fac86322d833	019f09e3-5454-7561-b9ec-e24339a724a1	Soy Milk	30.00
019f09e3-545b-7fae-9450-f04d16ce5f9b	019f09e3-5459-7040-b27c-fac86322d833	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f04ec3eda99f	019f09e3-5459-7040-b27c-facc40d22b07	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f04ffe8d2088	019f09e3-5459-7040-b27c-face787d575c	019f09e3-5454-7561-b9ec-e242c6ba66cf	Almond Milk	40.00
019f09e3-545b-7fae-9450-f05077736b9a	019f09e3-5459-7040-b27c-face787d575c	019f09e3-5454-7561-b9ec-e244d080c56a	No Sugar	0.00
019f09e3-545b-7fae-9450-f0517481bc8a	019f09e3-5459-7040-b27c-facf6c9b7471	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f052e2618a13	019f09e3-5459-7040-b27c-fad6abe7f68c	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f05333692f94	019f09e3-5459-7040-b27c-fad6abe7f68c	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f054d03c43ad	019f09e3-5459-7040-b27c-fad9f550395f	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f0559ad58ce7	019f09e3-5459-7040-b27c-fad9f550395f	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f056509c01b3	019f09e3-5459-7040-b27c-fadbca4c9864	019f09e3-5454-7561-b9ec-e24ef4ee1b1b	Ciabatta	20.00
019f09e3-545b-7fae-9450-f057071dcdf0	019f09e3-5459-7040-b27c-fae1279478e3	019f09e3-5454-7561-b9ec-e24de77071b7	Whole Wheat	0.00
019f09e3-545b-7fae-9450-f058b99f63b9	019f09e3-5459-7040-b27c-fae7f4f420d8	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f059e2a3aa01	019f09e3-5459-7040-b27c-fae89e1d6ec5	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f05a21fa70aa	019f09e3-5459-7040-b27c-faeb4d6957fc	019f09e3-5454-7561-b9ec-e241435c627a	Oat Milk	40.00
019f09e3-545b-7fae-9450-f05b98e87c09	019f09e3-5459-7040-b27c-faeb4d6957fc	019f09e3-5454-7561-b9ec-e245f863c8a3	Less Sugar	0.00
019f09e3-545b-7fae-9450-f05c2709875d	019f09e3-5459-7040-b27c-faf08765f04e	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f05df2bb908b	019f09e3-5459-7040-b27c-faf1c1342866	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f05e69c13061	019f09e3-5459-7040-b27c-faf1c1342866	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f05fea40b36a	019f09e3-5459-7040-b27c-faf298a8aabb	019f09e3-5454-7561-b9ec-e23f09bcd4da	Whole Milk	0.00
019f09e3-545b-7fae-9450-f060bd598f23	019f09e3-5459-7040-b27c-faf298a8aabb	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f0616cc87aa9	019f09e3-5459-7040-b27c-faf6c825c02b	019f09e3-5454-7561-b9ec-e24ef4ee1b1b	Ciabatta	20.00
019f09e3-545b-7fae-9450-f0621ff2c355	019f09e3-5459-7040-b27c-fafaf1f3a31e	019f09e3-5454-7561-b9ec-e24339a724a1	Soy Milk	30.00
019f09e3-545b-7fae-9450-f063138d67d9	019f09e3-5459-7040-b27c-fafaf1f3a31e	019f09e3-5454-7561-b9ec-e246746b0790	Normal	0.00
019f09e3-545b-7fae-9450-f064513f0273	019f09e3-5459-7040-b27c-fafef2d57292	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
019f09e3-545b-7fae-9450-f0658c11a155	019f09e3-5459-7040-b27c-fb008bdc1b9a	019f09e3-5454-7561-b9ec-e242c6ba66cf	Almond Milk	40.00
019f09e3-545b-7fae-9450-f0666249c5bc	019f09e3-5459-7040-b27c-fb008bdc1b9a	019f09e3-5454-7561-b9ec-e244d080c56a	No Sugar	0.00
019f09e3-545b-7fae-9450-f067611d16ab	019f09e3-5459-7040-b27c-fb019c8c2211	019f09e3-5454-7561-b9ec-e24cb71f9431	Sourdough	0.00
\.


--
-- Data for Name: order_items; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.order_items (id, order_id, menu_item_id, variant_id, quantity, unit_price, line_total, snapshot_name, created_at) FROM stdin;
019f09e3-5459-7040-b27c-fa8d2368147c	019f09e3-5457-70fb-820c-b7eb090f2dc5	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	380.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa8e8a101fdb	019f09e3-5457-70fb-820c-b7ecd578786e	019f09e3-5452-71ba-a465-4cabdf313210	\N	1	240.00	240.00	Long Black	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa8fbd117ecf	019f09e3-5457-70fb-820c-b7ecd578786e	019f09e3-5452-71ba-a465-4cb29c358ed7	\N	2	240.00	480.00	Almond Danish	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa90795598cf	019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	360.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9178517a5e	019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5452-71ba-a465-4cb5e0a557d4	\N	2	380.00	760.00	Avocado Toast	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa929b78a0e8	019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5452-71ba-a465-4cbca84dbaf8	\N	1	400.00	400.00	Falafel Wrap	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9366fd8cd2	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5452-71ba-a465-4cb119d7d975	\N	1	220.00	220.00	Pain au Chocolat	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9448facf1a	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5452-71ba-a465-4cb898c932f7	\N	2	360.00	720.00	Banana Pancakes	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa958184a09d	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5452-71ba-a465-4cbf63802fc8	\N	1	320.00	320.00	Tiramisu Cup	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa96df3ccfda	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	2	410.00	820.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa973af0a3ed	019f09e3-5457-70fb-820c-b7efd97a1689	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	320.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa988bc7719d	019f09e3-5457-70fb-820c-b7f060508c6b	019f09e3-5452-71ba-a465-4cb7aa36e904	\N	1	320.00	320.00	Classic Omelette	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa99f17be061	019f09e3-5457-70fb-820c-b7f060508c6b	019f09e3-5452-71ba-a465-4cbe4b6bb815	\N	2	280.00	560.00	Red Velvet Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9a9e4215cb	019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5452-71ba-a465-4cba035441e9	\N	1	420.00	420.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9bf7f53e55	019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5452-71ba-a465-4cc10d44dadd	\N	2	320.00	640.00	Chocolate Fudge Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9c6c99f5ce	019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5452-71ba-a465-4caa0b18f629	\N	1	320.00	320.00	Cappuccino	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9d825c3fb1	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	1	440.00	440.00	Tuna Melt	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9eaff9739d	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5452-71ba-a465-4cc4d9541029	\N	2	180.00	360.00	Chocolate Chip Muffin	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fa9fdc1b127d	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5452-71ba-a465-4cad92ca4061	\N	1	220.00	220.00	Lemon Mint Fizz	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa0ad8aca75	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	2	320.00	640.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa14e5845af	019f09e3-5457-70fb-820c-b7f3b082e1c7	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	1	300.00	300.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa2129dc305	019f09e3-5457-70fb-820c-b7f4248ea60f	019f09e3-5452-71ba-a465-4cc3f23c86a5	\N	1	340.00	340.00	New York Cheesecake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa3dc8eeff8	019f09e3-5457-70fb-820c-b7f4248ea60f	019f09e3-5452-71ba-a465-4cacba4db301	\N	2	320.00	640.00	Iced Latte	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa4e0748a9e	019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	380.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa5f6f9dd44	019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5452-71ba-a465-4caf98b840ae	\N	2	340.00	680.00	Mango Smoothie	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa664f9cd56	019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5452-71ba-a465-4cb671a7e228	\N	1	460.00	460.00	Shakshuka	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa7cb24297d	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5452-71ba-a465-4cabdf313210	\N	1	240.00	240.00	Long Black	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa8ad98f4cd	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5452-71ba-a465-4cb29c358ed7	\N	2	240.00	480.00	Almond Danish	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faa966cc1858	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5452-71ba-a465-4cb975a98552	\N	1	500.00	500.00	Emerald Club	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faaac86b21c4	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	2	300.00	600.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faab6a706da0	019f09e3-5457-70fb-820c-b7f7f360daf9	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	360.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faacfb7bf8fb	019f09e3-5457-70fb-820c-b7f85bf8f694	019f09e3-5452-71ba-a465-4cb119d7d975	\N	1	220.00	220.00	Pain au Chocolat	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faad66f16a10	019f09e3-5457-70fb-820c-b7f85bf8f694	019f09e3-5452-71ba-a465-4cb898c932f7	\N	2	360.00	720.00	Banana Pancakes	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faae52058229	019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	320.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faaf3bf80c8d	019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5452-71ba-a465-4cbbad56b0d2	\N	2	520.00	1040.00	Smoked Salmon Bagel	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab0eaed7773	019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5452-71ba-a465-4cc259ec69c6	\N	1	290.00	290.00	Carrot Walnut Cake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab1bf01e8d3	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5452-71ba-a465-4cb7aa36e904	\N	1	320.00	320.00	Classic Omelette	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab26a2150ab	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5452-71ba-a465-4cbe4b6bb815	\N	2	280.00	560.00	Red Velvet Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab35b3dc4f3	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5452-71ba-a465-4cc5e71a257a	\N	1	200.00	200.00	Blueberry Scone	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab4f5713ee7	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5452-71ba-a465-4cae707ee024	\N	2	360.00	720.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab56f7106b1	019f09e3-5457-70fb-820c-b7fb9a3a04fb	019f09e3-5452-71ba-a465-4cba035441e9	\N	1	420.00	420.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab63b4da1fb	019f09e3-5457-70fb-820c-b7fc7dd1ca8a	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	1	440.00	440.00	Tuna Melt	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab708b6e4fb	019f09e3-5457-70fb-820c-b7fc7dd1ca8a	019f09e3-5452-71ba-a465-4cc4d9541029	\N	2	180.00	360.00	Chocolate Chip Muffin	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab8cd4758ae	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	1	300.00	300.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fab9d8e2ae9c	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5452-71ba-a465-4ca93d66db4f	\N	2	330.00	660.00	Flat White	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faba3c0f4c10	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5452-71ba-a465-4cb00757b39f	\N	1	180.00	180.00	Bakery Special Croissant	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fabbae6c3aca	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5452-71ba-a465-4cc3f23c86a5	\N	1	340.00	340.00	New York Cheesecake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fabc2707192e	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5452-71ba-a465-4cacba4db301	\N	2	320.00	640.00	Iced Latte	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fabdecd719c6	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5452-71ba-a465-4cb31033def5	\N	1	160.00	160.00	Cheese Twist	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fabe37e22d9d	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5452-71ba-a465-4cba035441e9	\N	2	420.00	840.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fabf6cb1fcb8	019f09e3-5457-70fb-820c-b7ff0900f2e9	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	380.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac005ecf0c2	019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5452-71ba-a465-4cabdf313210	\N	1	240.00	240.00	Long Black	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac1d44372a7	019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5452-71ba-a465-4cb29c358ed7	\N	2	240.00	480.00	Almond Danish	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac2a84bde47	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	360.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac3e7b1e90c	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5452-71ba-a465-4cb5e0a557d4	\N	2	380.00	760.00	Avocado Toast	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac4224d329a	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5452-71ba-a465-4cbca84dbaf8	\N	1	400.00	400.00	Falafel Wrap	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac536efb16a	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5452-71ba-a465-4cb119d7d975	\N	1	220.00	220.00	Pain au Chocolat	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac6373acf92	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5452-71ba-a465-4cb898c932f7	\N	2	360.00	720.00	Banana Pancakes	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac74ad0c5f5	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5452-71ba-a465-4cbf63802fc8	\N	1	320.00	320.00	Tiramisu Cup	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac86322d833	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	2	410.00	820.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fac938eb5c71	019f09e3-5457-70fb-820c-b8035a6701f6	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	320.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faca55d70544	019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5452-71ba-a465-4cb7aa36e904	\N	1	320.00	320.00	Classic Omelette	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-facbf04cc317	019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5452-71ba-a465-4cbe4b6bb815	\N	2	280.00	560.00	Red Velvet Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-facc40d22b07	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5452-71ba-a465-4cba035441e9	\N	1	420.00	420.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-facd7ee3426f	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5452-71ba-a465-4cc10d44dadd	\N	2	320.00	640.00	Chocolate Fudge Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-face787d575c	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5452-71ba-a465-4caa0b18f629	\N	1	320.00	320.00	Cappuccino	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-facf6c9b7471	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	1	440.00	440.00	Tuna Melt	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad08fca51a3	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5452-71ba-a465-4cc4d9541029	\N	2	180.00	360.00	Chocolate Chip Muffin	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad185ae1938	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5452-71ba-a465-4cad92ca4061	\N	1	220.00	220.00	Lemon Mint Fizz	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad2a3c4e0b8	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	2	320.00	640.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad3cb410a17	019f09e3-5457-70fb-820c-b80788ff8136	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	1	300.00	300.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad4dd5bfba0	019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5452-71ba-a465-4cc3f23c86a5	\N	1	340.00	340.00	New York Cheesecake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad5409cea82	019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5452-71ba-a465-4cacba4db301	\N	2	320.00	640.00	Iced Latte	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad6abe7f68c	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	380.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad7136ac3cf	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5452-71ba-a465-4caf98b840ae	\N	2	340.00	680.00	Mango Smoothie	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad8d0e3a847	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5452-71ba-a465-4cb671a7e228	\N	1	460.00	460.00	Shakshuka	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fad9f550395f	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5452-71ba-a465-4cabdf313210	\N	1	240.00	240.00	Long Black	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fada7847307f	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5452-71ba-a465-4cb29c358ed7	\N	2	240.00	480.00	Almond Danish	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fadbca4c9864	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5452-71ba-a465-4cb975a98552	\N	1	500.00	500.00	Emerald Club	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fadcbbcc1ca4	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	2	300.00	600.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fadd15449301	019f09e3-5457-70fb-820c-b80baa631ff0	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	360.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fade0358fb13	019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5452-71ba-a465-4cb119d7d975	\N	1	220.00	220.00	Pain au Chocolat	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fadf0da39bbf	019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5452-71ba-a465-4cb898c932f7	\N	2	360.00	720.00	Banana Pancakes	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae0a9a9a694	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	320.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae1279478e3	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5452-71ba-a465-4cbbad56b0d2	\N	2	520.00	1040.00	Smoked Salmon Bagel	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae26202cdd4	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5452-71ba-a465-4cc259ec69c6	\N	1	290.00	290.00	Carrot Walnut Cake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae3efe128eb	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5452-71ba-a465-4cb7aa36e904	\N	1	320.00	320.00	Classic Omelette	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae46b083200	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5452-71ba-a465-4cbe4b6bb815	\N	2	280.00	560.00	Red Velvet Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae5d323ed07	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5452-71ba-a465-4cc5e71a257a	\N	1	200.00	200.00	Blueberry Scone	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae66714c499	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5452-71ba-a465-4cae707ee024	\N	2	360.00	720.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae7f4f420d8	019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5452-71ba-a465-4cba035441e9	\N	1	420.00	420.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae89e1d6ec5	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	1	440.00	440.00	Tuna Melt	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fae9bd2ce2da	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5452-71ba-a465-4cc4d9541029	\N	2	180.00	360.00	Chocolate Chip Muffin	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faead698551f	019f09e3-5457-70fb-820c-b811f2b5171b	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	1	300.00	300.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faeb4d6957fc	019f09e3-5457-70fb-820c-b811f2b5171b	019f09e3-5452-71ba-a465-4ca93d66db4f	\N	2	330.00	660.00	Flat White	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faec132e9189	019f09e3-5457-70fb-820c-b811f2b5171b	019f09e3-5452-71ba-a465-4cb00757b39f	\N	1	180.00	180.00	Bakery Special Croissant	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faedeff88dbd	019f09e3-5457-70fb-820c-b812214b2b56	019f09e3-5452-71ba-a465-4cc3f23c86a5	\N	1	340.00	340.00	New York Cheesecake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faee59536869	019f09e3-5457-70fb-820c-b812214b2b56	019f09e3-5452-71ba-a465-4cacba4db301	\N	2	320.00	640.00	Iced Latte	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faef28866876	019f09e3-5457-70fb-820c-b812214b2b56	019f09e3-5452-71ba-a465-4cb31033def5	\N	1	160.00	160.00	Cheese Twist	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf08765f04e	019f09e3-5457-70fb-820c-b812214b2b56	019f09e3-5452-71ba-a465-4cba035441e9	\N	2	420.00	840.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf1c1342866	019f09e3-5457-70fb-820c-b813f82db22f	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	1	380.00	380.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf298a8aabb	019f09e3-5457-70fb-820c-b814d7d9057f	019f09e3-5452-71ba-a465-4cabdf313210	\N	1	240.00	240.00	Long Black	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf3807bebc9	019f09e3-5457-70fb-820c-b814d7d9057f	019f09e3-5452-71ba-a465-4cb29c358ed7	\N	2	240.00	480.00	Almond Danish	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf4830b4827	019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5452-71ba-a465-4cae707ee024	\N	1	360.00	360.00	Matcha Tonic	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf5b3e1b060	019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5452-71ba-a465-4cb5e0a557d4	\N	2	380.00	760.00	Avocado Toast	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf6c825c02b	019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5452-71ba-a465-4cbca84dbaf8	\N	1	400.00	400.00	Falafel Wrap	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf7fff3cb4b	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5452-71ba-a465-4cb119d7d975	\N	1	220.00	220.00	Pain au Chocolat	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf868fa058c	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5452-71ba-a465-4cb898c932f7	\N	2	360.00	720.00	Banana Pancakes	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faf9d996e106	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5452-71ba-a465-4cbf63802fc8	\N	1	320.00	320.00	Tiramisu Cup	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fafaf1f3a31e	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5452-71ba-a465-4ca8c71c9b3d	\N	2	410.00	820.00	Emerald Mocha	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fafbd955ae04	019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	1	320.00	320.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fafc2e3bfb77	019f09e3-5457-70fb-820c-b818f71c48da	019f09e3-5452-71ba-a465-4cb7aa36e904	\N	1	320.00	320.00	Classic Omelette	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fafd6d4c8d47	019f09e3-5457-70fb-820c-b818f71c48da	019f09e3-5452-71ba-a465-4cbe4b6bb815	\N	2	280.00	560.00	Red Velvet Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fafef2d57292	019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5452-71ba-a465-4cba035441e9	\N	1	420.00	420.00	Caprese Panini	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-faff5e036586	019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5452-71ba-a465-4cc10d44dadd	\N	2	320.00	640.00	Chocolate Fudge Slice	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb008bdc1b9a	019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5452-71ba-a465-4caa0b18f629	\N	1	320.00	320.00	Cappuccino	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb019c8c2211	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	1	440.00	440.00	Tuna Melt	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb022489a063	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5452-71ba-a465-4cc4d9541029	\N	2	180.00	360.00	Chocolate Chip Muffin	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb037fb3e153	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5452-71ba-a465-4cad92ca4061	\N	1	220.00	220.00	Lemon Mint Fizz	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb04ed57e19d	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	2	320.00	640.00	Spinach Quiche	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb05ec4f6d82	019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5452-71ba-a465-4cc0a14f2663	\N	1	300.00	300.00	Lemon Tart	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb0673384e8e	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5452-71ba-a465-4cc3f23c86a5	\N	1	340.00	340.00	New York Cheesecake	2026-06-27 16:22:01.049947+00
019f09e3-5459-7040-b27c-fb0775e8a434	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5452-71ba-a465-4cacba4db301	\N	2	320.00	640.00	Iced Latte	2026-06-27 16:22:01.049947+00
\.


--
-- Data for Name: order_promotions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.order_promotions (id, order_id, promo_code_id, customer_id, discount_amount, applied_at) FROM stdin;
019f09e3-5467-7b12-b7b1-d1c3d2f4096a	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5465-7e2b-9a2e-a885390a08db	019f09e3-5442-7d10-91b2-be1b7dad8028	50.00	2025-10-03 16:48:00+00
019f09e3-5467-7b12-b7b1-d1c4bb4b0f4f	019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5465-7e2b-9a2e-a886236e2b18	019f09e3-5442-7d10-91b2-be1a53eca627	30.00	2025-10-04 21:36:00+00
019f09e3-5467-7b12-b7b1-d1c531d17ff1	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5465-7e2b-9a2e-a887bda9571a	019f09e3-5442-7d10-91b2-be195b3c14e2	0.00	2025-10-06 02:24:00+00
019f09e3-5467-7b12-b7b1-d1c6fcf57c82	019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5465-7e2b-9a2e-a8887c20703b	019f09e3-5442-7d10-91b2-be18ecd018b1	200.00	2025-10-07 07:12:00+00
019f09e3-5467-7b12-b7b1-d1c74e313d9d	019f09e3-5457-70fb-820c-b818f71c48da	019f09e3-5465-7e2b-9a2e-a889fe191ce4	019f09e3-5442-7d10-91b2-be17f929eb2d	100.00	2025-10-08 12:00:00+00
019f09e3-5467-7b12-b7b1-d1c82067ac54	019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5465-7e2b-9a2e-a885390a08db	019f09e3-5442-7d10-91b2-be25179006d0	50.00	2025-10-09 16:48:00+00
019f09e3-5467-7b12-b7b1-d1c9a09be869	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5465-7e2b-9a2e-a886236e2b18	019f09e3-5442-7d10-91b2-be24727af8e6	100.00	2025-10-10 21:36:00+00
019f09e3-5467-7b12-b7b1-d1cae6f3150d	019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5465-7e2b-9a2e-a887bda9571a	019f09e3-5442-7d10-91b2-be233ab206bc	0.00	2025-10-12 02:24:00+00
019f09e3-5467-7b12-b7b1-d1cb524a4ac2	019f09e3-5457-70fb-820c-b814d7d9057f	019f09e3-5465-7e2b-9a2e-a8887c20703b	019f09e3-5442-7d10-91b2-be2261956b32	144.00	2025-10-13 07:12:00+00
019f09e3-5467-7b12-b7b1-d1cc9c8e1536	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5465-7e2b-9a2e-a889fe191ce4	019f09e3-5442-7d10-91b2-be1e19f149d9	100.00	2025-10-18 02:24:00+00
019f09e3-5467-7b12-b7b1-d1cd5c3d994d	019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5465-7e2b-9a2e-a885390a08db	019f09e3-5442-7d10-91b2-be1d7fed74b3	50.00	2025-10-19 07:12:00+00
019f09e3-5467-7b12-b7b1-d1cefbca2ded	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5465-7e2b-9a2e-a886236e2b18	019f09e3-5442-7d10-91b2-be1c99a7d8f2	100.00	2025-10-20 12:00:00+00
\.


--
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.orders (id, customer_id, address_id, order_number, status, placed_at, subtotal, delivery_fee, discount_total, tax, grand_total, payment_method, scheduled_for, customer_notes, internal_notes, created_at, updated_at, location_id) FROM stdin;
019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5443-7fbe-b1a5-b1128f3e4441	ED-01003	delivered	2025-11-27 21:36:00+00	2080.00	60.00	0.00	104.00	2244.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5443-7fbe-b1a5-b11d86170386	ED-01011	delivered	2025-11-18 07:12:00+00	1820.00	60.00	0.00	91.00	1971.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7f3b082e1c7	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5443-7fbe-b1a5-b11961535e61	ED-01008	delivered	2025-11-21 21:36:00+00	300.00	60.00	0.00	15.00	375.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7f7f360daf9	019f09e3-5442-7d10-91b2-be233ab206bc	019f09e3-5443-7fbe-b1a5-b11ea3efbb3d	ED-01012	delivered	2025-11-17 02:24:00+00	360.00	60.00	0.00	18.00	438.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7f85bf8f694	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5443-7fbe-b1a5-b1205cad0de2	ED-01013	delivered	2025-11-15 21:36:00+00	940.00	60.00	0.00	47.00	1047.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5443-7fbe-b1a5-b1210ab70e3b	ED-01014	delivered	2025-11-14 16:48:00+00	1650.00	60.00	0.00	83.00	1793.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7fb9a3a04fb	019f09e3-5442-7d10-91b2-be18ecd018b1	019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	ED-01016	delivered	2025-11-12 07:12:00+00	420.00	60.00	0.00	21.00	501.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7fc7dd1ca8a	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5443-7fbe-b1a5-b111a8b8d367	ED-01017	delivered	2025-11-11 02:24:00+00	800.00	60.00	0.00	40.00	900.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5443-7fbe-b1a5-b1128f3e4441	ED-01018	delivered	2025-11-09 21:36:00+00	1140.00	60.00	0.00	57.00	1257.00	card	2025-11-10 00:36:00+00	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5443-7fbe-b1a5-b114d0308581	ED-01019	delivered	2025-11-08 16:48:00+00	1980.00	60.00	0.00	99.00	2139.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7ff0900f2e9	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5443-7fbe-b1a5-b1152464cd40	ED-01020	delivered	2025-11-07 12:00:00+00	380.00	60.00	0.00	19.00	459.00	cod	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5443-7fbe-b1a5-b11657ae506c	ED-01021	delivered	2025-11-06 07:12:00+00	720.00	60.00	0.00	36.00	816.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5443-7fbe-b1a5-b118417956c0	ED-01022	delivered	2025-11-05 02:24:00+00	1520.00	60.00	0.00	76.00	1656.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5443-7fbe-b1a5-b11961535e61	ED-01023	delivered	2025-11-03 21:36:00+00	2080.00	60.00	0.00	104.00	2244.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b8035a6701f6	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5443-7fbe-b1a5-b11aff50cbea	ED-01024	delivered	2025-11-02 16:48:00+00	320.00	60.00	0.00	16.00	396.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5443-7fbe-b1a5-b11d86170386	ED-01026	delivered	2025-10-31 07:12:00+00	1380.00	60.00	0.00	69.00	1509.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5442-7d10-91b2-be2144da3493	019f09e3-5443-7fbe-b1a5-b11cba8fdf04	ED-01025	delivered	2025-11-01 12:00:00+00	880.00	60.00	0.00	44.00	984.00	bkash	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b80788ff8136	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5443-7fbe-b1a5-b1205cad0de2	ED-01028	delivered	2025-10-28 21:36:00+00	300.00	60.00	0.00	15.00	375.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5443-7fbe-b1a5-b1210ab70e3b	ED-01029	delivered	2025-10-27 16:48:00+00	980.00	60.00	0.00	49.00	1089.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5442-7d10-91b2-be18ecd018b1	019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	ED-01031	delivered	2025-10-25 07:12:00+00	1820.00	60.00	0.00	91.00	1971.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b80baa631ff0	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5443-7fbe-b1a5-b111a8b8d367	ED-01032	delivered	2025-10-24 02:24:00+00	360.00	60.00	0.00	18.00	438.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5443-7fbe-b1a5-b1152464cd40	ED-01035	delivered	2025-10-20 12:00:00+00	1800.00	60.00	0.00	90.00	1950.00	cod	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5443-7fbe-b1a5-b1128f3e4441	ED-01033	delivered	2025-10-22 21:36:00+00	940.00	60.00	0.00	47.00	1047.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5443-7fbe-b1a5-b114d0308581	ED-01034	delivered	2025-10-21 16:48:00+00	1650.00	60.00	0.00	83.00	1793.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5443-7fbe-b1a5-b118417956c0	ED-01037	delivered	2025-10-18 02:24:00+00	800.00	60.00	0.00	40.00	900.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b811f2b5171b	019f09e3-5442-7d10-91b2-be1f5114d14b	019f09e3-5443-7fbe-b1a5-b11961535e61	ED-01038	cancelled	2025-10-16 21:36:00+00	1140.00	60.00	0.00	57.00	1257.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5442-7d10-91b2-be233ab206bc	019f09e3-5443-7fbe-b1a5-b11ea3efbb3d	ED-01042	preparing	2025-10-12 02:24:00+00	1520.00	60.00	0.00	76.00	1656.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b812214b2b56	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5443-7fbe-b1a5-b11aff50cbea	ED-01039	cancelled	2025-10-15 16:48:00+00	1980.00	60.00	0.00	99.00	2139.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b814d7d9057f	019f09e3-5442-7d10-91b2-be2261956b32	019f09e3-5443-7fbe-b1a5-b11d86170386	ED-01041	preparing	2025-10-13 07:12:00+00	720.00	60.00	0.00	36.00	816.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5442-7d10-91b2-be24727af8e6	019f09e3-5443-7fbe-b1a5-b1205cad0de2	ED-01043	dispatched	2025-10-10 21:36:00+00	2080.00	60.00	0.00	104.00	2244.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b813f82db22f	019f09e3-5442-7d10-91b2-be2144da3493	019f09e3-5443-7fbe-b1a5-b11cba8fdf04	ED-01040	cancelled	2025-10-14 12:00:00+00	380.00	60.00	0.00	19.00	459.00	bkash	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5443-7fbe-b1a5-b1210ab70e3b	ED-01044	dispatched	2025-10-09 16:48:00+00	320.00	60.00	0.00	16.00	396.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5442-7d10-91b2-be18ecd018b1	019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	ED-01046	placed	2025-10-07 07:12:00+00	1380.00	60.00	0.00	69.00	1509.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b818f71c48da	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5443-7fbe-b1a5-b10eaebf83c5	ED-01045	confirmed	2025-10-08 12:00:00+00	880.00	60.00	0.00	44.00	984.00	card	2025-10-08 15:00:00+00	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5443-7fbe-b1a5-b111a8b8d367	ED-01047	delivered	2025-10-06 02:24:00+00	1660.00	60.00	0.00	83.00	1803.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7eb090f2dc5	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5443-7fbe-b1a5-b10eaebf83c5	ED-01000	delivered	2025-12-01 12:00:00+00	380.00	60.00	0.00	19.00	459.00	card	2025-12-01 15:00:00+00	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7efd97a1689	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5443-7fbe-b1a5-b114d0308581	ED-01004	delivered	2025-11-26 16:48:00+00	320.00	60.00	0.00	16.00	396.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7ecd578786e	019f09e3-5442-7d10-91b2-be18ecd018b1	019f09e3-5443-7fbe-b1a5-b1109fb8d2cc	ED-01001	delivered	2025-11-30 07:12:00+00	720.00	60.00	0.00	36.00	816.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7f060508c6b	019f09e3-5442-7d10-91b2-be1c99a7d8f2	019f09e3-5443-7fbe-b1a5-b1152464cd40	ED-01005	delivered	2025-11-25 12:00:00+00	880.00	60.00	0.00	44.00	984.00	cod	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5442-7d10-91b2-be2144da3493	019f09e3-5443-7fbe-b1a5-b11cba8fdf04	ED-01010	delivered	2025-11-19 12:00:00+00	1520.00	60.00	0.00	76.00	1656.00	bkash	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5442-7d10-91b2-be1e19f149d9	019f09e3-5443-7fbe-b1a5-b118417956c0	ED-01007	delivered	2025-11-23 02:24:00+00	1660.00	60.00	0.00	83.00	1803.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5443-7fbe-b1a5-b11657ae506c	ED-01006	delivered	2025-11-24 07:12:00+00	1380.00	60.00	0.00	69.00	1509.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5442-7d10-91b2-be195b3c14e2	019f09e3-5443-7fbe-b1a5-b111a8b8d367	ED-01002	delivered	2025-11-29 02:24:00+00	1520.00	60.00	0.00	76.00	1656.00	cod	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000003-0000-7000-8000-000000000003
019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5443-7fbe-b1a5-b10eaebf83c5	ED-01015	delivered	2025-11-13 12:00:00+00	1800.00	60.00	0.00	90.00	1950.00	card	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b7f4248ea60f	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5443-7fbe-b1a5-b11aff50cbea	ED-01009	delivered	2025-11-20 16:48:00+00	980.00	60.00	0.00	49.00	1089.00	card	2025-11-20 19:48:00+00	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5442-7d10-91b2-be233ab206bc	019f09e3-5443-7fbe-b1a5-b11ea3efbb3d	ED-01027	delivered	2025-10-30 02:24:00+00	1660.00	60.00	0.00	83.00	1803.00	card	2025-10-30 05:24:00+00	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5443-7fbe-b1a5-b10eaebf83c5	ED-01030	delivered	2025-10-26 12:00:00+00	1520.00	60.00	0.00	76.00	1656.00	card	\N	Please pack with care.	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5443-7fbe-b1a5-b11657ae506c	ED-01036	delivered	2025-10-19 07:12:00+00	420.00	60.00	0.00	21.00	501.00	card	2025-10-19 10:12:00+00	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5443-7fbe-b1a5-b1128f3e4441	ED-01048	delivered	2025-10-04 21:36:00+00	300.00	60.00	0.00	15.00	375.00	card	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000001-0000-7000-8000-000000000001
019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5443-7fbe-b1a5-b114d0308581	ED-01049	delivered	2025-10-03 16:48:00+00	980.00	60.00	0.00	49.00	1089.00	bkash	\N	\N	\N	2026-06-27 16:22:01.048397+00	2026-06-27 16:22:01.048397+00	00000002-0000-7000-8000-000000000002
\.


--
-- Data for Name: payments; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.payments (id, order_id, customer_id, method, provider_reference, amount, status, captured_at, refunded_at, created_at) FROM stdin;
019f09e3-545b-7fae-9450-f0681caff0ce	019f09e3-5457-70fb-820c-b7eb090f2dc5	019f09e3-5442-7d10-91b2-be17f929eb2d	card	pay_000000	459.00	refunded	2025-12-01 12:00:30+00	2025-12-01 13:00:00+00	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f069c76b2791	019f09e3-5457-70fb-820c-b7ecd578786e	019f09e3-5442-7d10-91b2-be18ecd018b1	bkash	pay_000001	816.00	captured	2025-11-30 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06a82e9db60	019f09e3-5457-70fb-820c-b7ed59eb2dd4	019f09e3-5442-7d10-91b2-be195b3c14e2	cod	pay_000002	1656.00	captured	2025-11-29 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06b32b9fd29	019f09e3-5457-70fb-820c-b7eecd8ef115	019f09e3-5442-7d10-91b2-be1a53eca627	card	pay_000003	2244.00	captured	2025-11-27 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06c1a5beda2	019f09e3-5457-70fb-820c-b7efd97a1689	019f09e3-5442-7d10-91b2-be1b7dad8028	bkash	pay_000004	396.00	captured	2025-11-26 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06d4deb1ea6	019f09e3-5457-70fb-820c-b7f060508c6b	019f09e3-5442-7d10-91b2-be1c99a7d8f2	cod	pay_000005	984.00	captured	2025-11-25 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06ea52b188d	019f09e3-5457-70fb-820c-b7f140dd23d3	019f09e3-5442-7d10-91b2-be1d7fed74b3	card	pay_000006	1509.00	captured	2025-11-24 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f06fbd798a32	019f09e3-5457-70fb-820c-b7f26644452d	019f09e3-5442-7d10-91b2-be1e19f149d9	bkash	pay_000007	1803.00	captured	2025-11-23 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f070089c867f	019f09e3-5457-70fb-820c-b7f3b082e1c7	019f09e3-5442-7d10-91b2-be1f5114d14b	cod	pay_000008	375.00	captured	2025-11-21 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0712a05eb0c	019f09e3-5457-70fb-820c-b7f4248ea60f	019f09e3-5442-7d10-91b2-be2064b523b7	card	pay_000009	1089.00	captured	2025-11-20 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0725376e780	019f09e3-5457-70fb-820c-b7f58cde71d0	019f09e3-5442-7d10-91b2-be2144da3493	bkash	pay_000010	1656.00	captured	2025-11-19 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07324c20ca1	019f09e3-5457-70fb-820c-b7f60c423fb4	019f09e3-5442-7d10-91b2-be2261956b32	cod	pay_000011	1971.00	captured	2025-11-18 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07496b3574f	019f09e3-5457-70fb-820c-b7f7f360daf9	019f09e3-5442-7d10-91b2-be233ab206bc	card	pay_000012	438.00	captured	2025-11-17 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0757c195015	019f09e3-5457-70fb-820c-b7f85bf8f694	019f09e3-5442-7d10-91b2-be24727af8e6	bkash	pay_000013	1047.00	refunded	2025-11-15 21:36:30+00	2025-11-15 22:36:00+00	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0768ba08105	019f09e3-5457-70fb-820c-b7f99c6a23a9	019f09e3-5442-7d10-91b2-be25179006d0	cod	pay_000014	1793.00	captured	2025-11-14 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07710f521ad	019f09e3-5457-70fb-820c-b7fab9e7a570	019f09e3-5442-7d10-91b2-be17f929eb2d	card	pay_000015	1950.00	captured	2025-11-13 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07842d380e6	019f09e3-5457-70fb-820c-b7fb9a3a04fb	019f09e3-5442-7d10-91b2-be18ecd018b1	bkash	pay_000016	501.00	captured	2025-11-12 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0798a885bcf	019f09e3-5457-70fb-820c-b7fc7dd1ca8a	019f09e3-5442-7d10-91b2-be195b3c14e2	cod	pay_000017	900.00	captured	2025-11-11 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07a32da162f	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5442-7d10-91b2-be1a53eca627	card	pay_000018	1257.00	captured	2025-11-09 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07b849787b3	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5442-7d10-91b2-be1b7dad8028	bkash	pay_000019	2139.00	captured	2025-11-08 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07c21e7c5e5	019f09e3-5457-70fb-820c-b7ff0900f2e9	019f09e3-5442-7d10-91b2-be1c99a7d8f2	cod	pay_000020	459.00	captured	2025-11-07 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07d2f645a79	019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5442-7d10-91b2-be1d7fed74b3	card	pay_000021	816.00	captured	2025-11-06 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07ed0b159c8	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5442-7d10-91b2-be1e19f149d9	bkash	pay_000022	1656.00	captured	2025-11-05 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f07ff2d3e1b1	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5442-7d10-91b2-be1f5114d14b	cod	pay_000023	2244.00	captured	2025-11-03 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0803bf12c8b	019f09e3-5457-70fb-820c-b8035a6701f6	019f09e3-5442-7d10-91b2-be2064b523b7	card	pay_000024	396.00	captured	2025-11-02 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f081848180d6	019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5442-7d10-91b2-be2144da3493	bkash	pay_000025	984.00	captured	2025-11-01 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0823faaa65f	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5442-7d10-91b2-be2261956b32	cod	pay_000026	1509.00	refunded	2025-10-31 07:12:30+00	2025-10-31 08:12:00+00	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08391c26c27	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5442-7d10-91b2-be233ab206bc	card	pay_000027	1803.00	captured	2025-10-30 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f084ee663b58	019f09e3-5457-70fb-820c-b80788ff8136	019f09e3-5442-7d10-91b2-be24727af8e6	bkash	pay_000028	375.00	captured	2025-10-28 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f085bd6bbb27	019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5442-7d10-91b2-be25179006d0	cod	pay_000029	1089.00	captured	2025-10-27 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f086da954b09	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5442-7d10-91b2-be17f929eb2d	card	pay_000030	1656.00	captured	2025-10-26 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f087c74c4a7a	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5442-7d10-91b2-be18ecd018b1	bkash	pay_000031	1971.00	captured	2025-10-25 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08842063eb3	019f09e3-5457-70fb-820c-b80baa631ff0	019f09e3-5442-7d10-91b2-be195b3c14e2	cod	pay_000032	438.00	captured	2025-10-24 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f089503c4d3d	019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5442-7d10-91b2-be1a53eca627	card	pay_000033	1047.00	captured	2025-10-22 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08ae75d9c8e	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5442-7d10-91b2-be1b7dad8028	bkash	pay_000034	1793.00	captured	2025-10-21 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08b59caa736	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5442-7d10-91b2-be1c99a7d8f2	cod	pay_000035	1950.00	captured	2025-10-20 12:00:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08c3207eccb	019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5442-7d10-91b2-be1d7fed74b3	card	pay_000036	501.00	captured	2025-10-19 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08d37d0f7e9	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5442-7d10-91b2-be1e19f149d9	bkash	pay_000037	900.00	captured	2025-10-18 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08e9e2960d4	019f09e3-5457-70fb-820c-b814d7d9057f	019f09e3-5442-7d10-91b2-be2261956b32	cod	pay_000041	816.00	captured	2025-10-13 07:12:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f08f51b95587	019f09e3-5457-70fb-820c-b815aa76d847	019f09e3-5442-7d10-91b2-be233ab206bc	card	pay_000042	1656.00	captured	2025-10-12 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0907439c62f	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5442-7d10-91b2-be24727af8e6	bkash	pay_000043	2244.00	captured	2025-10-10 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f091ef8a5896	019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5442-7d10-91b2-be25179006d0	cod	pay_000044	396.00	captured	2025-10-09 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0922bb8a875	019f09e3-5457-70fb-820c-b818f71c48da	019f09e3-5442-7d10-91b2-be17f929eb2d	card	pay_000045	984.00	pending	\N	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f093f7dcf58b	019f09e3-5457-70fb-820c-b8199c81e149	019f09e3-5442-7d10-91b2-be18ecd018b1	bkash	pay_000046	1509.00	pending	\N	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f0948d846d12	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5442-7d10-91b2-be195b3c14e2	cod	pay_000047	1803.00	captured	2025-10-06 02:24:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f095ad282b0e	019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5442-7d10-91b2-be1a53eca627	card	pay_000048	375.00	captured	2025-10-04 21:36:30+00	\N	2026-06-27 16:22:01.052152+00
019f09e3-545b-7fae-9450-f096741987d7	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5442-7d10-91b2-be1b7dad8028	bkash	pay_000049	1089.00	captured	2025-10-03 16:48:30+00	\N	2026-06-27 16:22:01.052152+00
\.


--
-- Data for Name: promo_codes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.promo_codes (id, code, description, discount_type, discount_value, min_order_amount, max_discount, valid_from, valid_until, usage_limit, used_count, is_active, created_at) FROM stdin;
019f09e3-5465-7e2b-9a2e-a885390a08db	WELCOME50	Welcome offer — 50 BDT off	flat	50.00	200.00	\N	2025-09-02 12:00:00+00	2026-09-02 12:00:00+00	1000	0	t	2026-06-27 16:22:01.06129+00
019f09e3-5465-7e2b-9a2e-a886236e2b18	BREAKFAST10	10% off breakfast items	percent	10.00	300.00	100.00	2025-11-01 12:00:00+00	2025-12-31 12:00:00+00	500	0	t	2026-06-27 16:22:01.06129+00
019f09e3-5465-7e2b-9a2e-a887bda9571a	FREEDELIVERY	Free delivery on orders above 500	free_delivery	0.00	500.00	\N	2025-11-17 12:00:00+00	2025-12-17 12:00:00+00	2000	0	t	2026-06-27 16:22:01.06129+00
019f09e3-5465-7e2b-9a2e-a8887c20703b	CAKE20	20% off all cakes	percent	20.00	250.00	200.00	2025-11-24 12:00:00+00	2025-12-08 12:00:00+00	300	0	t	2026-06-27 16:22:01.06129+00
019f09e3-5465-7e2b-9a2e-a889fe191ce4	WEEKEND100	100 BDT off weekend orders	flat	100.00	600.00	\N	2025-11-28 12:00:00+00	2025-12-05 12:00:00+00	200	0	t	2026-06-27 16:22:01.06129+00
\.


--
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reviews (id, order_id, customer_id, menu_item_id, driver_id, rating, title, body, photos, is_published, response, responded_at, created_at) FROM stdin;
019f09e3-5470-7b06-b18a-497b3d0cc1f6	019f09e3-5457-70fb-820c-b81c37a0871d	019f09e3-5442-7d10-91b2-be1b7dad8028	019f09e3-5452-71ba-a465-4cc3f23c86a5	019f09e3-5456-7446-ba30-bd89b5b9f431	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	Thanks for the feedback!	2025-10-05 16:48:00+00	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d743e6230062	019f09e3-5457-70fb-820c-b81b4b9dc427	019f09e3-5442-7d10-91b2-be1a53eca627	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d744e3f44e40	019f09e3-5457-70fb-820c-b81aa71679ba	019f09e3-5442-7d10-91b2-be195b3c14e2	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d6e0106cab	019f09e3-5457-70fb-820c-b8171cace121	019f09e3-5442-7d10-91b2-be25179006d0	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d7fce72969	019f09e3-5457-70fb-820c-b816bd1873be	019f09e3-5442-7d10-91b2-be24727af8e6	\N	019f09e3-5456-7446-ba30-bd89b5b9f431	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d7452794e247	019f09e3-5457-70fb-820c-b8101276b16a	019f09e3-5442-7d10-91b2-be1e19f149d9	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-497d9e80ac0c	019f09e3-5457-70fb-820c-b80f0832fbf9	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5452-71ba-a465-4cba035441e9	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d746d69cd133	019f09e3-5457-70fb-820c-b80e0158a4d7	019f09e3-5442-7d10-91b2-be1c99a7d8f2	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d5937ddb69	019f09e3-5457-70fb-820c-b80d12f1d6de	019f09e3-5442-7d10-91b2-be1b7dad8028	\N	019f09e3-5456-7446-ba30-bd8c67a34936	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-497ca8ab500e	019f09e3-5457-70fb-820c-b80cd473ce94	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5452-71ba-a465-4cb119d7d975	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	Thanks for the feedback!	2025-10-24 21:36:00+00	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d7474147530f	019f09e3-5457-70fb-820c-b80baa631ff0	019f09e3-5442-7d10-91b2-be195b3c14e2	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d748403e1ff7	019f09e3-5457-70fb-820c-b80ae19741d1	019f09e3-5442-7d10-91b2-be18ecd018b1	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5471-7606-ba72-0c7dfce9ca02	019f09e3-5457-70fb-820c-b809709e73e2	019f09e3-5442-7d10-91b2-be17f929eb2d	019f09e3-5452-71ba-a465-4ca8c71c9b3d	019f09e3-5456-7446-ba30-bd8860c1a192	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d749b30c2989	019f09e3-5457-70fb-820c-b808a56cb42c	019f09e3-5442-7d10-91b2-be25179006d0	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d74a6b67c354	019f09e3-5457-70fb-820c-b80788ff8136	019f09e3-5442-7d10-91b2-be24727af8e6	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-497ebd41029c	019f09e3-5457-70fb-820c-b80628f307ca	019f09e3-5442-7d10-91b2-be233ab206bc	019f09e3-5452-71ba-a465-4cbd5bbe94ff	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-497a73a9ac81	019f09e3-5457-70fb-820c-b80576bf5eb5	019f09e3-5442-7d10-91b2-be2261956b32	\N	019f09e3-5456-7446-ba30-bd8a9394b803	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546e-71bd-9614-d74bd219dd53	019f09e3-5457-70fb-820c-b804031f3f4c	019f09e3-5442-7d10-91b2-be2144da3493	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-4980ed5f5ae5	019f09e3-5457-70fb-820c-b8035a6701f6	019f09e3-5442-7d10-91b2-be2064b523b7	019f09e3-5452-71ba-a465-4cb45101ddbf	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	Thanks for the feedback!	2025-11-04 16:48:00+00	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d26c252c37	019f09e3-5457-70fb-820c-b802ec565457	019f09e3-5442-7d10-91b2-be1f5114d14b	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5470-7b06-b18a-497fc031e4cb	019f09e3-5457-70fb-820c-b801d964b111	019f09e3-5442-7d10-91b2-be1e19f149d9	\N	019f09e3-5456-7446-ba30-bd8c67a34936	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5471-7606-ba72-0c7ce6834db7	019f09e3-5457-70fb-820c-b8007301815b	019f09e3-5442-7d10-91b2-be1d7fed74b3	019f09e3-5452-71ba-a465-4cabdf313210	\N	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d3ac3002c6	019f09e3-5457-70fb-820c-b7ff0900f2e9	019f09e3-5442-7d10-91b2-be1c99a7d8f2	\N	\N	4	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-546f-7aeb-8203-b1d46c094dc6	019f09e3-5457-70fb-820c-b7fe32fb109f	019f09e3-5442-7d10-91b2-be1b7dad8028	\N	\N	5	Great experience	Loved the food and quick delivery!	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
019f09e3-5471-7606-ba72-0c7b2c5c0df3	019f09e3-5457-70fb-820c-b7fd427a8c56	019f09e3-5442-7d10-91b2-be1a53eca627	019f09e3-5452-71ba-a465-4cc0a14f2663	019f09e3-5456-7446-ba30-bd8860c1a192	3	Could be better	Order was okay but a bit cold on arrival.	[]	t	\N	\N	2026-06-27 16:22:01.073806+00
\.


--
-- Name: bgw_job_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.bgw_job_id_seq', 1000, false);


--
-- Name: chunk_column_stats_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.chunk_column_stats_id_seq', 1, false);


--
-- Name: chunk_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.chunk_id_seq', 1, false);


--
-- Name: dimension_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.dimension_id_seq', 1, false);


--
-- Name: dimension_slice_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.dimension_slice_id_seq', 1, false);


--
-- Name: hypertable_id_seq; Type: SEQUENCE SET; Schema: _timescaledb_catalog; Owner: -
--

SELECT pg_catalog.setval('_timescaledb_catalog.hypertable_id_seq', 1, false);


--
-- Name: allergens allergens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.allergens
    ADD CONSTRAINT allergens_pkey PRIMARY KEY (id);


--
-- Name: allergens allergens_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.allergens
    ADD CONSTRAINT allergens_slug_key UNIQUE (slug);


--
-- Name: cafe_locations cafe_locations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafe_locations
    ADD CONSTRAINT cafe_locations_pkey PRIMARY KEY (id);


--
-- Name: cafe_locations cafe_locations_slug_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cafe_locations
    ADD CONSTRAINT cafe_locations_slug_key UNIQUE (slug);


--
-- Name: cart_items cart_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cart_items
    ADD CONSTRAINT cart_items_pkey PRIMARY KEY (id);


--
-- Name: carts carts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.carts
    ADD CONSTRAINT carts_pkey PRIMARY KEY (id);


--
-- Name: customer_addresses customer_addresses_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customer_addresses
    ADD CONSTRAINT customer_addresses_pkey PRIMARY KEY (id);


--
-- Name: customers customers_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_email_key UNIQUE (email);


--
-- Name: customers customers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.customers
    ADD CONSTRAINT customers_pkey PRIMARY KEY (id);


--
-- Name: deliveries deliveries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.deliveries
    ADD CONSTRAINT deliveries_pkey PRIMARY KEY (id);


--
-- Name: drivers drivers_phone_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_phone_key UNIQUE (phone);


--
-- Name: drivers drivers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.drivers
    ADD CONSTRAINT drivers_pkey PRIMARY KEY (id);


--
-- Name: kysely_migration_lock kysely_migration_lock_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kysely_migration_lock
    ADD CONSTRAINT kysely_migration_lock_pkey PRIMARY KEY (id);


--
-- Name: kysely_migration kysely_migration_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.kysely_migration
    ADD CONSTRAINT kysely_migration_pkey PRIMARY KEY (name);


--
-- Name: loyalty_accounts loyalty_accounts_customer_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT loyalty_accounts_customer_id_key UNIQUE (customer_id);


--
-- Name: loyalty_accounts loyalty_accounts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_accounts
    ADD CONSTRAINT loyalty_accounts_pkey PRIMARY KEY (id);


--
-- Name: loyalty_transactions loyalty_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.loyalty_transactions
    ADD CONSTRAINT loyalty_transactions_pkey PRIMARY KEY (id);


--
-- Name: menu_categories menu_categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_categories
    ADD CONSTRAINT menu_categories_pkey PRIMARY KEY (id);


--
-- Name: menu_item_allergens menu_item_allergens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_allergens
    ADD CONSTRAINT menu_item_allergens_pkey PRIMARY KEY (id);


--
-- Name: menu_item_modifier_groups menu_item_modifier_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_modifier_groups
    ADD CONSTRAINT menu_item_modifier_groups_pkey PRIMARY KEY (id);


--
-- Name: menu_item_variants menu_item_variants_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_item_variants
    ADD CONSTRAINT menu_item_variants_pkey PRIMARY KEY (id);


--
-- Name: menu_items menu_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.menu_items
    ADD CONSTRAINT menu_items_pkey PRIMARY KEY (id);


--
-- Name: modifier_groups modifier_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modifier_groups
    ADD CONSTRAINT modifier_groups_pkey PRIMARY KEY (id);


--
-- Name: modifiers modifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.modifiers
    ADD CONSTRAINT modifiers_pkey PRIMARY KEY (id);


--
-- Name: notifications notifications_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.notifications
    ADD CONSTRAINT notifications_pkey PRIMARY KEY (id);


--
-- Name: order_item_modifiers order_item_modifiers_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_item_modifiers
    ADD CONSTRAINT order_item_modifiers_pkey PRIMARY KEY (id);


--
-- Name: order_items order_items_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_items
    ADD CONSTRAINT order_items_pkey PRIMARY KEY (id);


--
-- Name: order_promotions order_promotions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.order_promotions
    ADD CONSTRAINT order_promotions_pkey PRIMARY KEY (id);


--
-- Name: orders orders_order_number_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_order_number_key UNIQUE (order_number);


--
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (id);


--
-- Name: payments payments_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.payments
    ADD CONSTRAINT payments_pkey PRIMARY KEY (id);


--
-- Name: promo_codes promo_codes_code_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_code_key UNIQUE (code);


--
-- Name: promo_codes promo_codes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.promo_codes
    ADD CONSTRAINT promo_codes_pkey PRIMARY KEY (id);


--
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

\unrestrict lwgGQZ4PFEziDh5zTfrRYXaeWD4ZEyHDnIMPXT3GIqJci3Uia2Fi2N6QmWlmb6L

