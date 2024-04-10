#!/bin/bash
set -e

go build -o cccat.exe cccat.go

echo Testing stdin
echo -e "abc\ndef\nghi" | cat > cat.out
echo -e "abc\ndef\nghi" | ./cccat.exe > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing concatenating multiple files
cat test.txt test2.txt > cat.out
./cccat.exe test.txt test2.txt > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -n option
cat -n test.sh > cat.out
./cccat.exe -n test.sh > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -b option
cat -b test.sh > cat.out
./cccat.exe -b test.sh > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -v option
cat -v test.txt > cat.out
./cccat.exe -v test.txt > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -E option
cat -E test.txt > cat.out
./cccat.exe -E test.txt > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -T option
cat -T test.txt > cat.out
./cccat.exe -T test.txt > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing -A option
cat -A test.txt > cat.out
./cccat.exe -A test.txt > cccat.out
diff --ignore-all-space cat.out cccat.out

echo Testing on binary
cat cccat.exe > cat.out
./cccat.exe cccat.exe > cccat.out
diff --ignore-all-space cat.out cccat.out

echo All tests passed
rm cccat.exe cat.out cccat.out