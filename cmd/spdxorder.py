# SPDX-FileCopyrightText: none
#
# SPDX-License-Identifier: CC0-1.0

from itertools import groupby
from pathlib import Path
import re
import subprocess
import sys


def check(text, name, write=False):
    lines = text.splitlines(keepends=True)
    bodies = [line.rstrip('\r\n') for line in lines]
    matches = [re.match(r'(?://|#|--| \*) SPDX-FileCopyrightText: ', line) for line in bodies]
    failed = False
    for key, group in groupby(range(len(lines)), key=lambda i: matches[i][0] if matches[i] else None):
        if key is None:
            continue
        indices = list(group)
        entries = [bodies[i] for i in indices]
        ordered = sorted(entries)
        if entries != ordered:
            if write:
                for i, entry in zip(indices, ordered):
                    lines[i] = entry + lines[i][len(bodies[i]):]
            else:
                print(f'{name}:{indices[0] + 1}\tSPDX entries are not in lexicographical order')
                failed = True
    return ''.join(lines), failed


if __name__ == '__main__':
    root = Path(__file__).resolve().parent.parent
    paths = subprocess.check_output([
        'git', '-C', str(root), 'ls-files', '-z', '--cached', '--others', '--exclude-standard'
    ])

    failed = False
    for name in sorted(set(paths.decode().split('\0')) - {''}):
        path = root / name
        if path.is_symlink() or not path.is_file():
            continue

        try:
            original = path.read_bytes().decode('utf-8')
        except UnicodeDecodeError:
            continue # Not text file.
        updated, invalid = check(original, name, '--fix' in sys.argv[1:])
        failed |= invalid
        if updated != original:
            path.write_bytes(updated.encode('utf-8'))

    raise SystemExit(int(failed))
