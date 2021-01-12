-- public.ar_internal_metadata definition

-- Drop table

-- DROP TABLE public.ar_internal_metadata;

CREATE TABLE public.ar_internal_metadata (
	"key" varchar NOT NULL,
	value varchar NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT ar_internal_metadata_pkey PRIMARY KEY (key)
);


-- public.bulk_delete_jobs definition

-- Drop table

-- DROP TABLE public.bulk_delete_jobs;

CREATE TABLE public.bulk_delete_jobs (
	id bigserial NOT NULL,
	requested_by varchar NULL,
	institutional_approver varchar NULL,
	aptrust_approver varchar NULL,
	institutional_approval_at timestamp NULL,
	aptrust_approval_at timestamp NULL,
	note text NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	institution_id int4 NOT NULL,
	CONSTRAINT bulk_delete_jobs_pkey PRIMARY KEY (id)
);


-- public.bulk_delete_jobs_emails definition

-- Drop table

-- DROP TABLE public.bulk_delete_jobs_emails;

CREATE TABLE public.bulk_delete_jobs_emails (
	bulk_delete_job_id int8 NULL,
	email_id int8 NULL
);
CREATE INDEX index_bulk_delete_jobs_emails_on_bulk_delete_job_id ON public.bulk_delete_jobs_emails USING btree (bulk_delete_job_id);
CREATE INDEX index_bulk_delete_jobs_emails_on_email_id ON public.bulk_delete_jobs_emails USING btree (email_id);


-- public.bulk_delete_jobs_generic_files definition

-- Drop table

-- DROP TABLE public.bulk_delete_jobs_generic_files;

CREATE TABLE public.bulk_delete_jobs_generic_files (
	bulk_delete_job_id int8 NULL,
	generic_file_id int8 NULL
);
CREATE INDEX index_bulk_delete_jobs_generic_files_on_bulk_delete_job_id ON public.bulk_delete_jobs_generic_files USING btree (bulk_delete_job_id);
CREATE INDEX index_bulk_delete_jobs_generic_files_on_generic_file_id ON public.bulk_delete_jobs_generic_files USING btree (generic_file_id);


-- public.bulk_delete_jobs_institutions definition

-- Drop table

-- DROP TABLE public.bulk_delete_jobs_institutions;

CREATE TABLE public.bulk_delete_jobs_institutions (
	bulk_delete_job_id int8 NULL,
	institution_id int8 NULL
);
CREATE INDEX index_bulk_delete_jobs_institutions_on_bulk_delete_job_id ON public.bulk_delete_jobs_institutions USING btree (bulk_delete_job_id);
CREATE INDEX index_bulk_delete_jobs_institutions_on_institution_id ON public.bulk_delete_jobs_institutions USING btree (institution_id);


-- public.bulk_delete_jobs_intellectual_objects definition

-- Drop table

-- DROP TABLE public.bulk_delete_jobs_intellectual_objects;

CREATE TABLE public.bulk_delete_jobs_intellectual_objects (
	bulk_delete_job_id int8 NULL,
	intellectual_object_id int8 NULL
);
CREATE INDEX index_bulk_delete_jobs_intellectual_objects_on_bulk_job_id ON public.bulk_delete_jobs_intellectual_objects USING btree (bulk_delete_job_id);
CREATE INDEX index_bulk_delete_jobs_intellectual_objects_on_object_id ON public.bulk_delete_jobs_intellectual_objects USING btree (intellectual_object_id);


-- public.confirmation_tokens definition

-- Drop table

-- DROP TABLE public.confirmation_tokens;

CREATE TABLE public.confirmation_tokens (
	id bigserial NOT NULL,
	"token" varchar NULL,
	intellectual_object_id int4 NULL,
	generic_file_id int4 NULL,
	institution_id int4 NULL,
	user_id int4 NULL,
	CONSTRAINT confirmation_tokens_pkey PRIMARY KEY (id)
);


-- public.emails definition

-- Drop table

-- DROP TABLE public.emails;

CREATE TABLE public.emails (
	id bigserial NOT NULL,
	email_type varchar NULL,
	event_identifier varchar NULL,
	item_id int4 NULL,
	email_text text NULL,
	user_list text NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	intellectual_object_id int4 NULL,
	generic_file_id int4 NULL,
	institution_id int4 NULL,
	CONSTRAINT emails_pkey PRIMARY KEY (id)
);


