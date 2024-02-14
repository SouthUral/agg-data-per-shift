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
-- Name: db_test; Type: DATABASE; Schema: -; Owner: kovalenko
--

-- CREATE DATABASE db_test WITH TEMPLATE = template0 ENCODING = 'UTF8' LOCALE_PROVIDER = libc LOCALE = 'en_US.utf8';


-- ALTER DATABASE db_test OWNER TO kovalenko;

\connect db_test

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
-- Name: test_system; Type: SCHEMA; Schema: -; Owner: kovalenko
--

CREATE SCHEMA test_system;


ALTER SCHEMA test_system OWNER TO kovalenko;

--
-- Name: get_last_offset(); Type: FUNCTION; Schema: test_system; Owner: kovalenko
--

CREATE FUNCTION test_system.get_last_offset() RETURNS bigint
    LANGUAGE sql
    AS $$
	select coalesce(
		(select id_offset
		from test_system.offset_storage
		order by id_offset desc
		limit 1), 0
	);
$$;


ALTER FUNCTION test_system.get_last_offset() OWNER TO kovalenko;

--
-- Name: get_messages(bigint, integer); Type: FUNCTION; Schema: test_system; Owner: kovalenko
--

CREATE FUNCTION test_system.get_messages(p_offset bigint, p_limit integer) RETURNS TABLE(id_offset bigint, message jsonb)
    LANGUAGE plpgsql
    AS $$
begin
	return query
	select
		id_offset,
		message
	from test_system.messages
	where id_offset > p_offset
	order by id_offset
	limit p_limit;
end;
$$;


ALTER FUNCTION test_system.get_messages(p_offset bigint, p_limit integer) OWNER TO kovalenko;

--
-- Name: get_offset(); Type: FUNCTION; Schema: test_system; Owner: kovalenko
--

CREATE FUNCTION test_system.get_offset() RETURNS bigint
    LANGUAGE sql
    AS $$
	select coalesce(
		(select id_offset
		from test_system.messages
		order by id desc
		limit 1), 0
	);
$$;


ALTER FUNCTION test_system.get_offset() OWNER TO kovalenko;

--
-- Name: rec_message(jsonb, bigint); Type: PROCEDURE; Schema: test_system; Owner: kovalenko
--

CREATE PROCEDURE test_system.rec_message(IN p_message jsonb, IN p_offset bigint)
    LANGUAGE plpgsql
    AS $$
	begin 
		insert into test_system.messages(
			id_offset,
			message
		) 
		values (
			p_offset,
			p_message
		);
	end;
$$;


ALTER PROCEDURE test_system.rec_message(IN p_message jsonb, IN p_offset bigint) OWNER TO kovalenko;

--
-- Name: update_last_offset(bigint); Type: PROCEDURE; Schema: test_system; Owner: kovalenko
--

CREATE PROCEDURE test_system.update_last_offset(IN p_offset bigint)
    LANGUAGE plpgsql
    AS $$
	declare 
		_count int;
	begin 
		select count(*) into _count from test_system.offset_storage;
		if _count = 0 then
			insert into test_system.offset_storage(
				id_offset
			) 
			values (
				p_offset
			);
		else 
			update test_system.offset_storage
			set 
				id_offset = p_offset,
				mess_date = CURRENT_TIMESTAMP
			where id = 1;
		end if;
	end;
$$;


ALTER PROCEDURE test_system.update_last_offset(IN p_offset bigint) OWNER TO kovalenko;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: messages; Type: TABLE; Schema: test_system; Owner: kovalenko
--

CREATE TABLE test_system.messages (
    id integer NOT NULL,
    id_offset bigint,
    mess_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP,
    message jsonb
);


ALTER TABLE test_system.messages OWNER TO kovalenko;

--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: test_system; Owner: kovalenko
--

CREATE SEQUENCE test_system.messages_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE test_system.messages_id_seq OWNER TO kovalenko;

--
-- Name: messages_id_seq; Type: SEQUENCE OWNED BY; Schema: test_system; Owner: kovalenko
--

ALTER SEQUENCE test_system.messages_id_seq OWNED BY test_system.messages.id;


--
-- Name: offset_storage; Type: TABLE; Schema: test_system; Owner: kovalenko
--

CREATE TABLE test_system.offset_storage (
    id integer NOT NULL,
    id_offset bigint,
    mess_date timestamp without time zone DEFAULT CURRENT_TIMESTAMP
);


ALTER TABLE test_system.offset_storage OWNER TO kovalenko;

--
-- Name: offset_storage_id_seq; Type: SEQUENCE; Schema: test_system; Owner: kovalenko
--

CREATE SEQUENCE test_system.offset_storage_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE test_system.offset_storage_id_seq OWNER TO kovalenko;

--
-- Name: offset_storage_id_seq; Type: SEQUENCE OWNED BY; Schema: test_system; Owner: kovalenko
--

ALTER SEQUENCE test_system.offset_storage_id_seq OWNED BY test_system.offset_storage.id;


--
-- Name: messages id; Type: DEFAULT; Schema: test_system; Owner: kovalenko
--

ALTER TABLE ONLY test_system.messages ALTER COLUMN id SET DEFAULT nextval('test_system.messages_id_seq'::regclass);


--
-- Name: offset_storage id; Type: DEFAULT; Schema: test_system; Owner: kovalenko
--

ALTER TABLE ONLY test_system.offset_storage ALTER COLUMN id SET DEFAULT nextval('test_system.offset_storage_id_seq'::regclass);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: test_system; Owner: kovalenko
--

ALTER TABLE ONLY test_system.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- PostgreSQL database dump complete
--

