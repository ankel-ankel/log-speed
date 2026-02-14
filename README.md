# How to use this mess

> UI inspired by [keilerkonzept/sliding-topk-tui-demo](https://github.com/keilerkonzept/sliding-topk-tui-demo)
>
> Sample data from [Web Server Access Logs](https://www.kaggle.com/datasets/eliasdabbas/web-server-access-logs) on Kaggle

## Data

Create a `data` folder in the project root and place the access log file downloaded from Kaggle inside it:

```
project/
├── data/
│   └── access.log
├── program/
│   └── ...
└── ...
```

## Run (Recommended)

```powershell
go build -mod=vendor -o .\logspeed.exe .\program; .\logspeed.exe -in .\data\access.log -access-log -replay -replay-speed 500 -replay-max-sleep 10ms -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 15 -items-fps 2 -item-counts-fps 2 -search -full-refresh 3s -partial-size 30 -stats -stats-window 256 -alt-screen=false
```

## Run Fast

```powershell
go build -mod=vendor -o .\logspeed.exe .\program; .\logspeed.exe -in .\data\access.log -access-log -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 5 -items-fps 1 -item-counts-fps 0 -search=false -full-refresh 0 -stats -stats-window 256 -alt-screen=false
```

## Metrics

- `records`: total ingested records.
- `ingest rate`: recent ingest throughput (records/sec).
- `pipeline lag p95`: p95 delay from the latest ingest to the next ranking update.
- `data freshness lag`: delay from now to the latest ingested record.
- `top-1`: current #1 item and count.
- `track`: current tracked item when `t` is enabled (`off` if tracking is disabled).

## Keys

- `p`: pause/resume.
- `t` or `Space`: track selected item.
- `s`: toggle linear/log scale.
- `q` or `Ctrl+C`: quit.
