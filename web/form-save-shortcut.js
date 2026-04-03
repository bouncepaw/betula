// SPDX-FileCopyrightText: 2026 Danila Gorelko
//
// SPDX-License-Identifier: AGPL-3.0-only

(function () {
	document.addEventListener("keydown", function (event) {
		if (event.key !== "Enter" || (!event.ctrlKey && !event.metaKey)) {
			return;
		}
		var form = document.querySelector("form[supports-ctrl-enter]");
		if (!form) {
			return;
		}
		event.preventDefault();
		if (typeof form.requestSubmit === "function") {
			form.requestSubmit();
		} else {
			form.submit();
		}
	});
})();
