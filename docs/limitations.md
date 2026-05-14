# Limitations

## Parser Generator

All frontends have the limitation that they can handle 32,768 terminals and 32,768 nonterminals at
maximum. That limitation comes from the data types involved and the encoding of the data. In practice that limitation
should not pose any issues for real world grammars used by humans.

All backends have the limitation that they can handle 65,536 productions at maximum and each production can have a
maximum length of 65,535 symbols. The maximum number of states the generated parser can have is 65,536. Those
limitations come from the data types involved and the encoding of the data. In practice those limitations should not
pose any issue for real world grammars used by humans.

The IELR(1) core implementation delegates the parser generation to GNU Bison for now. It writes out a GNU Bison grammar
file, calls GNU Bison to generate the parser and output an XML report with the parser states. The XML report is then
loaded into the backend parser representation. This means that an up-to-date GNU Bison binary needs to be available
on your system for the IELR(1) core to work.
