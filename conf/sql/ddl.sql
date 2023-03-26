--file_obj_t
CREATE TABLE if not exists file_obj_t (
    md5_hex varchar(32) not null,
    file_name varchar(2000) not null,
    file_extension varchar(20),
    file_time varchar(19) not null,
    file_date varchar(10) not null,
    file_month varchar(7) not null,
    time_zone varchar(6) not null default '+8:00',
    time_origin varchar(2000),
    label varchar(2000),
    task_id varchar(50),
    create_time BIGINT not null,
    update_time BIGINT not null,
    valid_flag int not null,
    CONSTRAINT pk_file_obj_t PRIMARY KEY (md5_hex),
    CONSTRAINT un_file_name UNIQUE (file_name)
);
CREATE INDEX if not exists idx_file_obj_t_1 ON file_obj_t (file_name, valid_flag);
CREATE INDEX if not exists idx_file_obj_t_2 ON file_obj_t (valid_flag,file_month);
CREATE INDEX if not exists idx_file_obj_t_3 ON file_obj_t (valid_flag,label);
