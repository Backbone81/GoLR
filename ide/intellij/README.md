# GoLR Plugin

This plugin for IntelliJ provides language support for [GoLR](https://github.com/Backbone81/GoLR) grammar files
(`.golr`) with:

- Syntax highlighting
- Code completion for terminal names, nonterminal names and keywords in the context they are valid
- Rename for terminal names, terminal string aliases and nonterminal names
- Go to declaration for terminal names, terminal string aliases and nonterminal names
- Usage count above every terminal and nonterminal declaration, which opens the usages on click
- Find usages for terminal names, terminal string aliases and nonterminal names
- Reformat code
- Comment with line comment and block comment
- Brace matching

## Requirements

IntelliJ IDEA 2025.3 or later. The plugin only depends on the IntelliJ Platform, so other JetBrains IDEs of the same
versions work too.

## Supported Grammar Syntax

The plugin supports the `.golr` syntax of GoLR v0.5.0, which is described in
[Parser Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/parsergen-frontend-golr.md) and
[Scanner Generator Frontend GoLR](https://github.com/Backbone81/GoLR/blob/main/docs/scannergen-frontend-golr.md).

## Installation

Install the plugin in Settings | Plugins | Marketplace by searching for "GoLR".

To install it from a `.zip` file, select "Install Plugin from Disk..." from the gear icon in Settings | Plugins.

## Limitations

Errors in the grammar file are not reported in the editor. Run `golr parser` or `golr scanner` on the file to see them.
