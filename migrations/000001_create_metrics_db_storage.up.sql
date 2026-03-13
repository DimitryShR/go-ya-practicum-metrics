CREATE SCHEMA IF NOT EXISTS metrics;
GRANT ALL PRIVILEGES ON SCHEMA metrics TO as_admin;
GRANT USAGE ON SCHEMA metrics TO user_main;

CREATE TABLE IF NOT EXISTS metrics.gauges (
	"name" varchar NOT NULL,
	value float8 NOT NULL,
	updated_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT gauges_pkey PRIMARY KEY ("name")
);
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE metrics.gauges TO user_main;


CREATE TABLE IF NOT EXISTS metrics.counters (
	"name" varchar NOT NULL,
	value int8 NOT NULL,
	updated_at timestamptz DEFAULT now() NOT NULL,
	CONSTRAINT counters_pkey PRIMARY KEY ("name")
);
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE metrics.counters TO user_main;
