# GoLR for Neovim

This setup provides language support for [GoLR](https://github.com/Backbone81/GoLR) grammar files (`.golr`) in Neovim
with:

- Syntax highlighting
- Toggle line comments (`gc`)
- Through the GoLR language server (`golr lsp`):
  - Go to definition and find references, with string aliases resolved to their terminals
  - Rename, of a symbol or of the string alias of a terminal
  - Code completion of keywords and symbols
  - Semantic highlighting of terminals and nonterminals, definitions apart from references
  - Reference counts above every definition (code lens)
  - Format document

Neovim uses the Vim syntax files from [`ide/vim`](../vim) and its built-in LSP client.

## Requirements

- Neovim 0.11 or later
- The `golr` binary on your `PATH`, see [Getting Started](https://github.com/Backbone81/GoLR#getting-started)

## Syntax Highlighting

Clone the GoLR repository and link the Vim runtime files into a Neovim package:

```shell
git clone https://github.com/Backbone81/GoLR.git ~/src/GoLR
mkdir -p ~/.local/share/nvim/site/pack/golr/start
ln -s ~/src/GoLR/ide/vim ~/.local/share/nvim/site/pack/golr/start/golr
```

Neovim now detects `.golr` files and sets the filetype `golr`.

## Language Server

Add the language server to your `init.lua`:

```lua
vim.lsp.config('golr', {
  cmd = { 'golr', 'lsp' },
  filetypes = { 'golr' },
})
vim.lsp.enable('golr')
```

Neovim starts `golr lsp` when you open a `.golr` file. `:checkhealth vim.lsp` shows whether it is attached.

Format the file with `:lua vim.lsp.buf.format()`. To format on every save, add:

```lua
vim.api.nvim_create_autocmd('BufWritePre', {
  pattern = '*.golr',
  callback = function(args)
    vim.lsp.buf.format({ bufnr = args.buf })
  end,
})
```

The features use the default LSP keymaps of Neovim:

| Feature | Key or command |
|---|---|
| Go to definition | `CTRL-]` |
| Find references | `grr` |
| Rename | `grn` |
| Code completion | `CTRL-X CTRL-O` in insert mode |
| Format document | `:lua vim.lsp.buf.format()` |

Semantic highlighting is on by default. The reference counts show after enabling code lenses, e.g. in your
`init.lua`:

```lua
vim.lsp.codelens.enable()
```

This is the API of Neovim 0.12. Neovim 0.11 shows the code lenses after `:lua vim.lsp.codelens.refresh()`, see
`:help lsp-codelens`. A reference count only informs, running it (`grx` in Neovim 0.12) does nothing.

The language server writes its log to stderr, which Neovim copies into its LSP log file. Open it with
`:lua vim.cmd.tabnew(vim.lsp.log.get_filename())`.

## Supported Grammar Syntax

The setup supports the `.golr` syntax described in
[Parser Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/parsergen-frontend-golr.md) and
[Scanner Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/scannergen-frontend-golr.md).

## Limitations

Errors in the grammar file are not reported in the editor. Run `golr parser` or `golr scanner` on the file to see them.
