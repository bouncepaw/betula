create table Notifications
(
    ID        integer primary key autoincrement,
    CreatedAt text not null default current_timestamp,
    Kind      text not null,
    Payload   blob not null -- JSON, schema varies by Kind
);