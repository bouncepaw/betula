// SPDX-FileCopyrightText: 2025 Guilherme Puida Moreira
//
// SPDX-License-Identifier: AGPL-3.0-only

self.addEventListener("install", function(_e) {
    self.skipWaiting();
});

self.addEventListener("activate", function (event) {
    event.waitUntil(clients.claim());
});