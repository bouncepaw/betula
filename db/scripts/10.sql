-- NOT USED. Deprecated in favor of in-memory storage of these things. Though we might be using it again in the future...
-- Storing these things:
-- https://docs.joinmastodon.org/spec/activitypub/#publicKey
create table PublicKeys (
    ID text not null primary key,
    Owner text not null,
    PublicKeyPEM text not null
);