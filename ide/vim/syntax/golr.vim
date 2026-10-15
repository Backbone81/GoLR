" Vim syntax file for GoLR grammar files.

if exists('b:current_syntax')
  finish
endif

syntax match golrSection /@\(scanner\|parser\)\>/
syntax match golrKeyword /@\(skip\|fragment\|empty\|error\|start\|left\|right\|none\|precedence\|name\)\>/
syntax match golrOperator /|/
syntax match golrDelimiter /[;{}]/

" A string or regular expression ends at the end of its line, as in the GoLR scanner.
syntax match golrEscape /\\./ contained
syntax region golrString start=/"/ skip=/\\./ end=/"/ end=/$/ contains=golrEscape
syntax region golrRegex start=+/[/*]\@!+ skip=/\\./ end=+/+ end=/$/ contains=golrEscape

syntax region golrComment start=+/\*+ end=+\*/+ contains=@Spell
syntax match golrComment +//.*$+ contains=@Spell

highlight default link golrSection Structure
highlight default link golrKeyword Keyword
highlight default link golrOperator Operator
highlight default link golrDelimiter Delimiter
highlight default link golrEscape SpecialChar
highlight default link golrString String
highlight default link golrRegex String
highlight default link golrComment Comment

let b:current_syntax = 'golr'
