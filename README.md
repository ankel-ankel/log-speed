Instruction on how to use this mess

1. First, paste this into the IDE terminal (If the folder already has it, then move to the second part)
go build -mod=vendor -o .\experiment.exe .\program

2. Second, if you want to focus on the benchmark, then use this:
.\experiment.exe -in .\data\access.log -access-log -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 60 -items-fps 1 -item-counts-fps 2 -search -full-refresh 2s -partial-size 30 -stats -stats-window 256 -alt-screen=false

3. Third, if you want to make it run slowly, or pause the program, then use this:
.\experiment.exe -in .\data\access.log -access-log -replay -replay-speed 1000 -replay-max-sleep 5ms -k 20 -tick 1m -window 1h -json-timestamp-layout "02/Jan/2006:15:04:05 -0700" -view-split 30 -plot-fps 60 -items-fps 1 -item-counts-fps 2 -search -full-refresh 2s -partial-size 30 -stats -stats-window 256 -alt-screen=false
