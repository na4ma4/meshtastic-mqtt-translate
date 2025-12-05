package parser

type OptionFunc func(*Config)

func WithOnParseHandler(f OnParseHandler) OptionFunc {
	return func(c *Config) {
		c.OnParseHandler = f
	}
}
