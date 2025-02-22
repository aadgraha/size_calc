package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aquasecurity/table"
	"github.com/fatih/color"
	"github.com/google/uuid"
)

type Trade struct {
	ID             string
	DateTime       string
	Pair           string
	Direction      string
	AccountBalance float64
	RiskPercent    float64
	Spread         float64
	Entry          float64
	Pivot          float64
	EntryPrice     float64
	StopLoss       float64
	LotSize        float64
	TP1            float64
	TP2            float64
	TP1Profit      float64
	TP2Profit      float64
	RiskAmount     float64
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
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func processRecord(record []string) Trade {
	var trade Trade
	trade.ID = uuid.New().String()
	trade.DateTime = time.Now().Format("2006-01-02 15:04:05")
	trade.Pair = record[0]
	trade.AccountBalance, _ = strconv.ParseFloat(record[1], 64)
	trade.RiskPercent, _ = strconv.ParseFloat(record[2], 64)
	trade.Spread, _ = strconv.ParseFloat(record[3], 64)
	trade.Entry, _ = strconv.ParseFloat(record[4], 64)
	trade.Pivot, _ = strconv.ParseFloat(record[5], 64)

	// Determine direction
	if trade.Entry > trade.Pivot {
		trade.Direction = "BUY"
	} else {
		trade.Direction = "SELL"
	}

	pointCount := float64(magnitudeCalculation(trade.Entry))
	magnitude := math.Pow(10, pointCount)
	if trade.Direction == "BUY" {
		trade.EntryPrice = trade.Entry + (trade.Spread / magnitude)
		trade.StopLoss = trade.Pivot - (trade.Spread / magnitude) - (5 / magnitude)
	} else {
		trade.EntryPrice = trade.Entry - (trade.Spread / magnitude)
		trade.StopLoss = trade.Pivot + (trade.Spread / magnitude) + (5 / magnitude)
	}

	pointDistance := (math.Abs(trade.EntryPrice - trade.StopLoss))
	trade.RiskAmount = trade.RiskPercent / 100 * trade.AccountBalance
	if strings.HasPrefix(trade.Pair, "usd") {
		trade.LotSize = trade.RiskAmount / (pointDistance * magnitude / pipRatio(trade.Pair))
	} else {
		trade.LotSize = trade.RiskAmount / (pointDistance / trade.StopLoss * magnitude)
	}
	trade.LotSize = math.Floor(trade.LotSize*100) / 100
	trade.RiskAmount = trade.LotSize * pointDistance * magnitude / pipRatio(trade.Pair)
	trade.TP1 = trade.EntryPrice + (trade.EntryPrice - trade.StopLoss)
	trade.TP2 = trade.EntryPrice + 3*(trade.EntryPrice-trade.StopLoss)
	trade.TP1Profit = trade.LotSize * (math.Abs(trade.TP1-trade.EntryPrice) * magnitude / pipRatio(trade.Pair))
	trade.TP2Profit = trade.LotSize * (math.Abs(trade.TP2-trade.EntryPrice) * magnitude / pipRatio(trade.Pair))
	return trade
}

func printTable(trades []Trade) {
	t := table.New(os.Stdout)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	t.SetHeaders("ID", "DateTime", "Pair", "Direction", "AccBalance", "Risk%", "Spread", "Entry", "Pivot", "EntryPrice", "SL", "LotSize", "TP1", "TP2", "$TP1", "$TP2", "$SL")

	for _, trade := range trades {
		t.AddRow(
			trade.ID[:4], trade.DateTime, trade.Pair, trade.Direction,
			fmt.Sprintf("%.2f", trade.AccountBalance), fmt.Sprintf("%.2f", trade.RiskPercent),
			fmt.Sprintf("%.2f", trade.Spread), fmt.Sprintf("%.5f", trade.Entry),
			fmt.Sprintf("%.5f", trade.Pivot), blue(fmt.Sprintf("%.5f", trade.EntryPrice)),
			red(fmt.Sprintf("%.5f", trade.StopLoss)), yellow(fmt.Sprintf("%.2f", trade.LotSize)),
			green(fmt.Sprintf("%.5f", trade.TP1)), green(fmt.Sprintf("%.5f", trade.TP2)),
			fmt.Sprintf("%.2f", trade.TP1Profit), fmt.Sprintf("%.2f", trade.TP2Profit),
			fmt.Sprintf("%.2f", trade.RiskAmount*-1),
		)
	}
	t.Render()
}

func magnitudeCalculation(num float64) int {
	str := strconv.FormatFloat(num, 'f', -1, 64)
	parts := strings.Split(str, ".")
	if len(parts) < 2 {
		return 0
	}
	return len(parts[1])
}

func pipRatio(pair string) float64 {
	switch pair {
	case "usd_jpy":
		return 1.49
	}
	return 1
}

//todo make rest api out of this, with svelte front end