package stdbank

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/go-msvc/msf/logger"
	"github.com/jansemmelink/money/bank"
)

var log = logger.New("money").New("stdbank")

func LoadStatement(fn string) (stmt bank.IStatement, err error) {
	csvFile, err := os.Open(fn)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s: %v", fn, err)
	}
	defer csvFile.Close()

	err = nil
	stmt = bank.NewStatement("Standard Bank")
	lineNr := 0
	defer func() {
		if err != nil {
			err = fmt.Errorf("line(%d): %v", lineNr, err)
		}
	}()

	csvReader := csv.NewReader(csvFile)

	total, _ := bank.NewAmount(0)
	for {
		lineNr++
		var record []string
		record, err = csvReader.Read()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			err = fmt.Errorf("invalid CSV: %v", err)
			return
		}

		switch lineNr {
		case 1:
			//Line(     1): [0 2645 BRANCH 0  CENTURION 0 0]
			stmt = stmt.WithBranchName(record[5])
			stmt = stmt.WithBranchCode(record[1])
			continue

		case 2:
			//Line(     2): [ 12319791 ACC-NO 0   0 0]
			stmt = stmt.WithAccountNumber(record[1])
			continue

		} //switch(header lines)

		var amount bank.Amount
		amount, err = bank.NewAmount(record[3])
		if err != nil {
			err = fmt.Errorf("col[4]=%s is not valid amount: %v", record[3], err)
			return
		}

		//open/closing balances:
		//Line(     2): [ 0 OPEN 44608.60 OPEN BALANCE  0 0]
		//...
		//Line(   295): [ 0 CLOSE -45775.73 CLOSE BALANCE  0 0]
		if record[0] == "" && record[1] == "0" {
			if record[2] == "OPEN" {
				stmt = stmt.WithOpeningBalance(amount)
				continue
			}
			if record[2] == "CLOSE" {
				stmt = stmt.WithClosingBalance(amount)
				continue
			}
		}

		//Transaction line:
		//Line(     3): [HIST 20200928  -211.22 TJEKKAART-AANKOOP Spar Midstrea 5222*7143 23 SEP 6076 0]
		if record[0] == "HIST" {
			//date: 20200928 -> 2020-09-28
			var date time.Time
			date, err = time.ParseInLocation("20060102", record[1], time.Now().Location())
			if err != nil {
				err = fmt.Errorf("invalid date=\"%s\" not CCYYMMDD", record[1])
				return
			}
			stmt = stmt.WithTransaction(bank.NewTransaction(date, amount, record[4], record[5], record[6]))
			total = total.Add(amount)
			log.Debugf("Line(%6d): %v total=%v", lineNr, record, total)
			continue
		}
	}

	err = stmt.Validate()
	return
}

/*
Line(     0): [0 2645 BRANCH 0  CENTURION 0 0]
Line(     1): [ 12319791 ACC-NO 0   0 0]
Line(     2): [ 0 OPEN 44608.60 OPEN BALANCE  0 0]
Line(     3): [HIST 20200928  -211.22 TJEKKAART-AANKOOP Spar Midstrea 5222*7143 23 SEP 6076 0]
Line(     4): [HIST 20200928  -932 TJEKKAART-AANKOOP C*SASOL MIDRI 5222*7143 24 SEP 6076 0]
Line(     5): [HIST 20200928  -541.87 TJEKKAART-AANKOOP C*THE APRON B 5222*7143 24 SEP 6076 0]
Line(     6): [HIST 20200928  -419.98 TJEKKAART-AANKOOP Spar Midstrea 5222*7143 24 SEP 6076 0]
Line(     7): [HIST 20200928  -580 TJEKKAART-AANKOOP C*PETENA PANC 5222*7143 25 SEP 6076 0]
Line(     8): [HIST 20200928  -400 IB-BETALING NA SANRAL CLEARING HOU 466557143 377 0]
Line(     9): [HIST 20200928  -899.12 VERSEKERINGSPREMIE OUTSURANCE OT11259326   94752Q 6021 0]
...
Line(   287): [HIST 20210317  -500 SELFOON KITSONTTR KONTANT NA 0723082168 10H44 088489836 760 0]
Line(   288): [HIST 20210317 ## -8 FOOI - KITSGELD 0723082168 10H44 088489836 01644 0]
Line(   289): [HIST 20210318  -500 TJEKKAART-AANKOOP SASOL MIDRIDG 5222*7143 15 MAR 6076 0]
Line(   290): [HIST 20210318  -1024.12 TJEKKAART-AANKOOP SASOL MIDRIDG 5222*7143 15 MAR 6076 0]
Line(   291): [HIST 20210318  -84.99 TJEKKAART-AANKOOP Spar Midstrea 5222*7143 15 MAR 6076 0]
Line(   292): [HIST 20210318  259.37 KREDIETOORPLASING CORNUEX 6088 0]
Line(   293): [HIST 20210318  -3924.13 ELEKTR. OORPL. - KREDIETKAART SB OUTOBETAAL 5221266468371116 6055 0]
Line(   294): [HIST 20210319  -51.98 TJEKKAART-AANKOOP Spar Midstrea 5222*7143 16 MAR 6076 0]
Line(   295): [ 0 CLOSE -45775.73 CLOSE BALANCE  0 0]

*/
