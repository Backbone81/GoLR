package convert

import intconvert "github.com/backbone81/golr/internal/convert"

// BisonToGoLR reads a GNU Bison grammar from reader and writes a GoLR grammar to writer.
var BisonToGoLR = intconvert.BisonToGoLR

// BisonToGoLRFile reads a GNU Bison grammar from inputFilePath and writes a GoLR grammar to outputFilePath.
var BisonToGoLRFile = intconvert.BisonToGoLRFile

// BisonToGoLRString reads a GNU Bison grammar from bisonGrammar and returns a GoLR grammar as return value.
var BisonToGoLRString = intconvert.BisonToGoLRString
