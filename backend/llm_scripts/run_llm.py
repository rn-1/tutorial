# requires accelerate to be installed.
import torch
import os
import json
from dotenv import load_dotenv # i think i might use llama or something here
from transformers import AutoTokenizer, AutoModelForCausalLM
import argparse
import transformers

# Functions for running the tokenizers and model.

def parse_chunks(workingdir): # do we need this?
    
    with open(workingdir + "/temp.json", "r") as file:
        chunks = json.load(file)
        assembled = "Here is some relevant information from the codebase:\n\n"
        for chunk in chunks:
            assembled += f"From file {chunk["id"]}:\n{chunk["text"]}\n"
            
        return assembled
    # we will need to change this later maybe...

def parse_messages(workingdir): # lol?
    with open(workingdir + "/convo.json", "r") as file:
        convo = json.load(file)
        return convo

def converse(args):
    messages = [
        {"role":"system","content":"You are an AI assistant that is an expert in reading and understanding code. Your task is to answer questions, asked by the user, about a specified code base based on the content of its files. Give a short synopsis of the following: \n\t1: What is this code meant to do? \n\t2:How does it accomplish this? Refer to specific sections of code or practices used in the codebase \n\t3: What are the basic things one must know to be able to use the codebase effectively in their own projects?\n Please refer to the code that is given as many times as is needed, and provide as much detail as you feel is needed. You may further need to ask the user what specific functionality they want out of the codebase, and adjust your later responses accordingly."}
    ]
    messages.append(parse_messages(args.workingdir))
    chunks = parse_chunks(args.workingdir) # ok cool.
    
    output = call_llm(messages, chunks)
    messages += {"role":"assistant","content":output}
    
    return output, messages

def initial_synopsis(args):
    messages = [
        {"role":"system","content":"You are an AI assistant that is an expert in reading and understanding code. Your task is to answer questions, asked by the user, about a specified code base based on the content of its files. Give a short synopsis of the following: \n\t1: What is this code meant to do? \n\t2:How does it accomplish this? Refer to specific sections of code or practices used in the codebase \n\t3: What are the basic things one must know to be able to use the codebase effectively in their own projects?\n Please refer to the code that is given as many times as is needed, and provide as much detail as you feel is needed. You may further need to ask the user what specific functionality they want out of the codebase, and adjust your later responses accordingly."},
        {"role":"user","content":"How can I get started with using this code repository for myself?"}
    ]
    retrieved = parse_chunks(args.workingdir)

    output = call_llm(messages, retrieved)

    messages += {"role":"assistant","content":output}
    return messages, output

def call_llm(convo, chunks):
    model_name = "Qwen/Qwen3-4B-Instruct-2507"

    convo.append({"role":"system","content":chunks})

    tokenizer = AutoTokenizer.from_pretrained(model_name)
    model = AutoModelForCausalLM.from_pretrained(
        model_name,
        torch_dtype=torch.float32,
        device_map="cpu"
    )

    text = tokenizer.apply_chat_template(
        convo,
        tokenize=False,
        add_generation_prompt=True,
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

    return content

def main(args):
    # run the llm
    if args.mode == "initial":
        output = initial_synopsis(args)
    else:
        output = converse(args)

    return output 

if __name__ == "__main__":

    parser = argparse.ArgumentParser()
    parser.add_argument("--mode", type = str, required=True)
    parser.add_argument("--workingdir", type = str, required = True)

    args = parser.parse_args()
    main(args)