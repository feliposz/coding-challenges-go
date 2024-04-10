#!/bin/bash
set -e

go build -o ccwc.exe ccwc.go

echo Testing -c option
wc -c test.txt > wc.out
./ccwc.exe -c test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing -l option
wc -l test.txt > wc.out
./ccwc.exe -l test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing -w option
wc -w test.txt > wc.out
./ccwc.exe -w test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing -m option
wc -m test.txt > wc.out
./ccwc.exe -m test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing -L option
wc -L test.txt > wc.out
./ccwc.exe -L test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing default options
wc test.txt > wc.out
./ccwc.exe test.txt > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing multiple
wc test.txt test.sh ccwc.go > wc.out
./ccwc.exe test.txt test.sh ccwc.go > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing stdin
cat test.txt | wc > wc.out
cat test.txt | ./ccwc.exe > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo Testing -c on binary
wc -c ccwc.exe > wc.out
./ccwc.exe -c ccwc.exe > ccwc.out
diff --ignore-all-space wc.out ccwc.out

echo All tests passed
rm ccwc.exe wc.out ccwc.out