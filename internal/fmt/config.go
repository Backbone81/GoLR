package fmt

type Config struct {
	Indentation string
}

var DefaultConfig = Config{
	Indentation: "    ",
}
