#!/bin/bash
set -e
go build -o huffman.exe huffman.go
if [ $? -ne 0 ] ; then
    exit
fi

for test in LesMiserables.txt ../readme.md huffman.go huffman.exe; do
    echo === Testing $test ===
    echo Compressing
    ./huffman.exe -c $test compressed.cchf
    echo Decompressing
    ./huffman.exe -d compressed.cchf decompressed.out
    echo Comparing
    diff $test decompressed.out
    ORIGINAL_SIZE=$(wc -c $test | cut -f1 -d ' ')
    COMPRESSED_SIZE=$(wc -c compressed.cchf | cut -f1 -d ' ')
    DECOMPRESSED_SIZE=$(wc -c decompressed.out | cut -f1 -d ' ')
    RATIO=$(( 100 * $COMPRESSED_SIZE / $ORIGINAL_SIZE ))
    echo "Original: $ORIGINAL_SIZE, Compressed: $COMPRESSED_SIZE, Ratio: $RATIO%"
done


rm compressed.cchf decompressed.out

echo All tests passed!