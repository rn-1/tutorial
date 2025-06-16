
import Textinput from "../widgets/textinput";
import ChatbotInterface from "../widgets/chatbot";
import Appbar from "../widgets/Appbar";
import '../output.css'
import React from "react";
import { useState } from "react";
import { useRef } from 'react';

// holy SHIT this needs a lot of help

const Main = () => {

    
    async function extractGithub(){
        try{
            let url = document.getElementById("github_url").value.trim();
            if(url === ""){
                console.log("empty repo link, ignore")
                return;
            }
            const response = await fetch("http://localhost:8080/initialExtract", {
                    method: "POST", 
                    mode:"cors", 
                    body: url,
                    headers: {
                        'Content-Type': 'application/json'
                    }
                }
            ).then(resp => resp.text());

            let resp = JSON.parse(response)
            let token = resp.token

            localStorage.setItem("sessionid", token) // yay!
            console.log("created session with uuid ",token)

            console.log(response);
        } catch(e){
            console.log(`failed to fetch: ${e}`);
        }
        // response is a json with status and a token
    }

    return (
        <div className="min-h-screen flex flex-col">
            <Appbar/>
            <div 
                id="body" 
                className="w-full flex flex-col items-center pt-20 px-4 flex-grow"
            >
                <Textinput initialVal="" _placeholder="github url to extract from" id="github_url" onsubmit={extractGithub}/>
                <div className="w-full max-w-4xl h-[70vh] mt-4 rounded-lg">
                    <ChatbotInterface/>
                </div>
            </div>
        </div>
    );
};

export default Main;