package output

import (
	"fmt"
	"strings"
)

const maxColWidth = 50

func PrintTable(columns []string, rows [][]string) {
	if len(columns) == 0 {
		fmt.Println("no rows")
		return
	}

	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = len(col)
	}

	for _, row := range rows {
		for i, val := range row {
			if len(val) > maxColWidth {
				val = val[:maxColWidth]
			}
			if len(val) > widths[i] {
				widths[i] = len(val)
			}
		}
	}

	printSeparator(widths)
	printRow(columns, widths)
	printSeparator(widths)

	for _, row := range rows {
		printRow(row, widths)
	}

	printSeparator(widths)
}

func printSeparator(widths []int) {
	fmt.Print("+")
	for _, w := range widths {
		fmt.Print(strings.Repeat("-", w+2))
		fmt.Print("+")
	}
	fmt.Println()
}

func printRow(row []string, widths []int) {
	fmt.Print("|")
	for i, col := range row {
		if len(col) > maxColWidth {
			col = col[:maxColWidth]
		}
		fmt.Printf(" %-*s |", widths[i], col)
	}
	fmt.Println()
}
