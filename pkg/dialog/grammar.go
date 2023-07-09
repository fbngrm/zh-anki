package dialog

type Section struct {
	Head        string      `yaml:"head"`
	Description string      `yaml:"description"`
	Structures  []Structure `yaml:"structures"`
}

type Structure struct {
	Head               string   `yaml:"head"`
	Structure          string   `yaml:"structure"`
	Description        string   `yaml:"description"`
	Examples           []string `yaml:"examples"`
	ExampleDescription string   `yaml:"exampleDescription"`
}

type Grammar struct {
	Head        string    `yaml:"head"`
	Description string    `yaml:"description"`
	Sections    []Section `yaml:"sections"`
}