-- public.emails_generic_files definition

-- Drop table

-- DROP TABLE public.emails_generic_files;

CREATE TABLE public.emails_generic_files (
	generic_file_id int8 NULL,
	email_id int8 NULL
);
CREATE INDEX index_emails_generic_files_on_email_id ON public.emails_generic_files USING btree (email_id);
CREATE INDEX index_emails_generic_files_on_generic_file_id ON public.emails_generic_files USING btree (generic_file_id);


-- public.emails_intellectual_objects definition

-- Drop table

-- DROP TABLE public.emails_intellectual_objects;

CREATE TABLE public.emails_intellectual_objects (
	intellectual_object_id int8 NULL,
	email_id int8 NULL
);
CREATE INDEX index_emails_intellectual_objects_on_email_id ON public.emails_intellectual_objects USING btree (email_id);
CREATE INDEX index_emails_intellectual_objects_on_intellectual_object_id ON public.emails_intellectual_objects USING btree (intellectual_object_id);


-- public.emails_premis_events definition

-- Drop table

-- DROP TABLE public.emails_premis_events;

CREATE TABLE public.emails_premis_events (
	premis_event_id int8 NULL,
	email_id int8 NULL
);
CREATE INDEX index_emails_premis_events_on_email_id ON public.emails_premis_events USING btree (email_id);
CREATE INDEX index_emails_premis_events_on_premis_event_id ON public.emails_premis_events USING btree (premis_event_id);


-- public.emails_work_items definition

-- Drop table

-- DROP TABLE public.emails_work_items;

CREATE TABLE public.emails_work_items (
	work_item_id int8 NULL,
	email_id int8 NULL
);
CREATE INDEX index_emails_work_items_on_email_id ON public.emails_work_items USING btree (email_id);
CREATE INDEX index_emails_work_items_on_work_item_id ON public.emails_work_items USING btree (work_item_id);


-- public.generic_files definition

-- Drop table

-- DROP TABLE public.generic_files;

CREATE TABLE public.generic_files (
	id serial NOT NULL,
	file_format varchar NULL,
	"size" int8 NULL,
	identifier varchar NULL,
	intellectual_object_id int4 NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	state varchar NULL,
	ingest_state text NULL,
	last_fixity_check timestamp NOT NULL DEFAULT '2000-01-01 00:00:00'::timestamp without time zone,
	institution_id int4 NOT NULL,
	storage_option varchar NOT NULL DEFAULT 'Standard'::character varying,
	uuid varchar NOT NULL,
	CONSTRAINT generic_files_pkey PRIMARY KEY (id)
);
CREATE INDEX index_files_on_inst_state_and_format ON public.generic_files USING btree (institution_id, state, file_format);
CREATE INDEX index_files_on_inst_state_and_updated ON public.generic_files USING btree (institution_id, state, updated_at);
CREATE INDEX index_generic_files_on_created_at ON public.generic_files USING btree (created_at);
CREATE INDEX index_generic_files_on_file_format ON public.generic_files USING btree (file_format);
CREATE INDEX index_generic_files_on_file_format_and_state ON public.generic_files USING btree (file_format, state);
CREATE UNIQUE INDEX index_generic_files_on_identifier ON public.generic_files USING btree (identifier);
CREATE INDEX index_generic_files_on_institution_id ON public.generic_files USING btree (institution_id);
CREATE INDEX index_generic_files_on_institution_id_and_size_and_state ON public.generic_files USING btree (institution_id, size, state);
CREATE INDEX index_generic_files_on_institution_id_and_state ON public.generic_files USING btree (institution_id, state);
CREATE INDEX index_generic_files_on_institution_id_and_updated_at ON public.generic_files USING btree (institution_id, updated_at);
CREATE INDEX index_generic_files_on_intellectual_object_id ON public.generic_files USING btree (intellectual_object_id);
CREATE INDEX index_generic_files_on_intellectual_object_id_and_file_format ON public.generic_files USING btree (intellectual_object_id, file_format);
CREATE INDEX index_generic_files_on_intellectual_object_id_and_state ON public.generic_files USING btree (intellectual_object_id, state);
CREATE INDEX index_generic_files_on_size ON public.generic_files USING btree (size);
CREATE INDEX index_generic_files_on_size_and_state ON public.generic_files USING btree (size, state);
CREATE INDEX index_generic_files_on_state ON public.generic_files USING btree (state);
CREATE INDEX index_generic_files_on_state_and_updated_at ON public.generic_files USING btree (state, updated_at);
CREATE INDEX index_generic_files_on_updated_at ON public.generic_files USING btree (updated_at);
CREATE UNIQUE INDEX index_generic_files_on_uuid ON public.generic_files USING btree (uuid);
CREATE INDEX ix_gf_last_fixity_check ON public.generic_files USING btree (last_fixity_check);


