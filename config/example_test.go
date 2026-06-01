package config_test

import (
	"fmt"
	"os"

	"github.com/azcov/gokit/config"
)

type AppConfig struct {
	Name    string  `config:"name" validate:"required"`
	Port    int     `config:"port"`
	Version string  `config:"version"`
	DB      DBConfig `config:"db"`
}

type DBConfig struct {
	Host string `config:"host"`
	Port int    `config:"port"`
}

func ExampleLoad() {
	os.Setenv("NAME", "myapp")
	os.Setenv("PORT", "8080")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	defer os.Unsetenv("NAME")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")

	var cfg AppConfig
	err := config.Load(&cfg, config.WithEnv())
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	fmt.Println(cfg.Port)
	fmt.Println(cfg.DB.Host)
	fmt.Println(cfg.DB.Port)
	// Output:
	// myapp
	// 8080
	// localhost
	// 5432
}

func ExampleNew() {
	os.Setenv("NAME", "generic-app")
	defer os.Unsetenv("NAME")

	type SimpleConfig struct {
		Name string `config:"name"`
	}

	cfg, err := config.New[SimpleConfig](config.WithEnv())
	if err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output:
	// generic-app
}

func ExampleFromEnv() {
	os.Setenv("VAL", "hello")
	defer os.Unsetenv("VAL")

	cfg := &struct {
		Val string `config:"val"`
	}{}

	src := config.FromEnv()
	if err := src.Load(cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Val)
	// Output:
	// hello
}

func ExampleChain() {
	os.Setenv("NAME", "chained")
	defer os.Unsetenv("NAME")

	cfg := &struct {
		Name string `config:"name"`
	}{}

	loader := config.Chain(config.FromEnv())
	if err := loader.Load(cfg); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Name)
	// Output:
	// chained
}
