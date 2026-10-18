# GoLR for Vim

This setup provides language support for [GoLR](https://github.com/Backbone81/GoLR) grammar files (`.golr`) in Vim
with:

- Syntax highlighting
- Comment settings (`commentstring`, `comments`) for comment plugins and formatting of comments
- Through the GoLR language server (`golr lsp`) and an LSP plugin:
  - Go to definition and find references, with string aliases resolved to their terminals
  - Rename, of a symbol or of the string alias of a terminal
  - Code completion of keywords and symbols
  - Semantic highlighting of terminals and nonterminals, definitions apart from references
  - Reference counts of the definitions (code lens)
  - Format document

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

Code completion needs the omnifunc of vim-lsp, and semantic highlighting has to be turned on. Add to your `vimrc`:

```vim
let g:lsp_semantic_enabled = 1
function! s:on_lsp_buffer_enabled() abort
  setlocal omnifunc=lsp#complete
endfunction
augroup lsp_install
  au!
  autocmd User lsp_buffer_enabled call s:on_lsp_buffer_enabled()
augroup END
```

| Feature | Command |
|---|---|
| Go to definition | `:LspDefinition` |
| Find references | `:LspReferences` |
| Rename | `:LspRename` |
| Code completion | `CTRL-X CTRL-O` in insert mode |
| Reference counts | `:LspCodeLens`, which lists the counts of the line; choosing one does nothing |
| Format document | `:LspDocumentFormat` |

vim-lsp shows code lenses only in that list, not in the text.

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

coc.nvim completes keywords and symbols while you type. Semantic highlighting and the reference counts above the
definitions are off by default in coc.nvim. Turn them on in your `coc-settings.json`:

```json
{
  "semanticTokens.enable": true,
  "codeLens.enable": true
}
```

The other features use the mappings of coc.nvim, e.g. in your `vimrc`:

```vim
nmap <silent> gd <Plug>(coc-definition)
nmap <silent> gr <Plug>(coc-references)
nmap <leader>rn <Plug>(coc-rename)
```

A reference count only informs, `<Plug>(coc-codelens-action)` on it does nothing.

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
