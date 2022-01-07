# TODO

  - [x] Undo
  - [x] Searching through the file with a shortcut
  - [x] Mechanism to run shell commands when buffer is saved
  - [x] Jump to previous and next empty lines with `Alt+Up` and
    `Alt+Down`
  - [x] Opening a file to the current buffer, fuzzily
  - [x] Managing multiple buffers
  - [x] Fuzzy searching should be fuzzier, ie. support matching multiple
    words
  - [x] Configurable tabsize (now hardcoded to 4 spaces)
  - [x] `Shift+Tab` should remove one tab level from line
  - [x] Display status messages somehow
  - [x] `Alt+Backspace` should remove one complete word
  - [x] Search & replace
  - [ ] `Mod+Tab` should tabulate line like the previous line
  - [ ] Make file browsing/buffer opening nonblocking & faster
    (parallelize etc.)
  - [x] Editor should indicate whether a buffer was changed since last
    saving
  - [x] Implement simple syntax highlighting
  - [x] Use configuration file for basic configuration
  - [x] Make highlighting configurable
  - [ ] Create a highlighting fastpath when there are no rules
  - [ ] Savehook execution should have a timeout
  - [ ] File browsing should maybe follow symlinks
  - [ ] Extend undo to work with savehooks
  - [x] Savehooks should be specified with glob patterns
  - [x] Implement an UI dialog for asking single-key answers
  - [x] Confirmation dialogs for quitting & closing nonsaved buffers
  - [ ] Delay buffer opens for command-line arguments until we have a
    full editor with a printable screen & functioning status messages
  - [ ] When doing buffer opens, warn about files which seem less than
    printable
  - [ ] When doing buffer opens, handle files with extremely long lines
    more robustly
