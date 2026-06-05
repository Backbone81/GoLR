package parser_test

import (
	"context"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"testing"

	"github.com/backbone81/golr/examples/golang/parser"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Golang Scanner", func() {
	Context("should produce identical tokens to the official Go scanner", func() {
		filePaths, err := stdLibSourceFilePaths(context.Background())
		Expect(err).ToNot(HaveOccurred())
		for _, filePath := range filePaths {
			It("should correctly tokenize "+filePath, func(ctx SpecContext) {
				source, err := os.ReadFile(filePath)
				Expect(err).ToNot(HaveOccurred())

				fileSet := token.NewFileSet()
				file := fileSet.AddFile(filepath.Base(filePath), fileSet.Base(), len(source))
				var goScanner scanner.Scanner
				goScanner.Init(file, source, nil, 0)

				golrScanner := parser.SemicolonInserter{
					Scanner: parser.NewTokenSkipper(parser.NewScanner(source, "in-memory")),
				}

				var line []int
				var goScannerTokens []token.Token
				var golrScannerTokens []token.Token
				endOfGoScanner := false
				endofGolrScanner := false
				for {
					_, goToken, _ := goScanner.Scan()
					if goToken == token.EOF {
						endOfGoScanner = true
					}
					if !golrScanner.Next() {
						endofGolrScanner = true
					}
					if endOfGoScanner || endofGolrScanner {
						break
					}
					goScannerTokens = append(goScannerTokens, goToken)
					golrScannerTokens = append(golrScannerTokens, tokenConvert(golrScanner.Token()))
					line = append(line, golrScanner.Line())
				}

				for i := range goScannerTokens {
					Expect(golrScannerTokens[i].String()).To(
						Equal(goScannerTokens[i].String()),
						"Line %d, Go: %s, GoLR: %s",
						line[i],
						goScannerTokens[max(0, i-5):min(len(goScannerTokens), i+5)],
						golrScannerTokens[max(0, i-5):min(len(golrScannerTokens), i+5)])
				}

				Expect(endOfGoScanner).To(
					Equal(endofGolrScanner),
					"Unexpected end of parse at line %d with Go scanner end %v and GoLR scanner end %v",
					golrScanner.Line(),
					endOfGoScanner,
					endofGolrScanner,
				)
			})
		}
	})
})

