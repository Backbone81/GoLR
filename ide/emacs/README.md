# GoLR for Emacs

This setup provides language support for [GoLR](https://github.com/Backbone81/GoLR) grammar files (`.golr`) in Emacs
with:

- Syntax highlighting, through the major mode `golr-mode`
- Comment commands (`M-;`)
- Through the GoLR language server (`golr lsp`) and Eglot:
  - Go to definition and find references, with string aliases resolved to their terminals
  - Rename, of a symbol or of the string alias of a terminal
  - Code completion of keywords and symbols
  - Format document

## Requirements

- Emacs 29.1 or later, which comes with Eglot
- The `golr` binary on your `PATH`, see [Getting Started](https://github.com/Backbone81/GoLR#getting-started)

## Syntax Highlighting

Clone the GoLR repository:

```shell
git clone https://github.com/Backbone81/GoLR.git ~/src/GoLR
```

Load `golr-mode` in your init file:

```elisp
(add-to-list 'load-path "~/src/GoLR/ide/emacs")
(require 'golr-mode)
```

Emacs now opens `.golr` files in `golr-mode`.

## Language Server

Tell Eglot to start `golr lsp` for `golr-mode`, and start Eglot when a `.golr` file is opened:

```elisp
(with-eval-after-load 'eglot
  (add-to-list 'eglot-server-programs '(golr-mode "golr" "lsp")))
(add-hook 'golr-mode-hook #'eglot-ensure)
```

| Feature | Key or command |
|---|---|
| Go to definition | `M-.` |
| Find references | `M-?` |
| Rename | `M-x eglot-rename` |
| Code completion | `C-M-i` |
| Format document | `M-x eglot-format-buffer` |

The language server also provides semantic highlighting and reference counts above the definitions, which the Eglot
of Emacs 29 does not show.

Format the file with `M-x eglot-format-buffer`. To format on every save, add:

```elisp
(add-hook 'golr-mode-hook
          (lambda () (add-hook 'before-save-hook #'eglot-format-buffer nil t)))
```

The language server writes its log to stderr, which Eglot shows with `M-x eglot-stderr-buffer`.

## Supported Grammar Syntax

The setup supports the `.golr` syntax described in
[Parser Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/parsergen-frontend-golr.md) and
[Scanner Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/scannergen-frontend-golr.md).

## Limitations

Errors in the grammar file are not reported in the editor. Run `golr parser` or `golr scanner` on the file to see them.
