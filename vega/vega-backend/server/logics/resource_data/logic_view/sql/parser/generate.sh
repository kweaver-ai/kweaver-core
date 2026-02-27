#!/bin/sh
# Copyright The kweaver.ai Authors.
# Licensed under the Apache License, Version 2.0.
# See the LICENSE file in the project root for details.

alias antlr4='java -Xmx500M -cp "./antlr-4.13.2-complete.jar:$CLASSPATH" org.antlr.v4.Tool'
antlr4 -Dlanguage=Go -no-visitor -package parsing -o ../parsing SqlBase.g4