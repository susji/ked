# ked

`ked` is a text editor for the terminal. It is purpose-built for myself
as a code editor embedded in `tmux`.

## shortcuts

Presently the only way to change keyboard shortcuts is by editing the
code. As ked is built with Go without many dependencies, recompiling
is trivial.

The hardcoded keyboard shortcuts are the following:

  * `Ctrl+C` exits the editor and also cancels dialogs
  * `Ctrl+W` saves the buffer
  * `Ctrl+S` searches the buffer (use `Ctrl+S` to jump through results)
  * `Alt+Left` and `Alt+Right` jump over wordish things
  * `Ctrl+A` and `Ctrl+E` move cursor to beginning and end of present line
  * `Ctrl+G` jumps to a specific line
  * `PageUp` and `PageDown` move, well, a single page up or down
  * `Ctrl+K` deletes from cursor to the end of line; also deletes empty lines
  * `Alt+Backspace` deletes current word
  * `Ctrl+_` undos recent actions
  * `Tab` inserts one tab character to cursor position
  * `Shift+Tab` (`Backtab`) removes one level of tabulation from line beginning
  * `Alt+Up` and `Alt+Down` jump to the previous or next empty line
  * `Ctrl+P` displays the buffer selection dialog
  * `Ctrl+F` displays the file-open dialog
  * `Alt+F` closes the current buffer
