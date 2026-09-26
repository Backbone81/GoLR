# Development

The extension is developed in TypeScript and needs Node.js and npm.

To run the extension with a new instance of the IDE:

```shell
make run
```

To only build the extension:

```shell
make build
```

To run the tests:

```shell
make test
```

To package the extension as a file for manual installation from disk:

```shell
make package
```

The installation file can then be found in the root of the extension.

## Release

To release a new version:

1. Set the version with `npm version X.Y.Z --no-git-tag-version`, which updates `package.json` and
   `package-lock.json`.
2. Run `changelog release --tag vX.Y.Z` and `changelog update` in this directory.
3. Commit the changes, tag the commit with `vscode/vX.Y.Z` and push the tag.
4. Publish the file built by `make package` on the Visual Studio Marketplace.
