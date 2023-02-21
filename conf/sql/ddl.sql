--file_obj_t
CREATE TABLE if not exists file_obj_t (
    md5_hex varchar(32) not null,
    file_name varchar(2000) not null,
    file_time varchar(19) not null,
    time_zone varchar(6) not null default '+8:00',
    time_origin varchar(2000),
    label varchar(2000),
    create_time BIGINT not null,
    update_time BIGINT not null,
    create_local_time BIGINT not null,
    update_local_time BIGINT not null,
    valid_flag int not null,
    CONSTRAINT pk_file_obj_t PRIMARY KEY (md5_hex),
    CONSTRAINT un_file_name UNIQUE (file_time)
);
CREATE INDEX if not exists idx_file_obj_t_1 ON file_obj_t (file_name, valid_flag);
CREATE INDEX if not exists idx_file_obj_t_2 ON file_obj_t (valid_flag,file_time);
CREATE INDEX if not exists idx_file_obj_t_3 ON file_obj_t (valid_flag,label);