// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

self.addEventListener("install", function(_e) {
    self.skipWaiting();
});

self.addEventListener("activate", function (event) {
    event.waitUntil(clients.claim());
});