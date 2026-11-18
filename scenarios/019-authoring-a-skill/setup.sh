# Two editor-setting commits: one subject that names its change, one that doesn't.
set -euo pipefail

printf 'insert_final_newline = true\n' >>dotfiles/.editorconfig
GIT_AUTHOR_DATE=2026-01-02T12:00:00Z GIT_COMMITTER_DATE=2026-01-02T12:00:00Z \
  git commit --quiet --all --message "Require a final newline in text files"

printf 'trim_trailing_whitespace = true\n' >>dotfiles/.editorconfig
GIT_AUTHOR_DATE=2026-01-03T12:00:00Z GIT_COMMITTER_DATE=2026-01-03T12:00:00Z \
  git commit --quiet --all --message "Update editor settings"
