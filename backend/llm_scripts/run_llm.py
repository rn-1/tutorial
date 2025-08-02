# requires accelerate to be installed.
import torch
import os
import json
from dotenv import load_dotenv # i think i might use llama or something here
from transformers import T5Tokenizer, T5ForConditionalGeneration
import argparse


# Functions for running the tokenizers and model.
#fuck man idk

def tokenize_chunk(text):
    tokenizer = T5Tokenizer.from_pretrained("google/flan-t5-base") # TODO conver to GPU
    embeds = tokenizer(text, return_tensors="pt")
    # Any other cleanup we need to do?

    return embeds