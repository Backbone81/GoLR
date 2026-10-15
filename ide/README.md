# IDE Plugins

This directory contains the IDE plugins for GoLR grammar files (`.golr`):

- [IntelliJ Plugin](intellij/README.md), published on the
  [JetBrains Marketplace](https://plugins.jetbrains.com/plugin/34604-golr)
- [Visual Studio Code Extension](vscode/README.md), published on the
  [Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=backbone81.golr)

Both provide syntax highlighting, navigation between declarations and references, rename, code completion and
formatting.

## Editor Setups

Other editors connect to the GoLR language server (`golr lsp`) with their own LSP client. These directories contain the
syntax highlighting and a README with the configuration of the language server:

- [Neovim](neovim/README.md)
- [Vim](vim/README.md)
- [Emacs](emacs/README.md)

They provide syntax highlighting and formatting.