//nolint:gocyclo,maintidx // There is no way we can simplify this function.
func tokenConvert(tok parser.Token) token.Token {
	//nolint:exhaustive // These are the tokens we are interested in. Others are technical tokens.
	switch tok {
	case parser.TokenBreak:
		return token.BREAK
	case parser.TokenCase:
		return token.CASE
	case parser.TokenChan:
		return token.CHAN
	case parser.TokenConst:
		return token.CONST
	case parser.TokenContinue:
		return token.CONTINUE
	case parser.TokenDefault:
		return token.DEFAULT
	case parser.TokenDefer:
		return token.DEFER
	case parser.TokenElse:
		return token.ELSE
	case parser.TokenFallthrough:
		return token.FALLTHROUGH
	case parser.TokenFor:
		return token.FOR
	case parser.TokenFunc:
		return token.FUNC
	case parser.TokenGo:
		return token.GO
	case parser.TokenGoto:
		return token.GOTO
	case parser.TokenIf:
		return token.IF
	case parser.TokenImport:
		return token.IMPORT
	case parser.TokenInterface:
		return token.INTERFACE
	case parser.TokenMap:
		return token.MAP
	case parser.TokenPackage:
		return token.PACKAGE
	case parser.TokenRange:
		return token.RANGE
	case parser.TokenReturn:
		return token.RETURN
	case parser.TokenSelect:
		return token.SELECT
	case parser.TokenStruct:
		return token.STRUCT
	case parser.TokenSwitch:
		return token.SWITCH
	case parser.TokenType:
		return token.TYPE
	case parser.TokenVar:
		return token.VAR
	case parser.TokenIdentifier:
		return token.IDENT
	case parser.TokenAdd:
		return token.ADD
	case parser.TokenSub:
		return token.SUB
	case parser.TokenMul:
		return token.MUL
	case parser.TokenQuo:
		return token.QUO
	case parser.TokenRem:
		return token.REM
	case parser.TokenAnd:
		return token.AND
	case parser.TokenOr:
		return token.OR
	case parser.TokenXor:
		return token.XOR
	case parser.TokenShiftLeft:
		return token.SHL
	case parser.TokenShiftRight:
		return token.SHR
	case parser.TokenAndNot:
		return token.AND_NOT
	case parser.TokenAddAssign:
		return token.ADD_ASSIGN
	case parser.TokenSubAssign:
		return token.SUB_ASSIGN
	case parser.TokenMulAssign:
		return token.MUL_ASSIGN
	case parser.TokenQuoAssign:
		return token.QUO_ASSIGN
	case parser.TokenRemAssign:
		return token.REM_ASSIGN
	case parser.TokenAndAssign:
		return token.AND_ASSIGN
	case parser.TokenOrAssign:
		return token.OR_ASSIGN
	case parser.TokenXorAssign:
		return token.XOR_ASSIGN
	case parser.TokenShiftLeftAssign:
		return token.SHL_ASSIGN
	case parser.TokenShiftRightAssign:
		return token.SHR_ASSIGN
	case parser.TokenAndNotAssign:
		return token.AND_NOT_ASSIGN
	case parser.TokenLogicalAnd:
		return token.LAND
	case parser.TokenLogicalOr:
		return token.LOR
	case parser.TokenArrow:
		return token.ARROW
	case parser.TokenIncrement:
		return token.INC
	case parser.TokenDecrement:
		return token.DEC
	case parser.TokenEqual:
		return token.EQL
	case parser.TokenLessThan:
		return token.LSS
	case parser.TokenGreaterThan:
		return token.GTR
	case parser.TokenAssign:
		return token.ASSIGN
	case parser.TokenNot:
		return token.NOT
	case parser.TokenTilde:
		return token.TILDE
	case parser.TokenNotEqual:
		return token.NEQ
	case parser.TokenLessEqual:
		return token.LEQ
	case parser.TokenGreaterEqual:
		return token.GEQ
	case parser.TokenDefine:
		return token.DEFINE
	case parser.TokenEllipsis:
		return token.ELLIPSIS
	case parser.TokenLeftParen:
		return token.LPAREN
	case parser.TokenLeftBracket:
		return token.LBRACK
	case parser.TokenLeftBrace:
		return token.LBRACE
	case parser.TokenComma:
		return token.COMMA
	case parser.TokenPeriod:
		return token.PERIOD
	case parser.TokenRightParen:
		return token.RPAREN
	case parser.TokenRightBracket:
		return token.RBRACK
	case parser.TokenRightBrace:
		return token.RBRACE
	case parser.TokenSemicolon:
		return token.SEMICOLON
	case parser.TokenColon:
		return token.COLON
	case parser.TokenIntLit:
		return token.INT
	case parser.TokenFloatLit:
		return token.FLOAT
	case parser.TokenImaginaryLit:
		return token.IMAG
	case parser.TokenRuneLit:
		return token.CHAR
	case parser.TokenStringLit:
		return token.STRING
	default:
		return token.ILLEGAL
	}
}

func BenchmarkGolangScanner(b *testing.B) {
	source, err := stdLibBenchmarkSourceFilePath(b.Context())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Official Go Scanner", func(b *testing.B) {
		fileSet := token.NewFileSet()
		file := fileSet.AddFile("server.go", fileSet.Base(), len(source))
		for b.Loop() {
			var myScanner scanner.Scanner
			myScanner.Init(file, source, nil, 0)
			for {
				_, tok, _ := myScanner.Scan()
				if tok == token.EOF {
					break
				}
			}
		}
	})

	b.Run("GoLR Generated Scanner", func(b *testing.B) {
		myScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(parser.NewScanner(nil, "in-memory")),
		}
		for b.Loop() {
			myScanner.Reset(source, 0)
			for myScanner.Next() {
				// Read all tokens
			}
		}
	})
}
