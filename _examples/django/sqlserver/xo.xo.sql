-- SQL Schema template for the django schema.
-- Generated on Sun Jul 25 07:11:29 WIB 2021 by xo.

CREATE TABLE auth_group (
    id INT(10) IDENTITY(1, 1),
    name NVARCHAR(150) NOT NULL,
    CONSTRAINT PK__auth_gro__3213E83FAD31572B PRIMARY KEY (id),
    CONSTRAINT auth_group_name_a6ea08ec_uniq UNIQUE (name)
);


CREATE TABLE auth_group_permissions (
    id BIGINT(19) IDENTITY(1, 1),
    group_id INT(10) NOT NULL CONSTRAINT auth_group_permissions_group_id_b120cbf9_fk_auth_group_id REFERENCES auth_group (group_id),
    permission_id INT(10) NOT NULL CONSTRAINT auth_group_permissions_permission_id_84c5c92e_fk_auth_permission_id REFERENCES auth_permission (permission_id),
    CONSTRAINT PK__auth_gro__3213E83F96B7E0C9 PRIMARY KEY (id),
    CONSTRAINT auth_group_permissions_group_id_permission_id_0cd325b0_uniq UNIQUE (group_id, permission_id)
);

CREATE INDEX auth_group_permissions_group_id_b120cbf9 ON auth_group_permissions (group_id);
CREATE INDEX auth_group_permissions_permission_id_84c5c92e ON auth_group_permissions (permission_id);

CREATE TABLE auth_permission (
    id INT(10) IDENTITY(1, 1),
    name NVARCHAR(255) NOT NULL,
    content_type_id INT(10) NOT NULL CONSTRAINT auth_permission_content_type_id_2f476e4b_fk_django_content_type_id REFERENCES django_content_type (content_type_id),
    codename NVARCHAR(100) NOT NULL,
    CONSTRAINT PK__auth_per__3213E83F59211B96 PRIMARY KEY (id),
    CONSTRAINT auth_permission_content_type_id_codename_01ab375a_uniq UNIQUE (content_type_id, codename)
);

CREATE INDEX auth_permission_content_type_id_2f476e4b ON auth_permission (content_type_id);

CREATE TABLE auth_user (
    id INT(10) IDENTITY(1, 1),
    password NVARCHAR(128) NOT NULL,
    last_login DATETIME2(27, 7),
    is_superuser BIT(1) NOT NULL,
    username NVARCHAR(150) NOT NULL,
    first_name NVARCHAR(150) NOT NULL,
    last_name NVARCHAR(150) NOT NULL,
    email NVARCHAR(254) NOT NULL,
    is_staff BIT(1) NOT NULL,
    is_active BIT(1) NOT NULL,
    date_joined DATETIME2(27, 7) NOT NULL,
    CONSTRAINT PK__auth_use__3213E83FB9D2BD57 PRIMARY KEY (id),
    CONSTRAINT auth_user_username_6821ab7c_uniq UNIQUE (username)
);


CREATE TABLE auth_user_groups (
    id BIGINT(19) IDENTITY(1, 1),
    user_id INT(10) NOT NULL CONSTRAINT auth_user_groups_user_id_6a12ed8b_fk_auth_user_id REFERENCES auth_user (user_id),
    group_id INT(10) NOT NULL CONSTRAINT auth_user_groups_group_id_97559544_fk_auth_group_id REFERENCES auth_group (group_id),
    CONSTRAINT PK__auth_use__3213E83F1C81E130 PRIMARY KEY (id),
    CONSTRAINT auth_user_groups_user_id_group_id_94350c0c_uniq UNIQUE (user_id, group_id)
);

CREATE INDEX auth_user_groups_group_id_97559544 ON auth_user_groups (group_id);
CREATE INDEX auth_user_groups_user_id_6a12ed8b ON auth_user_groups (user_id);

CREATE TABLE auth_user_user_permissions (
    id BIGINT(19) IDENTITY(1, 1),
    user_id INT(10) NOT NULL CONSTRAINT auth_user_user_permissions_user_id_a95ead1b_fk_auth_user_id REFERENCES auth_user (user_id),
    permission_id INT(10) NOT NULL CONSTRAINT auth_user_user_permissions_permission_id_1fbb5f2c_fk_auth_permission_id REFERENCES auth_permission (permission_id),
    CONSTRAINT PK__auth_use__3213E83FE743C00D PRIMARY KEY (id),
    CONSTRAINT auth_user_user_permissions_user_id_permission_id_14a6b632_uniq UNIQUE (user_id, permission_id)
);

