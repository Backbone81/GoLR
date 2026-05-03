/* Bison Grammar Parser                             -*- C -*-

   Copyright (C) 2002-2015, 2018-2021 Free Software Foundation, Inc.

   This file is part of Bison, the GNU Compiler Compiler.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.  */

/*
 * This file has been modified on 2026-04-27 to remove all semantic actions and code which is copied verbatim into
 * the generated parser to simplify the content of this file. Aliases have also been removed to allow for easier
 * bootstrapping of the parser, because the Bison XML report outputs the aliases and not the names which is difficult
 * for forming Go identifiers.
 */

%token
  STRING
  TSTRING             

  PERCENT_TOKEN       
  PERCENT_NTERM     

  PERCENT_TYPE        
  PERCENT_DESTRUCTOR  
  PERCENT_PRINTER     

  PERCENT_LEFT        
  PERCENT_RIGHT       
  PERCENT_NONASSOC    
  PERCENT_PRECEDENCE  

  PERCENT_PREC        
  PERCENT_DPREC       
  PERCENT_MERGE       

  PERCENT_CODE            
  PERCENT_DEFAULT_PREC    
  PERCENT_DEFINE          
  PERCENT_ERROR_VERBOSE   
  PERCENT_EXPECT          
  PERCENT_EXPECT_RR       
  PERCENT_FILE_PREFIX     
  PERCENT_FLAG            
  PERCENT_GLR_PARSER      
  PERCENT_HEADER          
  PERCENT_INITIAL_ACTION  
  PERCENT_LANGUAGE        
  PERCENT_NAME_PREFIX     
  PERCENT_NO_DEFAULT_PREC 
  PERCENT_NO_LINES        
  PERCENT_NONDETERMINISTIC_PARSER 
  PERCENT_OUTPUT          
  PERCENT_PURE_PARSER     
  PERCENT_REQUIRE         
  PERCENT_SKELETON        
  PERCENT_START           
  PERCENT_TOKEN_TABLE     
  PERCENT_VERBOSE         
  PERCENT_YACC            

  BRACED_CODE
  BRACED_PREDICATE
  BRACKETED_ID      
  CHAR_LITERAL      
  COLON
  EPILOGUE         
  EQUAL
  ID                
  ID_COLON
  PERCENT_PERCENT
  PIPE
  PROLOGUE
  SEMICOLON
  TAG               
  TAG_ANY
  TAG_NONE          

%token INT_LITERAL 

/*---------.
| %param.  |
`---------*/
%token PERCENT_PARAM;


                     /*==========\
                     | Grammar.  |
                     \==========*/
%%

input:
  prologue_declarations PERCENT_PERCENT grammar epilogue.opt
;


        /*------------------------------------.
        | Declarations: before the first %%.  |
        `------------------------------------*/

prologue_declarations:
  %empty
| prologue_declarations prologue_declaration
;

prologue_declaration:
  grammar_declaration
| PROLOGUE
| PERCENT_FLAG
| PERCENT_DEFINE variable value
| PERCENT_HEADER string.opt
| PERCENT_ERROR_VERBOSE
| PERCENT_EXPECT INT_LITERAL
| PERCENT_EXPECT_RR INT_LITERAL
| PERCENT_FILE_PREFIX STRING
| PERCENT_GLR_PARSER
| PERCENT_INITIAL_ACTION BRACED_CODE
| PERCENT_LANGUAGE STRING
| PERCENT_NAME_PREFIX STRING
| PERCENT_NO_LINES
| PERCENT_NONDETERMINISTIC_PARSER
| PERCENT_OUTPUT STRING
| PERCENT_PARAM params
| PERCENT_PURE_PARSER
| PERCENT_REQUIRE STRING
| PERCENT_SKELETON STRING
| PERCENT_TOKEN_TABLE
| PERCENT_VERBOSE
| PERCENT_YACC
| error SEMICOLON
| /*FIXME: Err?  What is this horror doing here? */ SEMICOLON
;

params:
   params BRACED_CODE
| BRACED_CODE
;


/*----------------------.
| grammar_declaration.  |
`----------------------*/

grammar_declaration:
  symbol_declaration
| PERCENT_START symbols.1
| code_props_type BRACED_CODE generic_symlist
| PERCENT_DEFAULT_PREC
| PERCENT_NO_DEFAULT_PREC
| PERCENT_CODE BRACED_CODE
| PERCENT_CODE ID BRACED_CODE
;

code_props_type:
  PERCENT_DESTRUCTOR
| PERCENT_PRINTER
;

/*---------.
| %union.  |
`---------*/

%token PERCENT_UNION;

union_name:
  %empty
| ID
;

grammar_declaration:
  PERCENT_UNION union_name BRACED_CODE
;


