// Method values and method expressions.
package main

import "fmt"

type Counter struct {
	Name string
	n    int
}

func (c *Counter) Inc()          { c.n++ }
func (c Counter) Report() string { return fmt.Sprintf("%s=%d", c.Name, c.n) }

// repeat calls f n times. It knows nothing about Counter.
func repeat(n int, f func()) {
	for range n {
		f()
	}
}

func main() {
	hits := &Counter{Name: "hits"}

	// A method value binds a method to one receiver: a func() here.
	inc := hits.Inc
	repeat(3, inc)
	fmt.Println(hits.Report())

	// A value receiver is copied when the method value is made.
	report := hits.Report
	hits.Inc()
	fmt.Println(report(), hits.Report())

	// A method expression turns a method into an ordinary function
	// whose first parameter is the receiver.
	incFn := (*Counter).Inc // type func(*Counter)
	reportFn := Counter.Report
	misses := &Counter{Name: "misses"}
	incFn(misses)
	fmt.Println(reportFn(*misses))
	fmt.Printf("%T\n%T\n", inc, incFn)
}
