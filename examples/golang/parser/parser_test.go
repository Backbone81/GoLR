package parser_test

import (
	"context"
	goparser "go/parser"
	gotoken "go/token"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/backbone81/golr/examples/golang/parser"
)

var _ = Describe("Golang Parser", func() {
	It("should parse a hello world", func() {
		source := `
			package main

			import "fmt"

			func main() {
				fmt.Println("Hello world!")
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse an if block as block and not as a struct literal", func() {
		source := `
			package main

			func foo() {
				if true {
					bar = 0
				}
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a make call", func() {
		source := `
			package main

			func foo() {
				bar := make([]*baz, 10)
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a return on the same line as the closing brace", func() {
		source := `
			package main

			func foo() { return }
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a make call for channel of function type", func() {
		source := `
			package main

			func foo() {
				bar := make(chan func())
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a type cast inside an if clause", func() {
		source := `
			package main

			func foo() {
				if perr, ok := err.(*ParseError); !ok {
					panic("error")
				}
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a type constraint with or", func() {
		source := `
			package main

			func foo[bytes []byte | string]() {
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse an embedded type on the same line as the closing brace", func() {
		source := `
			package main

			type foo struct{ bar }
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a switch with a semicolon", func() {
		source := `
			package main

			func foo() {
				switch bar := baz; {
				}
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a casted pointer", func() {
		source := `
			package main

			func foo() {
				bar := (*[4]byte)(nil)
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse an interface on a single line", func() {
		source := `
			package main

			type foo interface{ bar() }
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a generic function call", func() {
		source := `
			package main

			var foo = bar.baz[[]byte]()
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse new with generic type", func() {
		source := `
			package main

			func main() {
				foo(new(bar[any, any]))
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a generic function call", func() {
		source := `
			package main

			func main() {
				foo[bar, []int32]()
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a struct literal with new", func() {
		source := `
			package main

			var foo = []struct {
				bar any
			}{
				{new(<-chan int)},
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse new with pointer of pointer", func() {
		source := `
			package main

			func main() {
				foo = new(**[3]int)
			}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	It("should correctly parse a generic function", func() {
		source := `
			package main

			func F[foo *bar]() {}
		`
		golrScanner := parser.SemicolonInserter{
			Scanner: parser.NewTokenSkipper(
				parser.NewScanner([]byte(source), "in-memory"),
			),
		}

		golrParser := parser.NewParser()
		Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
	})

	Context("BasicLit", func() {
		// Integer literals (https://go.dev/ref/spec#Integer_literals)
		DescribeTable("Parsing int literals",
			func(literal string) {
				source := `
					@TestBasicLit ` + literal + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				ast, err := golrParser.Parse(golrScanner)
				Expect(err).ToNot(HaveOccurred())
				Expect(ast).To(Equal(buildBasicLitAST(parser.TokenIntLit, literal)), literal)
			},
			Entry("Integer", "42"),
			Entry("Integer", "4_2"),
			Entry("Integer", "170141183460469231731687303715884105727"),
			Entry("Integer", "170_141183_460469_231731_687303_715884_105727"),

			Entry("Octal integer", "0600"),
			Entry("Octal integer", "0_600"),
			Entry("Octal integer", "0o600"),
			Entry("Octal integer", "0O600"),

			Entry("Hex integer", "0xBadFace"),
			Entry("Hex integer", "0xBad_Face"),
			Entry("Hex integer", "0x_67_7a_2f_cc_40_c6"),
		)

		// Floating-point literals (https://go.dev/ref/spec#Floating-point_literals)
		DescribeTable("Parsing float literals",
			func(literal string) {
				source := `
					@TestBasicLit ` + literal + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				ast, err := golrParser.Parse(golrScanner)
				Expect(err).ToNot(HaveOccurred())
				Expect(ast).To(Equal(buildBasicLitAST(parser.TokenFloatLit, literal)))
			},
			Entry("Decimal float", "0."),
			Entry("Decimal float", "72.40"),
			Entry("Decimal float", "072.40"),
			Entry("Decimal float", "2.71828"),
			Entry("Decimal float", "1.e+0"),
			Entry("Decimal float", "6.67428e-11"),
			Entry("Decimal float", "1E6"),
			Entry("Decimal float", ".25"),
			Entry("Decimal float", ".12345E+5"),
			Entry("Decimal float", "1_5."),
			Entry("Decimal float", "0.15e+0_2"),

			Entry("Hex float", "0x1p-2"),
			Entry("Hex float", "0x2.p10"),
			Entry("Hex float", "0x1.Fp+0"),
			Entry("Hex float", "0X.8p-0"),
			Entry("Hex float", "0X_1FFFP-16"),
		)

		// Imaginary literals (https://go.dev/ref/spec#Imaginary_literals)
		DescribeTable("Parsing imaginary literals",
			func(literal string) {
				source := `
					@TestBasicLit ` + literal + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				ast, err := golrParser.Parse(golrScanner)
				Expect(err).ToNot(HaveOccurred())
				Expect(ast).To(Equal(buildBasicLitAST(parser.TokenImaginaryLit, literal)))
			},
			Entry("Imaginary", "0i"),
			Entry("Imaginary", "0123i"),
			Entry("Imaginary", "0o123i"),
			Entry("Imaginary", "0xabci"),
			Entry("Imaginary", "0.i"),
			Entry("Imaginary", "2.71828i"),
			Entry("Imaginary", "1.e+0i"),
			Entry("Imaginary", "6.67428e-11i"),
			Entry("Imaginary", "1E6i"),
			Entry("Imaginary", ".25i"),
			Entry("Imaginary", ".12345E+5i"),
			Entry("Imaginary", "0x1p-2i"),
		)

		// Rune literals (https://go.dev/ref/spec#Rune_literals)
		DescribeTable("Parsing rune literals",
			func(literal string) {
				source := `
					@TestBasicLit ` + literal + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				ast, err := golrParser.Parse(golrScanner)
				Expect(err).ToNot(HaveOccurred())
				Expect(ast).To(Equal(buildBasicLitAST(parser.TokenRuneLit, literal)))
			},
			Entry("Rune literal", `'a'`),
			Entry("Rune literal", `'ä'`),
			Entry("Rune literal", `'本'`), //nolint:gosmopolitan
			Entry("Rune literal", `'\t'`),
			Entry("Rune literal", `'\000'`),
			Entry("Rune literal", `'\007'`),
			Entry("Rune literal", `'\377'`),
			Entry("Rune literal", `'\x07'`),
			Entry("Rune literal", `'\xff'`),
			Entry("Rune literal", `'\u12e4'`),
			Entry("Rune literal", `'\U00101234'`),
			Entry("Rune literal", `'\''`),
		)

		// String literals (https://go.dev/ref/spec#String_literals)
		DescribeTable("Parsing string literals",
			func(literal string) {
				source := `
					@TestBasicLit ` + literal + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				ast, err := golrParser.Parse(golrScanner)
				Expect(err).ToNot(HaveOccurred())
				Expect(ast).To(Equal(buildBasicLitAST(parser.TokenStringLit, literal)))
			},
			Entry("String literal", "`abc`"),
			Entry("String literal", "`\\n\n\\n`"),
			Entry("String literal", `"\n"`),
			Entry("String literal", `"\""`),
			Entry("String literal", `"Hello, world!\n"`),
			Entry("String literal", `"日本語"`),               //nolint:gosmopolitan
			Entry("String literal", `"\u65e5本\U00008a9e"`), //nolint:gosmopolitan
			Entry("String literal", `"\xff\u00FF"`),
			Entry("String literal", `"\uD800"`),
			Entry("String literal", `"\U00110000"`),
		)
	})

	Context("Expression", func() {
		DescribeTable("Parsing expressions",
			func(expression string) {
				source := `
					@TestExpression ` + expression + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				Expect(golrParser.Parse(golrScanner)).Error().ToNot(HaveOccurred(), expression)
			},
			Entry("Primary expression", "x"),
			Entry("Primary expression", `(s + ".txt")`),
			Entry("Primary expression", "f(3.1415, true)"),
			Entry("Primary expression", `m["foo"]`),
			Entry("Primary expression", "s[i : j + 1]"),
			Entry("Primary expression", "obj.color"),
			Entry("Primary expression", "f.p[i].x()"),

			Entry("Unary expression", "-foo"),
			Entry("Unary expression", "+64"),
			Entry("Unary expression", "<-bar"),

			Entry("Binary expression", "1<<s"),
			Entry("Binary expression", "5 + 6 * 4"),
			Entry("Binary expression", "foo != bar"),
			Entry("Binary expression", "foo + -bar"),
		)
	})

	Context("Type", func() {
		DescribeTable("Parsing types",
			func(goType string) {
				source := `
					@TestType ` + goType + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				Expect(golrParser.Parse(golrScanner)).Error().ToNot(HaveOccurred(), goType)
			},
			Entry("Array type", "[32]byte"),
			Entry("Array type", "[2*N] struct { x, y int32; }"),
			Entry("Array type", "[1000]*float64"),
			Entry("Array type", "[3][5]int"),
			Entry("Array type", "[2][2][2]float64"),

			Entry("Struct type", "struct {}"),
			Entry("Struct type", `struct {
				x, y int;
				u float32;
				_ float32;
				A *[]int;
			}`),
			Entry("Struct type", `struct {
				T1
				*T2;
				P.T3;
				*P.T4;
				x, y int;
			}`),
			Entry("Struct type", `struct {
				x, y float64 "";
				name string  "any string is permitted as a tag";
				_    [4]byte "ceci n'est pas un champ de structure";
			}`),
			Entry("Struct type", `struct {
				microsec  uint64 "protobuf:\"1\"";
				serverIP6 uint64 "protobuf:\"2\"";
			}`),

			Entry("Pointer type", "*Point"),
			Entry("Pointer type", "*[4]int"),

			Entry("Map type", "map[string]int"),
			Entry("Map type", "map[*T]struct{ x, y float64; }"),
			Entry("Map type", "map[string]interface{}"),

			Entry("Channel type", "chan T"),
			Entry("Channel type", "chan<- float64"),
			Entry("Channel type", "<-chan int"),
			Entry("Channel type", "chan<- chan int"),
			Entry("Channel type", "chan<- <-chan int"),
			Entry("Channel type", "<-chan <-chan int"),
			Entry("Channel type", "chan (<-chan int)"),

			Entry("Function type", "func()"),
			Entry("Function type", "func(x int) int"),
			Entry("Function type", "func(a, _ int, z float32) bool"),
			Entry("Function type", "func(a, b int, z float32) (bool)"),
			Entry("Function type", "func(prefix string, values ...int)"),
			Entry("Function type", "func(a, b int, z float64, opt ...int) (success bool)"),
			Entry("Function type", "func(int, int, float64) (float64, *[]int)"),
			Entry("Function type", "func(n int) func(p *T)"),

			Entry("Interface type", `interface {
				Read([]byte) (int, error);
				Write([]byte) (int, error);
				Close() error;
			}`),
		)
	})

	Context("Statement", func() {
		DescribeTable("Parsing statements",
			func(statement string) {
				source := `
					@TestStatement ` + statement + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				Expect(golrParser.Parse(golrScanner)).Error().ToNot(HaveOccurred(), statement)
			},
			Entry("Labeled statement", `Error: log.Panic("error encountered")`),

			Entry("Send statement", "ch <- 3"),

			Entry("IncDec statement", "x++"),
			Entry("IncDec statement", "x--"),

			Entry("Assignment statement", "x = 1"),
			Entry("Assignment statement", "*p = f()"),
			Entry("Assignment statement", "a[i] = 23"),
			Entry("Assignment statement", "(k) = <-ch"),
			Entry("Assignment statement", "a[i] <<= 2"),
			Entry("Assignment statement", "i &^= 1<<n"),
			Entry("Assignment statement", "x, y = f()"),
			Entry("Assignment statement", "one, two, three = '一', '二', '三'"), //nolint:gosmopolitan
			Entry("Assignment statement", "_ = x"),
			Entry("Assignment statement", "x, _ = f()"),
			Entry("Assignment statement", "a, b = b, a"),
		)
	})

	Context("Declaration", func() {
		DescribeTable("Parsing declarations",
			func(declaration string) {
				source := `
					@TestDecl ` + declaration + `
				`
				golrScanner := parser.NewTokenSkipper(
					parser.NewScanner([]byte(source), "in-memory"),
				)

				golrParser := parser.NewParser()
				Expect(golrParser.Parse(golrScanner)).Error().ToNot(HaveOccurred(), declaration)
			},
			Entry("Constant declaration", "const Pi float64 = 3.14159265358979323846"),
			Entry("Constant declaration", "const zero = 0.0"),
			Entry("Constant declaration", `const (
				size int64 = 1024;
				eof        = -1;
			)`),
			Entry("Constant declaration", `const a, b, c = 3, 4, "foo"`),
			Entry("Constant declaration", "const u, v float32 = 0, 3"),
			Entry("Constant declaration", `const (
				Sunday = iota;
				Monday;
				Tuesday;
				Wednesday;
				Thursday;
				Friday;
				Partyday;
				numberOfDays; 
			)`),

			Entry("Type declaration", `type (
				nodeList = []*Node;
				Polar    = polar  ;
			)`),
			Entry("Type declaration", "type set[P comparable] = map[P]bool"),
			Entry("Type declaration", `type (
				Point struct{ x, y float64; };
				polar Point;
			)`),
			Entry("Type declaration", `type TreeNode struct {
				left, right *TreeNode;
				value any;
			}`),
			Entry("Type declaration", `type Block interface {
				BlockSize() int;
				Encrypt(src, dst []byte);
				Decrypt(src, dst []byte);
			}`),
			Entry("Type declaration", "type NewMutex Mutex"),
			Entry("Type declaration", "type PtrMutex *Mutex"),
			Entry("Type declaration", `type PrintableMutex struct {
				Mutex;
			}`),

			Entry("Variable declaration", "var i int"),
			Entry("Variable declaration", "var U, V, W float64"),
			Entry("Variable declaration", "var k = 0"),
			Entry("Variable declaration", "var x, y float32 = -1, -2"),
			Entry("Variable declaration", `var (
				i       int;
				u, v, s = 2.0, 3.0, "bar";
			)`),
			Entry("Variable declaration", "var re, im = complexSqrt(-1)"),
			Entry("Variable declaration", "var _, found = entries[name]"),

			Entry("Function declaration", "func IndexRune(s string, r rune) int {}"),
			Entry("Function declaration", "func flushICache(begin, end uintptr)"),

			Entry("Method declaration", "func (p *Point) Length() float64 {}"),
			Entry("Method declaration", "func (p *Point) Scale(factor float64) {}"),
			Entry("Method declaration", "func (p Pair[A, B]) Swap() Pair[B, A]{}"),
			Entry("Method declaration", "func (p Pair[First, _]) First() First{}"),
		)
	})

	Context("should parse valid Go files", func() {
		filePaths, err := stdLibSourceFilePaths(context.Background())
		Expect(err).ToNot(HaveOccurred())

		for _, filePath := range filePaths {
			It("should correctly parse "+filePath, func(ctx SpecContext) {
				source, err := os.ReadFile(filePath)
				Expect(err).ToNot(HaveOccurred())

				golrScanner := parser.SemicolonInserter{
					Scanner: parser.NewTokenSkipper(
						parser.NewScanner(source, "in-memory"),
					),
				}

				golrParser := parser.NewParser()
				Expect(golrParser.Parse(&golrScanner)).Error().ToNot(HaveOccurred())
			})
		}
	})
})

func buildBasicLitAST(token parser.Token, lexeme string) *parser.Node {
	return &parser.Node{
		Symbol: parser.NewNonterminal(parser.NonterminalSourcefile),
		Children: []*parser.Node{
			{
				Symbol: parser.NewTerminal(parser.TokenTestBasicLit),
				Lexeme: []byte("@TestBasicLit"),
			},
			{
				Symbol: parser.NewNonterminal(parser.NonterminalBasiclit),
				Children: []*parser.Node{
					{
						Symbol: parser.NewTerminal(token),
						Lexeme: []byte(lexeme),
					},
				},
			},
		},
	}
}

func BenchmarkGolangParser(b *testing.B) {
	source, err := stdLibBenchmarkSourceFilePath(b.Context())
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Official Go Parser", func(b *testing.B) {
		for b.Loop() {
			_, err := goparser.ParseFile(gotoken.NewFileSet(), "server.go", source, 0)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("GoLR Generated Parser", func(b *testing.B) {
		for b.Loop() {
			golrScanner := parser.SemicolonInserter{
				Scanner: parser.NewTokenSkipper(
					parser.NewScanner(source, "in-memory"),
				),
			}
			golrParser := parser.NewParser()
			_, err := golrParser.Parse(&golrScanner)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
