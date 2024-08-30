CREATE TABLE public.table_config (
    id serial4 NOT NULL,
    table_name varchar(255) NOT NULL,
    column_name varchar(255) NOT NULL,
    column_order int4 NOT NULL,
    CONSTRAINT table_config_pkey PRIMARY KEY (id),
    CONSTRAINT unique_table_order UNIQUE (table_name, column_order)
);

CREATE TABLE public.tasks (
    id serial4 NOT NULL,
    file_name varchar(255) NULL,
    table_name varchar(255) NULL,
    status varchar(50) NULL,
    error_message text NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
    updated_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT tasks_pkey PRIMARY KEY (id)
);

CREATE TABLE public.uploaded_files (
    id serial4 NOT NULL,
    filename varchar(255) NULL,
    md5 varchar(32) NULL,
    status varchar(50) DEFAULT 'In Progress'::character varying NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT uploaded_files_md5_key UNIQUE (md5),
    CONSTRAINT uploaded_files_pkey PRIMARY KEY (id)
);

CREATE TABLE public.users (
    id serial4 NOT NULL,
    username text NOT NULL,
    password_hash text NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id),
    CONSTRAINT users_username_key UNIQUE (username)
);