-- public.institutions definition

-- Drop table

-- DROP TABLE public.institutions;

CREATE TABLE public.institutions (
	id serial NOT NULL,
	"name" varchar NULL,
	identifier varchar NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	state varchar NULL,
	"type" varchar NULL,
	member_institution_id int4 NULL,
	deactivated_at timestamp NULL,
	otp_enabled bool NULL,
	receiving_bucket varchar NOT NULL,
	restore_bucket varchar NOT NULL,
	CONSTRAINT institutions_pkey PRIMARY KEY (id)
);
CREATE INDEX index_institutions_on_name ON public.institutions USING btree (name);


-- public.intellectual_objects definition

-- Drop table

-- DROP TABLE public.intellectual_objects;

CREATE TABLE public.intellectual_objects (
	id serial NOT NULL,
	title varchar NULL,
	description text NULL,
	identifier varchar NULL,
	alt_identifier varchar NULL,
	"access" varchar NULL,
	bag_name varchar NULL,
	institution_id int4 NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	state varchar NULL,
	etag varchar NULL,
	ingest_state text NULL,
	bag_group_identifier varchar NOT NULL DEFAULT ''::character varying,
	storage_option varchar NOT NULL DEFAULT 'Standard'::character varying,
	bagit_profile_identifier varchar NULL,
	source_organization varchar NULL,
	internal_sender_identifier varchar NULL,
	internal_sender_description text NULL,
	CONSTRAINT intellectual_objects_pkey PRIMARY KEY (id)
);
CREATE INDEX index_intellectual_objects_on_access ON public.intellectual_objects USING btree (access);
CREATE INDEX index_intellectual_objects_on_bag_name ON public.intellectual_objects USING btree (bag_name);
CREATE INDEX index_intellectual_objects_on_created_at ON public.intellectual_objects USING btree (created_at);
CREATE UNIQUE INDEX index_intellectual_objects_on_identifier ON public.intellectual_objects USING btree (identifier);
CREATE INDEX index_intellectual_objects_on_institution_id ON public.intellectual_objects USING btree (institution_id);
CREATE INDEX index_intellectual_objects_on_institution_id_and_state ON public.intellectual_objects USING btree (institution_id, state);
CREATE INDEX index_intellectual_objects_on_state ON public.intellectual_objects USING btree (state);
CREATE INDEX index_intellectual_objects_on_updated_at ON public.intellectual_objects USING btree (updated_at);


-- public.old_passwords definition

-- Drop table

-- DROP TABLE public.old_passwords;

CREATE TABLE public.old_passwords (
	id bigserial NOT NULL,
	encrypted_password varchar NOT NULL,
	password_salt varchar NULL,
	password_archivable_type varchar NOT NULL,
	password_archivable_id int4 NOT NULL,
	created_at timestamp NULL,
	CONSTRAINT old_passwords_pkey PRIMARY KEY (id)
);
CREATE INDEX index_password_archivable ON public.old_passwords USING btree (password_archivable_type, password_archivable_id);


-- public.premis_events definition

-- Drop table

-- DROP TABLE public.premis_events;

