import langchain

from run_llm import initial_synopsis, converse

from langchain_community.document_loaders import DirectoryLoader, TextLoader
from langchain.text_splitter import RecursiveCharacterTextSplitter
from langchain.schema import Document
from langchain.text_splitter import Language

import os

import argparse 
import json


def get_splitter_for_language(file_extension):
    language_map = {
        '.py': Language.PYTHON,
        '.js': Language.JS,
        '.ts': Language.TS,
        '.java': Language.JAVA,
        '.cpp': Language.CPP,
        '.c': Language.C,
        '.cs': Language.CSHARP,
        '.go': Language.GO,
        '.rb': Language.RUBY,
        '.php': Language.PHP,
        '.swift': Language.SWIFT,
        '.kt': Language.KOTLIN,
        '.rs': Language.RUST,
    } # this language splitter might be the reason it isn't working.
    
    if file_extension in language_map:
        return RecursiveCharacterTextSplitter.from_language(
            language=language_map[file_extension],
            chunk_size=1000,
            chunk_overlap=200
        )
    else:
        # Fallback for other file types
        return RecursiveCharacterTextSplitter(
            chunk_size=1000,
            chunk_overlap=200
        )


def writeout_chunks(record, uuid):
    with open(f"./working/{uuid}/temp.json", 'w') as f:
        json.dump(record, f, indent=4)
    
    
def load_documents(args):
    print(f"Attempting to chunk files from directory \'{args.workingdir}\'")
    loader = DirectoryLoader(
        args.workingdir,
        glob="**/*.[!o]",  # or *.js, *.java, etc.
        loader_cls=TextLoader,
        loader_kwargs={'encoding': 'utf8'}
    )
    documents = loader.load()
    
    chunks = []
    for doc in documents:
        try:
            file_extension = os.path.splitext(doc.metadata['source'])[1]
            splitter = get_splitter_for_language(file_extension)
            doc_chunks = splitter.split_documents([doc])
            chunks.extend(doc_chunks)
        except:
            print(f"Could not parse file {doc.metadata['source']}")

    return chunks
    

def main():

    # ingestion is and all, but yeah 
    parser = argparse.ArgumentParser()
    parser.add_argument("--workingdir", type = str) # TODO required
    
    args = parser.parse_args()
    chunks = load_documents(args)

    # print(chunks)
    print(args.workingdir)

    data = []
    for id in range(len(chunks)):
        filename = chunks[id].metadata['source']
        data.append({"id": filename.split("/")[-1],"text":chunks[id].page_content}) # TODO metadata is weird in its own way, how to handle?

    uuid = args.workingdir.replace("./working/",'')
    writeout_chunks(data, uuid)
    convo, output = initial_synopsis(args)
    with open(f"./working/{uuid}/convo.json", 'w') as f:
        json.dump(convo, f)
    print(output)


if __name__ == "__main__":
    main()