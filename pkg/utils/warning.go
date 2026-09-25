package utils

import intutils "github.com/backbone81/golr/internal/utils"

// Warning reports an issue which does not stop the build. It implements error, so it can be printed or joined into an
// error like one.
type Warning = intutils.Warning
