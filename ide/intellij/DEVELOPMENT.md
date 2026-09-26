# Development

The plugin is developed in Kotlin and needs JDK 21.

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

## Release

To release a new version:

1. Set `version` in `gradle.properties` to `X.Y.Z`.
2. Run `changelog release --tag vX.Y.Z` and `changelog update` in this directory.
3. Commit the changes, tag the commit with `intellij/vX.Y.Z` and push the tag.
4. Publish the file built by `make package` on the JetBrains Marketplace.
