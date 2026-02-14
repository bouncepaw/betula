// SPDX-FileCopyrightText: 2024 Timur Ismagilov <https://bouncepaw.com>
//
// SPDX-License-Identifier: AGPL-3.0-only

async function copyTextElem(text, elem) {
    await navigator.clipboard.writeText(text)
    elem.textContent = "Copied!"
}