symbol_declaration:
  PERCENT_NTERM nterm_decls[syms]
| PERCENT_TOKEN token_decls[syms]
| PERCENT_TYPE symbol_decls[syms]
| precedence_declarator token_decls_for_prec[syms]
;

precedence_declarator:
  PERCENT_LEFT
| PERCENT_RIGHT
| PERCENT_NONASSOC
| PERCENT_PRECEDENCE
;

string.opt:
  %empty
| STRING
;

tag.opt:
  %empty
| TAG
;

generic_symlist:
  generic_symlist_item
| generic_symlist generic_symlist_item
;

generic_symlist_item:
  symbol
| tag
;

tag:
  TAG
| TAG_ANY
| TAG_NONE
;

/*-----------------------.
| nterm_decls (%nterm).  |
`-----------------------*/

// A non empty list of possibly tagged symbols for %nterm.
//
// Can easily be defined like symbol_decls but restricted to ID, but
// using token_decls allows to reduce the number of rules, and also to
// make nicer error messages on "%nterm 'a'" or '%nterm FOO "foo"'.
nterm_decls:
  token_decls
;

/*-----------------------------------.
| token_decls (%token, and %nterm).  |
`-----------------------------------*/

// A non empty list of possibly tagged symbols for %token or %nterm.
token_decls:
  token_decl.1[syms]
| TAG token_decl.1[syms]
| token_decls TAG token_decl.1[syms]
;

// One or more symbol declarations for %token or %nterm.
token_decl.1:
  token_decl
| token_decl.1 token_decl
;

// One symbol declaration for %token or %nterm.
token_decl:
  id int.opt[num] alias
;

int.opt:
  %empty
| INT_LITERAL
;

alias:
  %empty
| string_as_id
| TSTRING
;


/*-------------------------------------.
| token_decls_for_prec (%left, etc.).  |
`-------------------------------------*/

// A non empty list of possibly tagged tokens for precedence declaration.
//
// Similar to %token (token_decls), but in '%left FOO 1 "foo"', it treats
// FOO and "foo" as two different symbols instead of aliasing them.
token_decls_for_prec:
  token_decl_for_prec.1[syms]
| TAG token_decl_for_prec.1[syms]
| token_decls_for_prec TAG token_decl_for_prec.1[syms]
;

// One or more token declarations for precedence declaration.
token_decl_for_prec.1:
  token_decl_for_prec
| token_decl_for_prec.1 token_decl_for_prec
;

// One token declaration for precedence declaration.
token_decl_for_prec:
  id int.opt[num]
| string_as_id
;


/*-----------------------------------.
| symbol_decls (argument of %type).  |
`-----------------------------------*/

// A non empty list of typed symbols (for %type).
symbol_decls:
  symbols.1[syms]
| TAG symbols.1[syms]
| symbol_decls TAG symbols.1[syms]
;

// One or more symbols.
symbols.1:
  symbol
  | symbols.1 symbol
;

        /*------------------------------------------.
        | The grammar section: between the two %%.  |
        `------------------------------------------*/

grammar:
  rules_or_grammar_declaration
| grammar rules_or_grammar_declaration
;

/* As a Bison extension, one can use the grammar declarations in the
   body of the grammar.  */
rules_or_grammar_declaration:
  rules
| grammar_declaration SEMICOLON
| error SEMICOLON
;

rules:
  id_colon named_ref.opt COLON rhses.1
;

rhses.1:
  rhs
| rhses.1 PIPE rhs
| rhses.1 SEMICOLON
;

%token PERCENT_EMPTY;
rhs:
  %empty
| rhs symbol named_ref.opt
| rhs tag.opt BRACED_CODE[action] named_ref.opt[name]
| rhs BRACED_PREDICATE
| rhs PERCENT_EMPTY
| rhs PERCENT_PREC symbol
| rhs PERCENT_DPREC INT_LITERAL
| rhs PERCENT_MERGE TAG
| rhs PERCENT_EXPECT INT_LITERAL
| rhs PERCENT_EXPECT_RR INT_LITERAL
;

named_ref.opt:
  %empty
| BRACKETED_ID
;


/*---------------------.
| variable and value.  |
`---------------------*/

variable:
  ID
;

value:
  %empty
| ID
| STRING
| BRACED_CODE
;


/*--------------.
| Identifiers.  |
`--------------*/

/* Identifiers are returned as uniqstr values by the scanner.
   Depending on their use, we may need to make them genuine symbols.  */

id:
  ID
| CHAR_LITERAL
;

id_colon:
  ID_COLON
;


symbol:
  id
| string_as_id
;

/* A string used as an ID.  */
string_as_id:
  STRING
;

epilogue.opt:
  %empty
| PERCENT_PERCENT EPILOGUE
;

%%

