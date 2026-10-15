# GoLR for Vim

This setup provides language support for [GoLR](https://github.com/Backbone81/GoLR) grammar files (`.golr`) in Vim
with:

- Syntax highlighting
- Comment settings (`commentstring`, `comments`) for comment plugins and formatting of comments
- Format document, through the GoLR language server (`golr lsp`) and an LSP plugin

The runtime files in this directory are also used by [Neovim](../neovim).

## Requirements

- Vim 9.0 or later
- An LSP plugin for the language server, e.g. [vim-lsp](https://github.com/prabirshrestha/vim-lsp) or
  [coc.nvim](https://github.com/neoclide/coc.nvim)
- The `golr` binary on your `PATH`, see [Getting Started](https://github.com/Backbone81/GoLR#getting-started)

## Syntax Highlighting

Clone the GoLR repository and link the runtime files into a Vim package:

```shell
git clone https://github.com/Backbone81/GoLR.git ~/src/GoLR
mkdir -p ~/.vim/pack/golr/start
ln -s ~/src/GoLR/ide/vim ~/.vim/pack/golr/start/golr
```

Vim now detects `.golr` files and sets the filetype `golr`, as long as your `vimrc` turns on filetype detection and
syntax highlighting:

```vim
filetype plugin on
syntax on
```

## Language Server with vim-lsp

Register the language server in your `vimrc`:

```vim
if executable('golr')
  au User lsp_setup call lsp#register_server({
        \ 'name': 'golr',
        \ 'cmd': {server_info -> ['golr', 'lsp']},
        \ 'allowlist': ['golr'],
        \ })
endif
```

vim-lsp starts `golr lsp` when you open a `.golr` file. `:LspStatus` shows whether it is running.

Format the file with `:LspDocumentFormat`. To format on every save, add:

```vim
autocmd BufWritePre *.golr LspDocumentFormatSync
```

The language server writes its log to stderr. vim-lsp keeps it only with a log file set, e.g.
`let g:lsp_log_file = expand('~/vim-lsp.log')`.

## Language Server with coc.nvim

Add the language server to your `coc-settings.json` (`:CocConfig`):

```json
{
  "languageserver": {
    "golr": {
      "command": "golr",
      "args": ["lsp"],
      "filetypes": ["golr"]
    }
  }
}
```

Format the file with `:call CocAction('format')`. To format on every save, add to your `coc-settings.json`:

```json
{
  "[golr]": {
    "coc.preferences.formatOnSave": true
  }
}
```

## Supported Grammar Syntax

The setup supports the `.golr` syntax described in
[Parser Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/parsergen-frontend-golr.md) and
[Scanner Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/scannergen-frontend-golr.md).

## Limitations

Errors in the grammar file are not reported in the editor. Run `golr parser` or `golr scanner` on the file to see them.
