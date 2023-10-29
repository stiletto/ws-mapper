package wsmapper

import "github.com/stiletto/ws-mapper/forwarder"

type Listen struct {
	Address string `yaml:"address"`
	Family  string `yaml:"family"`
}

type Config struct {
	Listen Listen  `yaml:"listen"`
	Routes []Route `yaml:"routes"`
}

type Route struct {
	Match  string           `yaml:"match"`
	Target forwarder.Target `yaml:"target"`
}

func DefaultConfig() Config {
	return Config{
		Listen: Listen{Address: "localhost:8400", Family: "tcp"},
		Routes: make([]Route, 0),
	}
}

func CheckAndFixConfig(cfg *Config) error {
	if cfg.Listen.Family == "" {
		cfg.Listen.Family = "tcp"
	}
	for i := range cfg.Routes {
		if cfg.Routes[i].Target.Family == "" {
			cfg.Routes[i].Target.Family = "tcp"
		}
	}
	return nil
}

func ExampleConfig() Config {
	cfg := DefaultConfig()
	cfg.Routes = []Route{
		Route{Match: "/wsto/ssh", Target: forwarder.Target{Address: "127.0.0.1:22", Family: "tcp"}},
		Route{Match: "/wsto/postgres", Target: forwarder.Target{Address: "/var/run/postgresql/.s.PGSQL.5432", Family: "unix"}},
	}
	return cfg
}
