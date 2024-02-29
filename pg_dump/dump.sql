--
-- PostgreSQL database dump
--

-- Dumped from database version 15.5 (Debian 15.5-1.pgdg120+1)
-- Dumped by pg_dump version 16.1

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
-- Name: report_bd; Type: DATABASE; Schema: -; Owner: kovalenko
--

-- CREATE DATABASE report_bd WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';


-- ALTER DATABASE report_bd OWNER TO kovalenko;

\connect report_bd

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
-- Name: reports; Type: SCHEMA; Schema: -; Owner: kovalenko
--

CREATE SCHEMA reports;


ALTER SCHEMA reports OWNER TO kovalenko;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: drivers_sessions; Type: TABLE; Schema: reports; Owner: kovalenko
--

CREATE TABLE reports.drivers_sessions (
    id integer NOT NULL,
    shift_id integer,
    driver_id integer,
    event_offset bigint,
    time_start_session timestamp without time zone,
    time_update_session timestamp without time zone,
    av_speed double precision,
    eng_hours_start double precision,
    eng_hours_current double precision,
    eng_hours_end double precision,
    mileage_start integer,
    mileage_current integer,
    mileage_end integer,
    mileage_loaded integer,
    mileage_at_beginning_of_loading integer,
    mileage_empty integer,
    mileage_gps_start integer,
    mileage_gps_current integer,
    mileage_gps_end integer,
    mileage_gps_loaded integer,
    mileage_gps_at_beginning_of_loading integer,
    mileage_gps_empty integer
);


ALTER TABLE reports.drivers_sessions OWNER TO kovalenko;

--
-- Name: drivers_sessions_id_seq; Type: SEQUENCE; Schema: reports; Owner: kovalenko
--

CREATE SEQUENCE reports.drivers_sessions_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE reports.drivers_sessions_id_seq OWNER TO kovalenko;

--
-- Name: drivers_sessions_id_seq; Type: SEQUENCE OWNED BY; Schema: reports; Owner: kovalenko
--

ALTER SEQUENCE reports.drivers_sessions_id_seq OWNED BY reports.drivers_sessions.id;


--
-- Name: shifts; Type: TABLE; Schema: reports; Owner: kovalenko
--

CREATE TABLE reports.shifts (
    id integer NOT NULL,
    num_shift integer,
    shift_date_start timestamp without time zone,
    shift_date_end timestamp without time zone,
    shift_date timestamp without time zone,
    updated_time timestamp without time zone,
    event_offset bigint,
    current_driver_id integer,
    loaded boolean,
    eng_hours_start double precision,
    eng_hours_current double precision,
    eng_hours_end double precision,
    mileage_start integer,
    mileage_current integer,
    mileage_end integer,
    mileage_loaded integer,
    mileage_at_beginning_of_loading integer,
    mileage_empty integer,
    mileage_gps_start integer,
    mileage_gps_current integer,
    mileage_gps_end integer,
    mileage_gps_loaded integer,
    mileage_gps_at_beginning_of_loading integer,
    mileage_gps_empty integer
);


ALTER TABLE reports.shifts OWNER TO kovalenko;

--
-- Name: shifts_id_seq; Type: SEQUENCE; Schema: reports; Owner: kovalenko
--

CREATE SEQUENCE reports.shifts_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE reports.shifts_id_seq OWNER TO kovalenko;

--
-- Name: shifts_id_seq; Type: SEQUENCE OWNED BY; Schema: reports; Owner: kovalenko
--

ALTER SEQUENCE reports.shifts_id_seq OWNED BY reports.shifts.id;


--
-- Name: drivers_sessions id; Type: DEFAULT; Schema: reports; Owner: kovalenko
--

ALTER TABLE ONLY reports.drivers_sessions ALTER COLUMN id SET DEFAULT nextval('reports.drivers_sessions_id_seq'::regclass);


--
-- Name: shifts id; Type: DEFAULT; Schema: reports; Owner: kovalenko
--

ALTER TABLE ONLY reports.shifts ALTER COLUMN id SET DEFAULT nextval('reports.shifts_id_seq'::regclass);


--
-- Name: drivers_sessions drivers_sessions_pkey; Type: CONSTRAINT; Schema: reports; Owner: kovalenko
--

ALTER TABLE ONLY reports.drivers_sessions
    ADD CONSTRAINT drivers_sessions_pkey PRIMARY KEY (id);


--
-- Name: shifts shifts_pkey; Type: CONSTRAINT; Schema: reports; Owner: kovalenko
--

ALTER TABLE ONLY reports.shifts
    ADD CONSTRAINT shifts_pkey PRIMARY KEY (id);


--
-- Name: drivers_sessions drivers_sessions_shift_id_fkey; Type: FK CONSTRAINT; Schema: reports; Owner: kovalenko
--

ALTER TABLE ONLY reports.drivers_sessions
    ADD CONSTRAINT drivers_sessions_shift_id_fkey FOREIGN KEY (shift_id) REFERENCES reports.shifts(id);


--
-- PostgreSQL database dump complete
--

