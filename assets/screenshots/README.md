# Screenshots

For screenshots make sure that the window is 1600x900 and that the IDE is using its default settings.

To find the correct window:

```shell
wmctrl -l
```

To force the required size:

```shell
wmctrl -r "golr-demo" -e 0,100,100,1600,900
```