CREATE INDEX auth_user_user_permissions_permission_id_1fbb5f2c ON auth_user_user_permissions (permission_id);
CREATE INDEX auth_user_user_permissions_user_id_a95ead1b ON auth_user_user_permissions (user_id);

CREATE TABLE authors (
    author_id BIGINT(19) IDENTITY(1, 1),
    name NVARCHAR NOT NULL,
    CONSTRAINT PK__authors__86516BCF90AC8A11 PRIMARY KEY (author_id)
);


CREATE TABLE books (
    book_id BIGINT(19) IDENTITY(1, 1),
    isbn NVARCHAR(255) NOT NULL,
    book_type INT(10) NOT NULL,
    title NVARCHAR(255) NOT NULL,
    year INT(10) NOT NULL,
    available DATETIME2(27, 7) NOT NULL,
    books_author_id_fkey BIGINT(19) NOT NULL CONSTRAINT books_books_author_id_fkey_73ac0c26_fk_authors_author_id REFERENCES authors (books_author_id_fkey),
    CONSTRAINT PK__books__490D1AE1692A90D9 PRIMARY KEY (book_id)
);

CREATE INDEX books_books_author_id_fkey_73ac0c26 ON books (books_author_id_fkey);

CREATE TABLE books_tags (
    id BIGINT(19) IDENTITY(1, 1),
    book_id BIGINT(19) NOT NULL CONSTRAINT books_tags_book_id_73d7d8e8_fk_books_book_id REFERENCES books (book_id),
    tag_id BIGINT(19) NOT NULL CONSTRAINT books_tags_tag_id_8d70b40a_fk_tags_tag_id REFERENCES tags (tag_id),
    CONSTRAINT PK__books_ta__3213E83F7238E299 PRIMARY KEY (id),
    CONSTRAINT books_tags_book_id_tag_id_29db9e39_uniq UNIQUE (book_id, tag_id)
);

CREATE INDEX books_tags_book_id_73d7d8e8 ON books_tags (book_id);
CREATE INDEX books_tags_tag_id_8d70b40a ON books_tags (tag_id);

CREATE TABLE django_admin_log (
    id INT(10) IDENTITY(1, 1),
    action_time DATETIME2(27, 7) NOT NULL,
    object_id NVARCHAR,
    object_repr NVARCHAR(200) NOT NULL,
    action_flag SMALLINT(5) NOT NULL,
    change_message NVARCHAR NOT NULL,
    content_type_id INT(10) CONSTRAINT django_admin_log_content_type_id_c4bce8eb_fk_django_content_type_id REFERENCES django_content_type (content_type_id),
    user_id INT(10) NOT NULL CONSTRAINT django_admin_log_user_id_c564eba6_fk_auth_user_id REFERENCES auth_user (user_id),
    CONSTRAINT PK__django_a__3213E83F94FFE9D7 PRIMARY KEY (id)
);

CREATE INDEX django_admin_log_content_type_id_c4bce8eb ON django_admin_log (content_type_id);
CREATE INDEX django_admin_log_user_id_c564eba6 ON django_admin_log (user_id);

CREATE TABLE django_content_type (
    id INT(10) IDENTITY(1, 1),
    app_label NVARCHAR(100) NOT NULL,
    model NVARCHAR(100) NOT NULL,
    CONSTRAINT PK__django_c__3213E83FC7138A59 PRIMARY KEY (id),
    CONSTRAINT django_content_type_app_label_model_76bd3d3b_uniq UNIQUE (app_label, model)
);


CREATE TABLE django_migrations (
    id BIGINT(19) IDENTITY(1, 1),
    app NVARCHAR(255) NOT NULL,
    name NVARCHAR(255) NOT NULL,
    applied DATETIME2(27, 7) NOT NULL,
    CONSTRAINT PK__django_m__3213E83F7AEDAF13 PRIMARY KEY (id)
);


CREATE TABLE django_session (
    session_key NVARCHAR(40) NOT NULL,
    session_data NVARCHAR NOT NULL,
    expire_date DATETIME2(27, 7) NOT NULL,
    CONSTRAINT PK__django_s__B3BA0F1F4E30DAE4 PRIMARY KEY (session_key)
);

CREATE INDEX django_session_expire_date_a5c62663 ON django_session (expire_date);

CREATE TABLE tags (
    tag_id BIGINT(19) IDENTITY(1, 1),
    tag NVARCHAR(50) NOT NULL,
    CONSTRAINT PK__tags__4296A2B65BF225F1 PRIMARY KEY (tag_id)
);


