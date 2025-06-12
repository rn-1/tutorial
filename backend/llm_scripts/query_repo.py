import langchain

from langchain.document_loaders import DirectoryLoader, TextLoader
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
    }
    
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
    loader = DirectoryLoader(args.workingdir)
    loader = DirectoryLoader(
        "path/to/code/directory",
        glob="**/*.py",  # or *.js, *.java, etc.
        loader_cls=TextLoader,
        loader_kwargs={'encoding': 'utf8'}
    )
    documents = loader.load()
    
    chunks = []
    for doc in documents:
        file_extension = os.path.splitext(doc.metadata['source'])[1]
        splitter = get_splitter_for_language(file_extension)
        doc_chunks = splitter.split_documents([doc])
        chunks.extend(doc_chunks)

    return chunks
    

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("workingdir", type = str, required = True)
    
    args = parser.parse_args()
    chunks = load_documents(args)

    print(chunks)
    data = []
    for id in range(len(chunks)):
        filename = chunks[id].metadata['source']
        data.append({"id": filename, "text":chunks[id].page_content}) # TODO metadata is weird in its own way, how to handle?

    writeout_chunks(data, args.workingdir)
    print("[INGEST] ingested documents for session", args.workingdir)