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
--



--
--



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
