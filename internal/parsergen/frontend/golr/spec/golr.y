// This GNU Bison grammar file describes the syntax of GoLR grammar files.

%token SCANNER    "@scanner"
%token PARSER     "@parser"
%token PRECEDENCE "@precedence"
%token START      "@start"
%token LEFT       "@left"
%token RIGHT      "@right"
%token NONE       "@none"
%token SKIP       "@skip"
%token EMPTY      "@empty"
%token FRAGMENT   "@fragment"
%token LBRACE     "{"
%token RBRACE     "}"
%token LPAREN     "("
%token RPAREN     ")"
%token COLON      ":"
%token SEMI       ";"
%token PIPE       "|"
%token COMMA      ","

%token NAME
%token REGEX
%token STRING

%%

file
    : scanner_section parser_section
    ;

// ================================================================================
// Scanner section
// ================================================================================

scanner_section
    : "@scanner" "{" scanner_decl_list "}"
    ;

scanner_decl_list
    : %empty
    | scanner_decl_list scanner_decl
    ;

scanner_decl
    : NAME ":" scanner_decl_rhs ";"
    ;

scanner_decl_rhs
    : scanner_pattern scanner_annotation_list
    | "@empty" scanner_annotation_list
    ;

scanner_pattern
    : REGEX
    | STRING
    ;

scanner_annotation_list
    : %empty
    | scanner_annotation_list scanner_annotation
    ;

scanner_annotation
    : "@skip"
    | "@fragment"
    ;

// ================================================================================
// Parser section
// ================================================================================

parser_section
    : "@parser" "{" start_decl precedence_section rule_decl_list "}"
    ;

start_decl
    : %empty
    | "@start" ":" NAME ";"
    ;

precedence_section
    : %empty
    | "@precedence" "{" precedence_decl_list "}"
    ;

precedence_decl_list
    : %empty
    | precedence_decl_list precedence_decl
    ;

precedence_decl
    : associativity ":" symbol_list ";"
    ;

associativity
    : "@left"
    | "@right"
    | "@none"
    | "@precedence"
    ;

rule_decl_list
    : %empty
    | rule_decl_list production_decl
    ;

production_decl
    : NAME ":" alternative_list ";"
    ;

alternative_list
    : alternative
    | alternative_list "|" alternative
    ;

alternative
    : symbol_list alternative_annotation_list
    | "@empty" alternative_annotation_list
    ;

alternative_annotation_list
    : %empty
    | alternative_annotation_list alternative_annotation
    ;

alternative_annotation
    : "@precedence" "(" symbol ")"
    ;

symbol_list
    : symbol
    | symbol_list symbol
    ;

symbol
    : NAME
    | STRING
    ;
