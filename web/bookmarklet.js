// SPDX-FileCopyrightText: 2022-2025 Betula contributors
//
// SPDX-License-Identifier: AGPL-3.0-only

// Save link bookmarklet for Betula
// 2023 Umar Getagazov <umar@handlerug.me>
// Public domain, but attribution appreciated.
// https://handlerug.me/betula-save-bookmarklet

(($) => {
    function getSelectionInMycomarkup() {
        function convert(node, parentNodeName = '') {
            if (node instanceof Text) {
                if (node.textContent.trim() === '') {
                    return '';
                }

                return node.textContent
                    .replace(/\\/g,   '\\\\')
                    .replace(/\*\*/g, '\\**')
                    .replace(/\/\//g, '\\//')
                    .replace(/\+\+/g, '\\++');
            }

            let nodeName = node.nodeName.toLowerCase();

            let result = '';
            for (const child of node.childNodes) {
                result += convert(child, nodeName);
            }

            if (nodeName === 'p') {
                return `\n\n${result.trim()}\n\n`;
            } else if (nodeName === 'br') {
                return '\n';
            } else if (nodeName === 'a') {
                return `[[${decodeURI(node.href)} | ${result}]]`;
            } else if (nodeName === 'b' || nodeName === 'strong') {
                return `**${result}**`;
            } else if (nodeName === 'i' || nodeName === 'em') {
                return `//${result}//`;
            } else if (nodeName === 'h1') {
                return `\n\n${result}\n\n`;
            } else if (nodeName === 'h2') {
                return `= ${result}\n\n`;
            } else if (nodeName === 'h3') {
                return `== ${result}\n\n`;
            } else if (nodeName === 'h4') {
                return `=== ${result}\n\n`;
            } else if (nodeName === 'h5') {
                return `==== ${result}\n\n`;
            } else if (nodeName === 'li') {
                if (node.children.length === 1) {
                    let link = node.children[0];
                    if (link.nodeName.toLowerCase() === 'a') {
                        if (link.href === link.innerText || decodeURI(link.href) === link.innerText) {
                            return `=> ${decodeURI(link.href)}\n`;
                        } else {
                            return `=> ${decodeURI(link.href)} | ${link.innerText}\n`;
                        }
                    }
                }
                return parentNodeName === 'ol'
                    ? `*. ${result}\n`
                    : `* ${result}\n`;
            } else {
                return result;
            }
        }

        let selection = window.getSelection();
        if (selection.rangeCount === 0) {
            return '';
        }
        let range = selection.getRangeAt(0);
        let contents = range.cloneContents();
        return convert(contents).replace(/\n\n+/g, '\n\n');
    }

    let u = '%s/save-link?' + new URLSearchParams({
        url: ($('link[rel=canonical]') || location).href,
        title: $('meta[property="og:title"]')?.content || document.title,
        description: (
            getSelectionInMycomarkup() ||
            $('meta[property="og:description"]')?.content ||
            $('meta[name=description]')?.content
        )?.trim().replace(/^/gm, '> ') || ''
    });

    try {
        window.open(u, '_blank', 'location=yes,width=600,height=800,scrollbars=yes,status=yes,noopener,noreferrer');
    } catch {
        location.href = u;
    }
})(document.querySelector.bind(document));
