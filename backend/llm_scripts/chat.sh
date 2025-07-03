#!/bin/bash
# This script is used to run tsetup and run the run_llm.py script


source ../llm_scripts/bin/activate # this will need to be better structured in the future. dockerize?
python3 ../llm_scripts/query_repo.py --working_dir "$1" # we should never pass in more than one argument.

# by knowing working dir we can find the chat history and relevant vectors.