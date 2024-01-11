async function copyTextElem(text, elem) {
    await navigator.clipboard.writeText(text)
    elem.textContent = "Copied!"
}
