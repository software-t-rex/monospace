#!/usr/bin/env bash
export NO_COLOR=1
docDir="$(dirname $0)/doc"
cd $docDir
go run doc.go