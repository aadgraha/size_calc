package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/aquasecurity/table"
	"github.com/fatih/color"
)

type Trade struct {
	Pair           string
	AccountBalance float64
	RiskPercent    float64
	Direction      string
	Spread         float64
	Entry          float64
	Pivot          float64
	RiskValue      float64
	EntryPrice     float64
	StopLoss       float64
	TP1            float64
	TP2            float64
	TP3            float64
	//for calculation not for output
	Magnitude int
}

func main() {
	records, err := readCSV("input.csv")
	if err != nil {
		log.Fatalf("Could not read input CSV: %s", err)
	}
	var trades []Trade

	for i, record := range records {
		if i == 0 {
			continue
		}
		trade := processRecord(record)
		trades = append(trades, trade)
	}
	printTable(trades)
}

func readCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comment = '#'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func processRecord(record []string) Trade {
	var trade Trade
	trade.Pair = record[0]
	trade.AccountBalance, _ = strconv.ParseFloat(record[1], 64)
	trade.RiskPercent, _ = strconv.ParseFloat(record[2], 64)
	trade.Spread, _ = strconv.ParseFloat(record[3], 64)
	trade.Entry, _ = strconv.ParseFloat(record[4], 64)
	trade.Pivot, _ = strconv.ParseFloat(record[5], 64)
	trade.Magnitude = int(math.Round(math.Pow(10, float64(magnitudeCalculation(record[4])))))
	magnitude := float64(trade.Magnitude)
	trade.RiskValue = (trade.RiskPercent / 100) * trade.AccountBalance

	// Determine direction
	if trade.Entry > trade.Pivot {
		trade.Direction = "BUY"
	} else {
		trade.Direction = "SELL"
	}

	if trade.Direction == "BUY" {
		trade.EntryPrice = trade.Entry
		trade.StopLoss = trade.Pivot - (trade.Spread / magnitude) - (50 / magnitude)
	} else {
		trade.EntryPrice = trade.Entry
		trade.StopLoss = trade.Pivot + (trade.Spread / magnitude) + (50 / magnitude)
	}
	trade.TP1 = trade.EntryPrice + (trade.EntryPrice - trade.StopLoss)
	trade.TP2 = trade.EntryPrice + 2*(trade.EntryPrice-trade.StopLoss)
	trade.TP3 = trade.EntryPrice + 3*(trade.EntryPrice-trade.StopLoss)
	return trade
}

func printTable(trades []Trade) {
	t := table.New(os.Stdout)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	t.SetHeaders(
		"Pair", "Direction", "AccBalance", "Risk%", "Spread", "Entry", "Pivot", "RiskValue",
		"EntryPrice", "SL", "TP1", "TP2", "TP3")

	for _, trade := range trades {
		t.AddRow(
			trade.Pair,
			trade.Direction,
			fmt.Sprintf("%.2f", trade.AccountBalance),
			fmt.Sprintf("%.2f", trade.RiskPercent),
			fmt.Sprintf("%.2f", trade.Spread),
			fmt.Sprintf("%.5f", trade.Entry),
			fmt.Sprintf("%.5f", trade.Pivot),
			yellow(fmt.Sprintf("%.5f", trade.RiskValue)),
			blue(fmt.Sprintf("%.5f", trade.EntryPrice)),
			red(fmt.Sprintf("%.5f", trade.StopLoss)),
			green(fmt.Sprintf("%.5f", trade.TP1)),
			green(fmt.Sprintf("%.5f", trade.TP2)),
			green(fmt.Sprintf("%.5f", trade.TP3)),
		)
	}
	t.Render()
}

func magnitudeCalculation(num string) int {
	parts := strings.Split(num, ".")
	if len(parts) < 2 {
		return 0
	}
	return len(parts[1])
}
