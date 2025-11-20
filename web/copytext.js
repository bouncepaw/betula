// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

async function copyTextElem(text, elem) {
    await navigator.clipboard.writeText(text)
    elem.textContent = "Copied!"
}
