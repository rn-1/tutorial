#!/bin/bash
# This script is used to run the run_llm.py script
# can i do this in a better way?

source ../llm_scripts/bin/activate  # this will need to be better structured in the future. dockerize?
python3 ../llm_scripts/query_repo.py --workingdir "$1"  # we should never pass in more than one argument.