CREATE TABLE public.premis_events (
	id serial NOT NULL,
	identifier varchar NULL,
	event_type varchar NULL,
	date_time varchar NULL,
	outcome_detail varchar NULL,
	detail varchar NULL,
	outcome_information varchar NULL,
	"object" varchar NULL,
	agent varchar NULL,
	intellectual_object_id int4 NULL,
	generic_file_id int4 NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	outcome varchar NULL,
	institution_id int4 NULL,
	intellectual_object_identifier varchar NOT NULL DEFAULT ''::character varying,
	generic_file_identifier varchar NOT NULL DEFAULT ''::character varying,
	old_uuid varchar NULL,
	CONSTRAINT premis_events_pkey PRIMARY KEY (id)
);
CREATE INDEX index_premis_events_date_time_desc ON public.premis_events USING btree (date_time DESC);
CREATE INDEX index_premis_events_on_event_type ON public.premis_events USING btree (event_type);
CREATE INDEX index_premis_events_on_event_type_and_outcome ON public.premis_events USING btree (event_type, outcome);
CREATE INDEX index_premis_events_on_generic_file_id ON public.premis_events USING btree (generic_file_id);
CREATE INDEX index_premis_events_on_generic_file_id_and_event_type ON public.premis_events USING btree (generic_file_id, event_type);
CREATE INDEX index_premis_events_on_generic_file_identifier ON public.premis_events USING btree (generic_file_identifier);
CREATE UNIQUE INDEX index_premis_events_on_identifier ON public.premis_events USING btree (identifier);
CREATE INDEX index_premis_events_on_identifier_and_institution_id ON public.premis_events USING btree (identifier, institution_id);
CREATE INDEX index_premis_events_on_institution_id ON public.premis_events USING btree (institution_id);
CREATE INDEX index_premis_events_on_intellectual_object_id ON public.premis_events USING btree (intellectual_object_id);
CREATE INDEX index_premis_events_on_intellectual_object_identifier ON public.premis_events USING btree (intellectual_object_identifier);
CREATE INDEX index_premis_events_on_outcome ON public.premis_events USING btree (outcome);


-- public.roles definition

-- Drop table

-- DROP TABLE public.roles;

CREATE TABLE public.roles (
	id serial NOT NULL,
	"name" varchar NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT roles_pkey PRIMARY KEY (id)
);


-- public.roles_users definition

-- Drop table

-- DROP TABLE public.roles_users;

CREATE TABLE public.roles_users (
	role_id int4 NULL,
	user_id int4 NULL
);
CREATE INDEX index_roles_users_on_role_id_and_user_id ON public.roles_users USING btree (role_id, user_id);
CREATE INDEX index_roles_users_on_user_id_and_role_id ON public.roles_users USING btree (user_id, role_id);


-- public.schema_migrations definition

-- Drop table

-- DROP TABLE public.schema_migrations;

CREATE TABLE public.schema_migrations (
	"version" varchar NOT NULL,
	CONSTRAINT schema_migrations_pkey PRIMARY KEY (version)
);


-- public.snapshots definition

-- Drop table

-- DROP TABLE public.snapshots;

CREATE TABLE public.snapshots (
	id bigserial NOT NULL,
	audit_date timestamp NULL,
	institution_id int4 NULL,
	apt_bytes int8 NULL,
	"cost" numeric NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	snapshot_type varchar NULL,
	cs_bytes int8 NULL,
	go_bytes int8 NULL,
	CONSTRAINT snapshots_pkey PRIMARY KEY (id)
);


-- public.usage_samples definition

-- Drop table

-- DROP TABLE public.usage_samples;

CREATE TABLE public.usage_samples (
	id serial NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	institution_id varchar NULL,
	"data" text NULL,
	CONSTRAINT usage_samples_pkey PRIMARY KEY (id)
);


-- public.work_items definition

-- Drop table

-- DROP TABLE public.work_items;

CREATE TABLE public.work_items (
	id serial NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	intellectual_object_id int4 NULL,
	generic_file_id int4 NULL,
	"name" varchar NULL,
	etag varchar NULL,
	bucket varchar NULL,
	"user" varchar NULL,
	note text NULL,
	"action" varchar NULL,
	stage varchar NULL,
	status varchar NULL,
	outcome text NULL,
	bag_date timestamp NULL,
	"date" timestamp NULL,
	retry bool NOT NULL DEFAULT false,
	object_identifier varchar NULL,
	generic_file_identifier varchar NULL,
	node varchar(255) NULL,
	pid int4 NULL DEFAULT 0,
	needs_admin_review bool NOT NULL DEFAULT false,
	institution_id int4 NULL,
	queued_at timestamp NULL,
	"size" int8 NULL,
	stage_started_at timestamp NULL,
	aptrust_approver varchar NULL,
	inst_approver varchar NULL,
	CONSTRAINT work_items_pkey PRIMARY KEY (id)
);
CREATE INDEX index_work_items_on_action ON public.work_items USING btree (action);
CREATE INDEX index_work_items_on_date ON public.work_items USING btree (date);
CREATE INDEX index_work_items_on_etag_and_name ON public.work_items USING btree (etag, name);
CREATE INDEX index_work_items_on_generic_file_id ON public.work_items USING btree (generic_file_id);
CREATE INDEX index_work_items_on_institution_id ON public.work_items USING btree (institution_id);
CREATE INDEX index_work_items_on_institution_id_and_date ON public.work_items USING btree (institution_id, date);
CREATE INDEX index_work_items_on_intellectual_object_id ON public.work_items USING btree (intellectual_object_id);
CREATE INDEX index_work_items_on_stage ON public.work_items USING btree (stage);
CREATE INDEX index_work_items_on_status ON public.work_items USING btree (status);


