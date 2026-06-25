# GoLR Extension

This extension for Visual Studio Code provides language support for [GoLR](https://github.com/backbone81/golr) grammar
files (`.golr`).

It provides:

- Syntax highlighting
- Code completion
- Go to definition
- Rename symbol
- Find all references
- Format document

## Release Notes

### v0.1.0

Initial release.

## Development

The plugin is developed in TypeScript.

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

The installation file can then be found in the root of the plugin.
