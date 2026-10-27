package main

import (
	"fmt"
	"time"
)

func main() {
	// Layouts are written with the reference time:
	//   Mon Jan 2 15:04:05 MST 2006   (1 2 3 4 5 6 7: month day hour min sec year zone)
	t := time.Date(2026, time.March, 7, 9, 5, 3, 0, time.UTC)

	fmt.Println(t.Format("2006-01-02"))
	fmt.Println(t.Format("02 Jan 2006 15:04"))
	fmt.Println(t.Format("Monday, January 2 at 3:04 PM"))
	fmt.Println(t.Format(time.RFC3339))
	fmt.Println(t.Format(time.Kitchen), t.Format(time.DateTime))

	// Parse uses the same layouts.
	p, err := time.Parse("2006-01-02 15:04", "2026-12-31 23:59")
	fmt.Println(p, err)
	_, err = time.Parse("2006-01-02", "31/12/2026")
	fmt.Println(err)

	// Calendar arithmetic and comparison.
	next := t.AddDate(0, 1, 0) // one month later
	fmt.Println(next.Format(time.DateOnly), next.Sub(t), next.After(t))
	fmt.Println(t.Year(), t.Month(), t.Day(), t.Weekday(), t.YearDay())
	fmt.Println(t.Unix(), time.Unix(0, 0).UTC())

	// Zones: the same instant shown in another location.
	est := time.FixedZone("EST", -5*60*60)
	fmt.Println(t.In(est).Format(time.RFC1123), t.Equal(t.In(est)))
}
