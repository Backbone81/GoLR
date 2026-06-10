# Scanner Generator Backend: Java

This backend outputs a scanner as Java source code. The generated Java code is a directly coded scanner which does not
use a dedicated table, but has the decisions encoded directly in code.

The generated scanner is fully Unicode capable and processes UTF-8 encoded input. The states are named after the rule
which has priority which makes it easy to debug. In case of conflicts between rules, the rule which was specified
earlier has priority over those rules specified later.
