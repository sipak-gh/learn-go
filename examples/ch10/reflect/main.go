// reflect inspects types and values at run time.
package main

import (
	"fmt"
	"reflect"
)

type Config struct {
	Name    string `env:"APP_NAME"`
	Port    int    `env:"APP_PORT"`
	Verbose bool
}

func main() {
	for _, v := range []any{42, "hi", 3.5, []int{1}, Config{}} {
		t := reflect.TypeOf(v)
		fmt.Printf("%-12v kind %v\n", t, t.Kind())
	}

	// Walk a struct's fields: names, types, values, and tags.
	c := Config{Name: "seek", Port: 8080}
	t := reflect.TypeOf(c)
	val := reflect.ValueOf(c)
	for i := range t.NumField() {
		f := t.Field(i)
		fmt.Printf("%s %v = %v  env=%q\n", f.Name, f.Type, val.Field(i), f.Tag.Get("env"))
	}
}
