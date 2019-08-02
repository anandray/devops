package main

import (
	"fmt"
	"sort"
	"strconv"
)

var numbers []Stat

type Stat struct {
	Tm float64
}
type Stats []Stat

func (a Stats) Len() int           { return len(a) }
func (a Stats) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a Stats) Less(i, j int) bool { return a[i].Tm < a[j].Tm }

func print_percentiles(numbers Stats) error {
	sort.Sort(numbers)
	l := len(numbers)
	printPercentileN(numbers, l, 50)
	printPercentileN(numbers, l, 66)
	printPercentileN(numbers, l, 75)
	printPercentileN(numbers, l, 80)
	printPercentileN(numbers, l, 90)
	printPercentileN(numbers, l, 95)
	printPercentileN(numbers, l, 98)
	printPercentileN(numbers, l, 99)
	printPercentileN(numbers, l, 100)
	return nil
}

func percentileN(numbers Stats, l, n int) float64 {
	i := l*n/100 - 1
	return numbers[i].Tm
}

func printPercentileN(numbers Stats, l, n int) {
	fmt.Printf("%d%%:\t%s\n", n, strconv.FormatFloat(percentileN(numbers, l, n), 'g', 16, 64))
}
