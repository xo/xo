-- SQL Schema template for the django schema.
-- Generated on Sun Jul 25 07:11:15 WIB 2021 by xo.

CREATE TABLE auth_group (
    id NUMBER(11) GENERATED ALWAYS AS IDENTITY,
    name NVARCHAR2,
    CONSTRAINT sys_c0012433 UNIQUE (id),
    CONSTRAINT sys_c0012434 UNIQUE (name)
);


CREATE TABLE auth_group_permissions (
    id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    group_id NUMBER(11) NOT NULL CONSTRAINT auth_grou_group_id_b120cbf9_f REFERENCES auth_group (group_id),
    permission_id NUMBER(11) NOT NULL CONSTRAINT auth_grou_permissio_84c5c92e_f REFERENCES auth_permission (permission_id),
    CONSTRAINT auth_grou_group_id__0cd325b0_u UNIQUE (group_id, permission_id),
    CONSTRAINT sys_c0012438 UNIQUE (id)
);

CREATE INDEX auth_group_group_id_b120cbf9 ON auth_group_permissions (group_id);
CREATE INDEX auth_group_permission_84c5c92e ON auth_group_permissions (permission_id);

CREATE TABLE auth_permission (
    id NUMBER(11) GENERATED ALWAYS AS IDENTITY,
    name NVARCHAR2,
    content_type_id NUMBER(11) NOT NULL CONSTRAINT auth_perm_content_t_2f476e4b_f REFERENCES django_content_type (content_type_id),
    codename NVARCHAR2,
    CONSTRAINT auth_perm_content_t_01ab375a_u UNIQUE (content_type_id, codename),
    CONSTRAINT sys_c0012431 UNIQUE (id)
);

CREATE INDEX auth_permi_content_ty_2f476e4b ON auth_permission (content_type_id);

CREATE TABLE auth_user (
    id NUMBER(11) GENERATED ALWAYS AS IDENTITY,
    password NVARCHAR2,
    last_login TIMESTAMP(6),
    is_superuser NUMBER(1) NOT NULL,
    username NVARCHAR2,
    first_name NVARCHAR2,
    last_name NVARCHAR2,
    email NVARCHAR2,
    is_staff NUMBER(1) NOT NULL,
    is_active NUMBER(1) NOT NULL,
    date_joined TIMESTAMP(6) NOT NULL,
    CONSTRAINT sys_c0012448 UNIQUE (id),
    CONSTRAINT sys_c0012449 UNIQUE (username)
);


CREATE TABLE auth_user_groups (
    id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    user_id NUMBER(11) NOT NULL CONSTRAINT auth_user_user_id_6a12ed8b_f REFERENCES auth_user (user_id),
    group_id NUMBER(11) NOT NULL CONSTRAINT auth_user_group_id_97559544_f REFERENCES auth_group (group_id),
    CONSTRAINT auth_user_user_id_g_94350c0c_u UNIQUE (user_id, group_id),
    CONSTRAINT sys_c0012453 UNIQUE (id)
);

CREATE INDEX auth_user__group_id_97559544 ON auth_user_groups (group_id);
CREATE INDEX auth_user__user_id_6a12ed8b ON auth_user_groups (user_id);

CREATE TABLE auth_user_user_permissions (
    id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    user_id NUMBER(11) NOT NULL CONSTRAINT auth_user_user_id_a95ead1b_f REFERENCES auth_user (user_id),
    permission_id NUMBER(11) NOT NULL CONSTRAINT auth_user_permissio_1fbb5f2c_f REFERENCES auth_permission (permission_id),
    CONSTRAINT auth_user_user_id_p_14a6b632_u UNIQUE (user_id, permission_id),
    CONSTRAINT sys_c0012457 UNIQUE (id)
);

CREATE INDEX auth_user__permission_1fbb5f2c ON auth_user_user_permissions (permission_id);
CREATE INDEX auth_user__user_id_a95ead1b ON auth_user_user_permissions (user_id);

CREATE TABLE authors (
    author_id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    name NCLOB,
    CONSTRAINT sys_c0012478 UNIQUE (author_id)
);


CREATE TABLE books (
    book_id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    isbn NVARCHAR2,
    book_type NUMBER(11) NOT NULL,
    title NVARCHAR2,
    year NUMBER(11) NOT NULL,
    available TIMESTAMP(6) NOT NULL,
    books_author_id_fkey NUMBER(19) NOT NULL CONSTRAINT books_books_aut_73ac0c26_f REFERENCES authors (books_author_id_fkey),
    CONSTRAINT sys_c0012486 UNIQUE (book_id)
);

CREATE INDEX books_books_auth_73ac0c26 ON books (books_author_id_fkey);

CREATE TABLE books_tags (
    id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    book_id NUMBER(19) NOT NULL CONSTRAINT books_tag_book_id_73d7d8e8_f REFERENCES books (book_id),
    tag_id NUMBER(19) NOT NULL CONSTRAINT books_tag_tag_id_8d70b40a_f REFERENCES tags (tag_id),
    CONSTRAINT books_tag_book_id_t_29db9e39_u UNIQUE (book_id, tag_id),
    CONSTRAINT sys_c0012490 UNIQUE (id)
);

CREATE INDEX books_tags_book_id_73d7d8e8 ON books_tags (book_id);
CREATE INDEX books_tags_tag_id_8d70b40a ON books_tags (tag_id);

CREATE TABLE django_admin_log (
    id NUMBER(11) GENERATED ALWAYS AS IDENTITY,
    action_time TIMESTAMP(6) NOT NULL,
    object_id NCLOB,
    object_repr NVARCHAR2,
    action_flag NUMBER(11) NOT NULL,
    change_message NCLOB,
    content_type_id NUMBER(11) CONSTRAINT django_ad_content_t_c4bce8eb_f REFERENCES django_content_type (content_type_id),
    user_id NUMBER(11) NOT NULL CONSTRAINT django_ad_user_id_c564eba6_f REFERENCES auth_user (user_id),
    CONSTRAINT sys_c0012474 UNIQUE (id)
);

CREATE INDEX django_adm_content_ty_c4bce8eb ON django_admin_log (content_type_id);
CREATE INDEX django_adm_user_id_c564eba6 ON django_admin_log (user_id);

CREATE TABLE django_content_type (
    id NUMBER(11) GENERATED ALWAYS AS IDENTITY,
    app_label NVARCHAR2,
    model NVARCHAR2,
    CONSTRAINT django_co_app_label_76bd3d3b_u UNIQUE (app_label, model),
    CONSTRAINT sys_c0012427 UNIQUE (id)
);


CREATE TABLE django_migrations (
    id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    app NVARCHAR2,
    name NVARCHAR2,
    applied TIMESTAMP(6) NOT NULL,
    CONSTRAINT sys_c0012425 UNIQUE (id)
);


CREATE TABLE django_session (
    session_key NVARCHAR2 NOT NULL,
    session_data NCLOB,
    expire_date TIMESTAMP(6) NOT NULL,
    CONSTRAINT sys_c0012497 UNIQUE (session_key)
);

CREATE INDEX django_ses_expire_dat_a5c62663 ON django_session (expire_date);

CREATE TABLE tags (
    tag_id NUMBER(19) GENERATED ALWAYS AS IDENTITY,
    tag NVARCHAR2,
    CONSTRAINT sys_c0012480 UNIQUE (tag_id)
);


