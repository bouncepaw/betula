# SPDX-FileCopyrightText: none
#
# SPDX-License-Identifier: CC0-1.0

import os, re, subprocess, sys

ROOT: str = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

def guess_version() -> str:
    with open(os.path.join(ROOT, 'web', 'views', 'about.gohtml')) as f:
        content = f.read()
    # Extract version.
    m = re.search(r'<dd>(\d+\.\d+[\.\d]*)</dd>', content)
    if m is None:
        sys.exit(
            "Could not find the version in web/views/about.gohtml. "
            "Make sure it contains a line like <dd>1.8.0</dd>."
        )
    return 'v' + m.group(1)

def pick_version() -> str:
    result = subprocess.run(
        ['git', 'tag', '--sort=-creatordate'],
        capture_output=True, text=True, cwd=ROOT
    )
    tags = [t for t in result.stdout.strip().splitlines() if t]
    for i, tag in enumerate(tags):
        print(f'  {i + 1}. {tag}')
    while True:
        try:
            n = int(input('Enter number: ').strip())
            if 1 <= n <= len(tags):
                return tags[n - 1]
        except (ValueError, EOFError):
            pass

def shortlog(version: str) -> str:
    result = subprocess.run(
        ['git', 'shortlog', f'{version}...HEAD'],
        capture_output=True, text=True, cwd=ROOT
    )
    return result.stdout.strip()

if __name__ == '__main__':
    print('genlog.py will help you prepare a Betula release.')

    while True:
        print('\nStep 1. New version')
        guessed = guess_version()
        answer = input(f'Is it {guessed}? [Enter to agree, or type a version]: ').strip()
        new_version = guessed if answer == '' else answer

        print('\nStep 2. Pick old version to compare to')
        old_version = pick_version()

        print('\nStep 3. The message')
        optional_text = input('Optional text (empty to skip): ').strip()
        log = shortlog(old_version)

        parts = [f'Betula {new_version}']
        if optional_text:
            parts.append('')
            parts.append(optional_text)
        parts.append('')
        parts.append('Release notes:')
        parts.append('')
        parts.append(f'=> https://joinbetula.org/{new_version}.html')
        parts.append('')
        parts.append(log)
        msg = '\n'.join(parts)

        print()
        print(msg)

        print('\nStep 4. Tag creation')
        action = input('Create tag? [y]es / [r]estart / [q]uit: ').strip().lower()
        if action == 'y':
            subprocess.run(['git', 'tag', '-a', new_version, '-m', msg], cwd=ROOT)
            print(f'Tag {new_version} created.')
            break
        elif action == 'r':
            continue
        else:
            sys.exit(0)