-- public.checksums definition

-- Drop table

-- DROP TABLE public.checksums;

CREATE TABLE public.checksums (
	id serial NOT NULL,
	algorithm varchar NULL,
	datetime timestamp NULL,
	digest varchar NULL,
	generic_file_id int4 NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT checksums_pkey PRIMARY KEY (id),
	CONSTRAINT fk_rails_89bb0866e7 FOREIGN KEY (generic_file_id) REFERENCES generic_files(id)
);
CREATE INDEX index_checksums_on_generic_file_id ON public.checksums USING btree (generic_file_id);


-- public.storage_records definition

-- Drop table

-- DROP TABLE public.storage_records;

CREATE TABLE public.storage_records (
	id bigserial NOT NULL,
	generic_file_id int4 NULL,
	url varchar NULL,
	CONSTRAINT storage_records_pkey PRIMARY KEY (id),
	CONSTRAINT fk_rails_a126ea6adc FOREIGN KEY (generic_file_id) REFERENCES generic_files(id)
);
CREATE INDEX index_storage_records_on_generic_file_id ON public.storage_records USING btree (generic_file_id);


-- public.users definition

-- Drop table

-- DROP TABLE public.users;

CREATE TABLE public.users (
	id serial NOT NULL,
	"name" varchar NULL,
	email varchar NULL,
	phone_number varchar NULL,
	institution_pid varchar NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL,
	encrypted_password varchar NOT NULL DEFAULT ''::character varying,
	reset_password_token varchar NULL,
	reset_password_sent_at timestamp NULL,
	remember_created_at timestamp NULL,
	sign_in_count int4 NOT NULL DEFAULT 0,
	current_sign_in_at timestamp NULL,
	last_sign_in_at timestamp NULL,
	current_sign_in_ip varchar NULL,
	last_sign_in_ip varchar NULL,
	institution_id int4 NULL,
	encrypted_api_secret_key text NULL,
	password_changed_at timestamp NULL,
	encrypted_otp_secret varchar NULL,
	encrypted_otp_secret_iv varchar NULL,
	encrypted_otp_secret_salt varchar NULL,
	consumed_timestep int4 NULL,
	otp_required_for_login bool NULL,
	deactivated_at timestamp NULL,
	enabled_two_factor bool NULL DEFAULT false,
	confirmed_two_factor bool NULL DEFAULT false,
	otp_backup_codes _varchar NULL,
	authy_id varchar NULL,
	last_sign_in_with_authy timestamp NULL,
	authy_status varchar NULL,
	email_verified bool NULL DEFAULT false,
	initial_password_updated bool NULL DEFAULT false,
	force_password_update bool NULL DEFAULT false,
	account_confirmed bool NULL DEFAULT true,
	grace_period timestamp NULL,
	CONSTRAINT users_pkey PRIMARY KEY (id),
	CONSTRAINT fk_rails_7fcf39ca13 FOREIGN KEY (institution_id) REFERENCES institutions(id)
);
CREATE INDEX index_users_on_authy_id ON public.users USING btree (authy_id);
CREATE UNIQUE INDEX index_users_on_email ON public.users USING btree (email);
CREATE INDEX index_users_on_institution_id ON public.users USING btree (institution_id);
CREATE INDEX index_users_on_password_changed_at ON public.users USING btree (password_changed_at);
CREATE UNIQUE INDEX index_users_on_reset_password_token ON public.users USING btree (reset_password_token);
