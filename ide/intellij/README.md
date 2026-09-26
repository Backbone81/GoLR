# GoLR Plugin

This plugin for IntelliJ provides language support for [GoLR](https://github.com/backbone81/golr) grammar files (`.golr`).

It provides:

- Syntax highlighting
- Go to declaration for terminal names, terminal string aliases and nonterminal names
- Rename for terminal names, terminal string aliases and nonterminal names
- Find usage for terminal names, terminal string aliases and nonterminal names
- Code completion for terminal names, nonterminal names and keywords in the context they are valid
- Reformat file

## Development

The plugin is developed in Kotlin.

To run the plugin with a new instance of the IDE:

```shell
make run
```

To only build the plugin:

```shell
make build
```

To run the tests:

```shell
make test
```

To package the plugin as a file for manual installation from disk:

```shell
make package
```

The installation file can then be found in the `build/distributions` folder.

To release a new version:

1. Set `version` in `gradle.properties` to `X.Y.Z`.
2. Run `changelog release` in this directory and enter the version as `vX.Y.Z`.
3. Commit the changes, tag the commit with `intellij/vX.Y.Z` and push the tag.
4. Publish the file built by `make package` on the JetBrains Marketplace.
