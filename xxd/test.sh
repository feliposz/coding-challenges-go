#!/bin/bash
set -e

echo "File 1 contents" >> file1.txt
echo "File 2 contents" >> file2.txt
echo "File 3 contents" >> file3.txt
tar -cf files.tar file1.txt file2.txt file3.txt

go build -o ccxxd.exe ccxxd.go

echo Testing default options
xxd files.tar xxd.out
./ccxxd.exe files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing stdin/stdout
cat files.tar | xxd > xxd.out
cat files.tar | ./ccxxd.exe > ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing -e and -g options
xxd -e -g 4 files.tar xxd.out
./ccxxd.exe -e -g 4 files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing -l option
xxd -l 128 files.tar xxd.out
./ccxxd.exe -l 128 files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing -c option
xxd -c 8 files.tar xxd.out
./ccxxd.exe -c 8 files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing -s option
xxd -s 512 -l 16 files.tar xxd.out
./ccxxd.exe -s 512 -l 16 files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing -p option
xxd -p files.tar xxd.out
./ccxxd.exe -p files.tar ccxxd.out
diff --brief --ignore-all-space xxd.out ccxxd.out

echo Testing reverse
xxd files.tar temp.hex
xxd -r temp.hex xxd.out
./ccxxd.exe -r temp.hex ccxxd.out
diff --brief xxd.out ccxxd.out

echo All tests passed
rm file1.txt file2.txt file3.txt files.tar temp.hex ccxxd.exe xxd.out ccxxd.out