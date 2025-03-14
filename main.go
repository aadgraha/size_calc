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
	TradeValue     float64
	//for calculation not for output
	Magnitude   int
	TP1Inserted bool
	TP2Inserted bool
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
	trade.TP1Inserted = true
	trade.TP2Inserted = true
	isTP1, err := strconv.ParseBool(record[6])
	if err != nil {
		trade.TP1Inserted = false
	}
	isTP2, err := strconv.ParseBool(record[7])
	if err != nil {
		trade.TP2Inserted = false
	}
	trade.ID = uuid.New().String()
	trade.DateTime = time.Now().Format("2006-01-02 15:04:05")
	trade.Pair = record[0]
	trade.AccountBalance, _ = strconv.ParseFloat(record[1], 64)
	trade.RiskPercent, _ = strconv.ParseFloat(record[2], 64)
	trade.Spread, _ = strconv.ParseFloat(record[3], 64)
	trade.Entry, _ = strconv.ParseFloat(record[4], 64)
	trade.Pivot, _ = strconv.ParseFloat(record[5], 64)
	trade.Magnitude = int(math.Round(math.Pow(10, float64(magnitudeCalculation(record[4])))))
	magnitude := float64(trade.Magnitude)

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

	pointDistance := (math.Abs(trade.EntryPrice - trade.StopLoss))
	trade.RiskAmount = trade.RiskPercent / 100 * trade.AccountBalance
	trade.LotSize = trade.RiskAmount / (pointDistance * magnitude / pipRatio(trade.Pair))
	trade.LotSize = math.Floor(trade.LotSize*100) / 100
	trade.RiskAmount = trade.LotSize * pointDistance * magnitude / pipRatio(trade.Pair)
	trade.TP1 = trade.EntryPrice + (trade.EntryPrice - trade.StopLoss)
	trade.TP2 = trade.EntryPrice + 3*(trade.EntryPrice-trade.StopLoss)
	trade.TP1Profit = trade.LotSize * (math.Abs(trade.TP1-trade.EntryPrice) * magnitude / pipRatio(trade.Pair))
	trade.TP2Profit = trade.LotSize * (math.Abs(trade.TP2-trade.EntryPrice) * magnitude / pipRatio(trade.Pair))
	trade.TradeValue = tradeValue(&trade, isTP1, isTP2)
	return trade
}

func printTable(trades []Trade) {
	t := table.New(os.Stdout)
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	blue := color.New(color.FgHiBlue).SprintFunc()
	yellow := color.New(color.FgHiYellow).SprintFunc()
	t.SetHeaders(
		"ID", "DateTime", "Pair", "Direction", "AccBalance", "Risk%", "Spread", "Entry", "Pivot",
		"EntryPrice", "SL", "LotSize", "TP1", "TP2", "$TP1", "$TP2", "$SL", "$Value")

	for _, trade := range trades {
		t.AddRow(
			trade.ID[:4], trade.DateTime, trade.Pair, trade.Direction,
			fmt.Sprintf("%.2f", trade.AccountBalance), fmt.Sprintf("%.2f", trade.RiskPercent),
			fmt.Sprintf("%.2f", trade.Spread), fmt.Sprintf("%.5f", trade.Entry),
			fmt.Sprintf("%.5f", trade.Pivot), blue(fmt.Sprintf("%.5f", trade.EntryPrice)),
			red(fmt.Sprintf("%.5f", trade.StopLoss)), yellow(fmt.Sprintf("%.2f", trade.LotSize)),
			green(fmt.Sprintf("%.5f", trade.TP1)), green(fmt.Sprintf("%.5f", trade.TP2)),
			fmt.Sprintf("%.2f", trade.TP1Profit), fmt.Sprintf("%.2f", trade.TP2Profit),
			fmt.Sprintf("%.2f", trade.RiskAmount*-1), fmt.Sprintf("%.2f", trade.TradeValue),
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

func tradeValue(trade *Trade, isTP1, isTP2 bool) float64 {
	if !trade.TP1Inserted || !trade.TP2Inserted {
		return math.NaN()
	}
	if (trade.TP1Inserted && isTP1) && (trade.TP2Inserted && isTP2) {
		return trade.TP1Profit + trade.TP2Profit
	} else if (trade.TP1Inserted && isTP1) && (trade.TP2Inserted && !isTP2) {
		return trade.TP1Profit + (trade.RiskAmount * -1)
	} else if (trade.TP1Inserted && !isTP1) && (trade.TP2Inserted && isTP2) {
		return trade.RiskAmount + (trade.TP2Profit * -1)
	} else if (trade.TP1Inserted && !isTP1) && (trade.TP2Inserted && !isTP2) {
		return (trade.RiskAmount * -1) * 2
	} else {
		return 0
	}

}

func pipRatio(pair string) float64 {
	switch pair {
	case "aud_cad":
		return 1
	case "aud_chf":
		return 1
	case "aud_jpy":
		return 1.49
	case "aud_nzd":
		return 1
	case "aud_usd":
		return 1
	case "cad_chf":
		return 1
	case "cad_jpy":
		return 1
	case "chf_jpy":
		return 1.49
	case "eur_aud":
		return 1
	case "eur_cad":
		return 1
	case "eur_chf":
		return 0.9
	case "eur_gbp":
		return 0.79
	case "eur_jpy":
		return 1.49
	case "eur_nzd":
		return 1
	case "eur_usd":
		return 1
	case "gbp_aud":
		return 1
	case "gbp_cad":
		return 1
	case "gbp_chf":
		return 1
	case "gbp_jpy":
		return 1.49
	case "gbp_nzd":
		return 1
	case "gbp_usd":
		return 1
	case "nzd_cad":
		return 1
	case "nzd_chf":
		return 1
	case "nzd_jpy":
		return 1
	case "nzd_usd":
		return 1
	case "usd_cad":
		return 1
	case "usd_chf":
		return 0.9
	case "usd_jpy":
		return 1.49
	case "xau_usd":
		return 1
	}

	return 1
}
