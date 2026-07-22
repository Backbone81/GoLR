# Limitations

## Parser Generator

All frontends have the limitation that they can handle 32,768 terminals and 32,768 nonterminals at
maximum. That limitation comes from the data types involved and the encoding of the data. In practice that limitation
should not pose any issues for real world grammars used by humans.

All backends have the limitation that they can handle 65,536 productions at maximum and each production can have a
maximum length of 65,535 symbols. The maximum number of states the generated parser can have is 65,536. Those
limitations come from the data types involved and the encoding of the data. In practice those limitations should not
pose any issue for real world grammars used by humans.
