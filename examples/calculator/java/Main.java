import parser.Scanner;
import parser.Scanner.TokenSkipper;
import parser.Scanner.Token;

import java.nio.charset.StandardCharsets;

public class Main {
    public static void main(String[] args) {
    	// We expect this binary to be used like "calculator '4 + 5 * 3'"
        if (args.length != 1) {
            System.err.println("usage: calculator <expression>");
            System.exit(1);
        }

        tokenize(args[0]);
    }

    private static void tokenize(String expression) {
    	// The generated TokenSkipper will skip all whitespaces which the parser is not interested in.
        TokenSkipper scanner = new TokenSkipper(
    		// The generated Scanner will convert the input into tokens. The filePath argument is used in error messages.
            new Scanner(expression.getBytes(StandardCharsets.UTF_8), "expression")
        );

        while (scanner.next()) {
            Token token = scanner.token();
            if (token == Token.INVALID_TOKEN) {
                System.err.printf(
                    "unexpected input at line %d, column %d%n",
                    scanner.line(),
                    scanner.column());
                System.exit(1);
            }
            String lexeme = StandardCharsets.UTF_8.decode(scanner.lexeme()).toString();
            System.out.printf(
                "line %d, col %-3d  %-20s  %s%n",
                scanner.line(),
                scanner.column(),
                token,
                lexeme);
        }
    }
}
