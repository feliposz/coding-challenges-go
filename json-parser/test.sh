#!/bin/bash
go build -o json-parser.exe json-parser.go
if [ $? -ne 0 ] ; then
    exit
fi

for test in $(find tests -name "*.json") ; do
    name=$(basename $test)
    echo === Testing $test ===
    ./json-parser.exe < $test
    if [[ ( $? -eq 0 && $name =~ 'invalid' ) || ( $? -ne 0 && ! $name =~ 'invalid' ) ]] ; then
        echo Test failed:
        cat $test
        exit
    fi
done

echo All tests passed