// Struct embedding: composition with promoted fields and methods.
package main

import "fmt"

// Logger writes messages with a prefix.
type Logger struct {
	Prefix string
}

func (l *Logger) Log(msg string) { fmt.Println(l.Prefix + msg) }

// A Server embeds a Logger: the field has a type but no name.
type Server struct {
	Logger // embedded; the field's name is Logger
	Addr   string
}

// Start calls Log as if it were Server's own method.
func (s *Server) Start() { s.Log("listening on " + s.Addr) }

// A Cache also embeds Logger, but defines its own Log.
type Cache struct {
	Logger
}

// Log shadows the promoted Logger.Log.
func (c *Cache) Log(msg string) { c.Logger.Log("[cache] " + msg) }

func main() {
	s := &Server{Logger: Logger{Prefix: "srv: "}, Addr: ":8080"}
	s.Start()
	s.Log("promoted method")     // same as s.Logger.Log(...)
	fmt.Println(s.Prefix)        // promoted field: same as s.Logger.Prefix
	s.Logger.Prefix = "server: " // the embedded value is an ordinary field
	s.Log("changed prefix")

	c := &Cache{Logger{Prefix: "app: "}}
	c.Log("hit") // the outer Log wins
}
