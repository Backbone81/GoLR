// Structural port of internal/fmt/config.go. Formatter's only configurable knob.
export interface Config {
  indentation: string;
}

// Structural port of internal/fmt/config.go's DefaultConfig.
export const DefaultConfig: Config = Object.freeze({ indentation: "    " });
