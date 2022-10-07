package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/jansemmelink/money/stdbank"
)

func main() {
	filePtr := flag.String("f", "", "Filename to import")
	yesPtr := flag.Bool("y", false, "Import without prompt")
	verbosePtr := flag.Bool("v", false, "Verbose output")
	flag.Parse()
	if *filePtr == "" {
		panic("Missing -f <filename>")
	}

	stmt, err := stdbank.LoadStatement(*filePtr)
	if err != nil {
		panic(fmt.Errorf("load failed: %v", err))
	}

	if !(*yesPtr) {
		fmt.Printf("Statement Loaded Successfully\n")
		fmt.Printf("Filename: %s\n", os.Args[1])
		fmt.Printf("Bank Name: %s\n", stmt.BankName())
		fmt.Printf("Branch Name: %s\n", stmt.BranchCode())
		fmt.Printf("Branch Code: %s\n", stmt.BranchCode())
		fmt.Printf("Account Number: %s\n", stmt.AccountNumber())
		fmt.Printf("Open Date: %s\n", stmt.OpenDate())
		fmt.Printf("Open Balance: %s\n", stmt.OpenBalance())
		fmt.Printf("Close Date: %s\n", stmt.CloseDate())
		fmt.Printf("Close Balance: %s\n", stmt.CloseBalance())
	}
	if *verbosePtr {
		for _, tx := range stmt.Transactions() {
			fmt.Printf("%s,%s,%s,%s\n",
				tx.Date.Local().Format("2006-01-02"),
				tx.Details,
				tx.Type,
				tx.Amount)
		}
	}
	if !(*yesPtr) {
		fmt.Printf("Import (y/n)[n] ?")
		answer := ""
		fmt.Scanf("%s", &answer)
		answer = strings.ToUpper(answer)
		if len(answer) < 1 || answer[0] != 'Y' {
			fmt.Printf("Not imported.\n")
			os.Exit(0)
		}
	}

	id, err := stmt.ImportToDb()
	if err != nil {
		panic(fmt.Sprintf("failed to import: %+v", err))
	}
	fmt.Printf("Imported successfully as statement \"%s\"", id)
}
