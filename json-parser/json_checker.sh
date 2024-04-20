#!/bin/bash
go build -o json-parser.exe json-parser.go
if [ $? -ne 0 ] ; then
    exit
fi

# Tests: https://www.json.org/JSON_checker/

for test in $(find json_checker -name "*.json") ; do
    name=$(basename $test)
    echo === Testing $test ===
    ./json-parser.exe --payload-only --max-depth 20 $test
    RESULT=$?
    if [ $RESULT -eq 2 ] ; then
        exit
    fi
    if [[ ( $RESULT -eq 0 && $name =~ 'fail' ) || ( $RESULT -eq 1 && $name =~ 'pass' ) ]] ; then
        echo Test failed:
        cat $test
        exit
    fi
done

echo All json_cheker tests passed