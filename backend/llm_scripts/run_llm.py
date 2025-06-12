import LangChain
# requires accelerate to be installed.
import torch
import os
import json
from dotenv import load_dotenv # i think i might use llama or something here
from transformers import T5Tokenizer, T5ForConditionalGeneration
import argparse




def main():

    parser = argparse.ArgumentParser()

    parser.add_argument("--vectorfile", required = True, tyep = str)

    tokenizer = T5Tokenizer.from_pretrained("google/flan-t5-xl")
    model = T5ForConditionalGeneration.from_pretrained("google/flan-t5-xl", device_map="auto")

    # what format can I use for the vector file? maybe I load in as a json and parse back.

    with open("vectors.json", 'r') as f:
        json = json


if __name__ == "__main__":
    main()