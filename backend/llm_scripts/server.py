from fastapi import FastAPI
from pydantic import BaseModel
from typing import List, Dict, Any

import torch
import os
import json
from dotenv import load_dotenv # i think i might use llama or something here
from transformers import AutoTokenizer, AutoModelForCausalLM, AutoConfig
import argparse
import transformers
import uvicorn

# TODO



class Prompt(BaseModel):
    conversation: List[dict]
    chunks: str


def initialize_model():
    global tokenizer, model


    # decide offloading stuff for testing
    # device = torch.device('cuda') if torch.cuda.is_available() else torch.device('cpu')
    
    model_name = "google/gemma-3-270m-it"
    # config = AutoConfig.from_pretrained(model_name)
    # print(config.torch_dtype)
    
    model = AutoModelForCausalLM.from_pretrained(
            model_name,
            device_map=None,
            torch_dtype=torch.float16,
    )
    tokenizer = AutoTokenizer.from_pretrained(model_name)

def call_llm(convo, chunks):

    converse = convo + [{"role": "system", "content": chunks}]

    text = tokenizer.apply_chat_template(
        converse,
        tokenize=False,    
    )
    model_inputs = tokenizer([text], return_tensors="pt").to(model.device)

    # conduct text completion
    generated_ids = model.generate(
        **model_inputs,
        max_new_tokens=32768
    )
    output_ids = generated_ids[0][len(model_inputs.input_ids[0]):].tolist() 

    try:
    # rindex finding 151668 (</think>)
        index = len(output_ids) - output_ids[::-1].index(151668)
    except ValueError:
        index = 0

    content = tokenizer.decode(output_ids[index:], skip_special_tokens=True).strip("\n")

    # content = "debug_output" # TODO undo this

    return content

app = FastAPI()

initialize_model()

print("Done initializing model.")

@app.post("/generate")
def generate(prompt: Prompt) -> Dict[str, Any]:

    # parse the conversation and chunks
    print("generating...")
    messages = prompt.conversation
    chunks = prompt.chunks
    output = call_llm(messages, chunks)
    return {"status": "success", "response": output}