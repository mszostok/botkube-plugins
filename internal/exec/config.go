package exec

import (
	"strings"
)

type (
	Config struct {
		Interactive []Interactive `yaml:"interactive"`
	}
	Interactive struct {
		Message Message            `yaml:"message"`
		Command InteractiveCommand `yaml:"command"`
	}

	InteractiveCommand struct {
		Parser string `yaml:"parser"`
		Prefix string `yaml:"prefix"`
	}

	Message struct {
		Select  Select            `yaml:"select"`
		Actions map[string]string `yaml:"actions"`
		Preview string            `yaml:"preview"`
	}

	Select struct {
		Name    string `yaml:"name"`
		ItemKey string `yaml:"itemKey"`
		ItemIdx int    `json:"-"`
		Replace bool   `json:"-"`
	}
)

func (e Config) GetInteractiveConfig(cmd string) (Interactive, bool) {
	for _, item := range e.Interactive {
		if strings.HasPrefix(cmd, item.Command.Prefix) {
			return item, true
		}
	}

	return Interactive{}, false
}
