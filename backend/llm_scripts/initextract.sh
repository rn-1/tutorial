#!/bin/bash
# This script is used to run the run_llm.py script


source ./Scripts/activate # this will need to be better structured in the future. dockerize?
python3 run_llm.py --working_dir "$@" # we should never pass in more than one argument.