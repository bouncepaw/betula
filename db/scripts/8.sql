-- DROPPED IN 11: it was causing confusion to me
create table WebFingerAccts (
    Acct text primary key, -- In acct:bouncepaw@links.bouncepaw.com, everything after acct:
    ActorURL text, -- This is the id that will be used for the actor
    Document blob, -- The whole document
    LastCheckedAt text not null default current_timestamp
)
-- END DROPPED