package com.backbone81.golr

// Structural port of internal/fmt/config.go. Formatter's only configurable knob.
data class Config(
    val indentation: String = "    ",
)

// Structural port of internal/fmt/config.go's DefaultConfig.
val DefaultConfig = Config()
