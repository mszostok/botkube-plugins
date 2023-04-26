package exec

type (
	Config struct {
		Interactive InteractiveSources `yaml:"interactive"`
	}
	InteractiveSources struct {
		Templates []Template `yaml:"templates"`
	}

	Template struct {
		Ref string `yaml:"ref"`
	}
)
