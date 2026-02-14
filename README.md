# How to use this mess

> UI inspired by [keilerkonzept/sliding-topk-tui-demo](https://github.com/keilerkonzept/sliding-topk-tui-demo)
>
> Sample data from [Web Server Access Logs](https://www.kaggle.com/datasets/eliasdabbas/web-server-access-logs) on Kaggle

## 1. Prepare data

Create a `data` folder in the project root and place the access log file downloaded from Kaggle inside it:

```
project/
├── data/
│   └── access.log
├── program/
│   └── ...
└── ...
```

## 2. Build

```bash
go build -mod=vendor -o .\experiment.exe .\program
```

> If the folder already has `experiment.exe`, skip to step 3.

## 3. Benchmark (full speed)

```bash
.\experiment.exe -in .\data\access.log -access-log -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 5 -items-fps 1 -item-counts-fps 0 -search=false -full-refresh 0 -stats -stats-window 256 -alt-screen=false
```

## 4. Replay (slow motion, pausable)

```bash
.\experiment.exe -in .\data\access.log -access-log -replay -replay-speed 500 -replay-max-sleep 10ms -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 15 -items-fps 2 -item-counts-fps 2 -search -full-refresh 3s -partial-size 30 -stats -stats-window 256 -alt-screen=false
```

## Keyboard

| Key | Action |
|-----|--------|
| `p` | Pause / Resume |
| `t` / `Space` | Track selected item |
| `s` | Toggle log / linear scale |
| `q` / `Ctrl+C` | Quit |
