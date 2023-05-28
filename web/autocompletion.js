(() => {
  const inputs = document.querySelectorAll('input[name=tags]');
  if (inputs.length > 0) {
    const tags = fetch('/tag')
      .then(resp => resp.text())
      .then(html => {
        let parser = new DOMParser();
        let tags = Array.from(parser.parseFromString(html, 'text/html').querySelectorAll('.mv-tags .u-url')).map(a => a.innerText);
        return tags;
      });

    for (const input of inputs) {
      let resultListElement = document.createElement('ul');
      resultListElement.className = 'autocompletion-list';

      let resultListWrapper = document.createElement('div');
      resultListWrapper.className = 'autocompletion-wrapper';

      resultListWrapper.appendChild(resultListElement);
      input.parentNode.insertBefore(resultListWrapper, input.nextSibling);

      let listShown = false;
      let selectedResult = -1;

      function showList() {
        resultListWrapper.style.display = '';
        listShown = true;
      }

      function hideList() {
        resultListWrapper.style.display = 'none';
        listShown = false;
        selectedResult = -1;
      }

      function updateSelection(newSelection) {
        if (newSelection < -1) {
          newSelection = -1;
        }
        if (newSelection > resultListElement.children.length - 1) {
          newSelection = -1;
        }
        if (selectedResult > -1 && resultListElement.children[selectedResult]) {
          resultListElement.children[selectedResult].classList.remove('selected');
        }
        selectedResult = newSelection;
        if (selectedResult > -1) {
          resultListElement.children[selectedResult].classList.add('selected');
          resultListElement.children[selectedResult].scrollIntoView({ block: 'nearest' });
        }
      }

      async function suggest() {
        let start = input.value.slice(0, input.selectionStart).lastIndexOf(',') + 1;
        // let end = input.selectionStart + input.value.slice(input.selectionStart).indexOf(',');
        // if (end < input.selectionStart) end = input.value.length;
        let end = input.selectionStart;
        let substring = input.value.slice(start, end).trim();
        if (substring === '') {
          hideList();
          return;
        }
        substring = substring.replace(/ +/g, '_');
        let matches = (await tags).filter(name => name.includes(substring));
        if (matches.length === 0) {
          hideList();
          return;
        }
        matches.sort((a, b) => {
          let idx1 = a.indexOf(substring);
          if (idx1 === -1) idx1 = Infinity;
          let idx2 = b.indexOf(substring);
          if (idx2 === -1) idx2 = Infinity;
          return idx1 - idx2;
        });
        while (resultListElement.children[0]) {
          resultListElement.removeChild(resultListElement.children[0]);
        }
        for (let match of matches) {
          let resultElement = document.createElement('li');
          resultElement.className = 'autocompletion-item';
          resultElement.textContent = match.replace(/_/g, ' ');
          resultElement.dataset.name = match;
          resultListElement.appendChild(resultElement);
        }
        showList();
        updateSelection(selectedResult);
      }

      function autocomplete(selection) {
        let start = input.value.slice(0, input.selectionStart).lastIndexOf(',') + 1;
        // let end = input.selectionStart + input.value.slice(input.selectionStart).indexOf(',');
        // if (end < input.selectionStart) end = input.value.length;
        let end = input.selectionStart;

        let substring = input.value.slice(start, end).trim();
        start += input.value.slice(start).indexOf(substring);
        end = start + substring.length;

        const beautifiedName = selection.dataset.name.replace(/_/g, ' ') + ', ';
        input.value = input.value.slice(0, start) + beautifiedName + input.value.slice(end);
        input.selectionStart = input.selectionEnd = start + beautifiedName.length;

        hideList();
      }

      input.addEventListener('input', () => {
        suggest();
      });

      document.addEventListener('selectionchange', () => {
        if (document.activeElement === input) {
          suggest();
        }
      });

      input.addEventListener('keydown', (event) => {
        if (!listShown) {
          return;
        }

        if (event.key === 'ArrowDown') {
          event.preventDefault();
          updateSelection(selectedResult + 1);
        }
        if (event.key === 'ArrowUp') {
          event.preventDefault();
          if (selectedResult === -1) {
            updateSelection(resultListElement.children.length - 1);
          } else {
            updateSelection(selectedResult - 1);
          }
        }
        if (event.key === 'Escape') {
          event.preventDefault();
          hideList();
        }
        if (event.key === 'Enter' && selectedResult > -1) {
          event.preventDefault();
          autocomplete(resultListElement.children[selectedResult]);
        }
      });

      let resultListPressed = false;

      input.addEventListener('blur', (event) => {
        if (resultListPressed) {
          resultListPressed = false;
          return;
        }
        hideList();
      });

      resultListWrapper.addEventListener('mousedown', () => {
        resultListPressed = true;
      });

      resultListWrapper.addEventListener('click', (event) => {
        if (event.target.classList.contains('autocompletion-item')) {
          autocomplete(event.target);
          input.focus();
        }
      });

      hideList();
    }
  }
})();
