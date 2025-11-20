-- SPDX-FileCopyrightText: 2022-2025 Betula contributors
--
-- SPDX-License-Identifier: AGPL-3.0-only

-- Storing these things:
-- https://docs.joinmastodon.org/spec/activitypub/#publicKey
create table PublicKeys (
    ID text not null primary key,
    Owner text not null,
    PublicKeyPEM text not null
);