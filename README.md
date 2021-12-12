# ked

`ked` is a text editor. It is purpose-built for myself as a code editor
for use within a terminal multiplexer.

## disclaimer

`ked` is unsupported software and I advise against using it for real.
There are probably a lot of edgecases I have yet to find. You may
encounter bugs, which will cause `ked` to crash, which will cause you to
lose your buffer modifications irrevocably. [I may make](TODO.md) minor
bug fixes, modifications, and improvements, but `ked` will never be much
more than it is now. Presently it works well enough that after
bootstrapping the project with another editor, `ked` itself has mostly
been with developed with `ked`.

## shortcuts

Presently the only way to change keyboard shortcuts is by editing the
code. As ked is built with Go without many dependencies, recompiling is
trivial.

The hardcoded keyboard shortcuts are the following:

-   `Ctrl+C` exits the editor and also cancels dialogs
-   `Ctrl+W` saves the buffer
-   `Ctrl+S` searches the buffer (use `Ctrl+S` to jump through results)
-   `Alt+Left` and `Alt+Right` jump over wordish things
-   `Ctrl+A` and `Ctrl+E` move cursor to beginning and end of present
    line
-   `Ctrl+G` jumps to a specific line
-   `PageUp` and `PageDown` move, well, a single page up or down
-   `Ctrl+K` deletes from cursor to the end of line; also deletes empty
    lines
-   `Alt+Backspace` deletes current word
-   `Ctrl+_` undos recent actions
-   `Tab` inserts one tab character to cursor position
-   `Shift+Tab` (`Backtab`) removes one level of tabulation from line
    beginning
-   `Alt+Up` and `Alt+Down` jump to the previous or next empty line
-   `Ctrl+P` displays the buffer selection dialog
-   `Ctrl+F` displays the file-open dialog
-   `Alt+F` closes the current buffer

Depending your terminal settings, `Alt` may be mapped to `Esc`.

## buffer management

We have a very minimalistic approach to buffer handling. You can open
new buffers, close them, save their contents to a file, and change
between them. A single buffer always occupies the available screen
space. I use it within `tmux`, and I need to see more than a single
buffer at a time, I will open two panes for it.

Opening files into new buffers is based on the idea of first selecting a
root directory and then fuzzily finding filenames matching your filter.
By default, we ignore certain directories like `.git` and
`node_modules`. You may specify these exactly with the `-ignoredirs`
argument.

## save hooks

The command-line argument `-savehooks` may be used to specify
command-lines, which are automatically run after a buffer is saved to a
file. Each hook consists of a match pattern and the command itself.

To make the mechanism more useful, all references to `__ABSPATH__` in
the command-line will be replaced with the current buffer's absolute
path. Thus the final invocation will be like

    /bin/sh -c <savehook-value-with-abspath-expansion>`

If the command returns successfully, `ked` will reload the buffer's
contents from the file.

For example, `gofmt` and `clang-format` may be used to autoformat
buffers upon saving:

    $ ked -savehooks '*.go=gofmt -w __ABSPATH__,*.c=clang-format -i __ABSPATH__'
    $ ked -savehooks '*.md=pandoc -f markdown -t markdown -o __ABSPATH__ __ABSPATH__'

Above we assume that `gofmt`, `clang-format`, and `pandoc` will be found
in the path. Note the `-w`, `-i`, and `-o` parameters, respectively,
which are used to enable formatting the files on disk instead of
standard I/O.
