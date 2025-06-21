#!/bin/bash
# This script is used to run tsetup and run the run_llm.py script


source ./Scripts/activate # this will need to be better structured in the future. dockerize?
python3 query_repo.py --working_dir "$@" # we should never pass in more than one argument.

# by knowing working dir we can find the chat history and relevant vectors.