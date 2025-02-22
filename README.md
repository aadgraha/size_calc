# FX Size Calculator

A simple Forex position size calculator based on price action strategy.

## Strategy
- **Entry and Pivot**: The area of supply or demand.
- **TP1**: 1:1 Risk-Reward (RR) ratio.
- **TP2**: 1:3 RR ratio.

## Usage
### Input
Provide an `input.csv` file with the following format:

```
pair,account_balance,risk_percent,spread,entry,pivot
usd_jpy,1000,1,15,152.379,152.838
```

### Output
The output is displayed in a table format:

```
┌────────────┬───────────┬─────────┬───────────┬───────────┬──────┬───────┬───────┐
│ EntryPrice │    SL     │ LotSize │    TP1    │    TP2    │ $TP1 │ $TP2  │  $SL  │
├────────────┼───────────┼─────────┼───────────┼───────────┼──────┼───────┼───────┤
│ 152.36400  │ 152.85800 │ 0.03    │ 151.87000 │ 150.88200 │ 9.95 │ 29.84 │ -9.95 │
└────────────┴───────────┴─────────┴───────────┴───────────┴──────┴───────┴───────┘
```