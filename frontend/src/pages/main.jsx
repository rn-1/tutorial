
import Textinput from "../widgets/textinput";
import ChatbotInterface from "../widgets/chatbot";
import Appbar from "../widgets/Appbar";
import '../output.css'
import React from "react";
import { useState, useEffect } from "react";
import { useRef } from 'react';

// holy SHIT this needs a lot of help

const Main = () => {

    const [messages, setMessages] = useState([])

    useEffect(() => {
        const cleanupSession = async () => {
            if (!localStorage["sessionid"]) {
                return;
            } else {
                await fetch("http://localhost:8080/cleanup", {
                    method: "POST",
                    mode: "cors",
                    body: JSON.stringify({ id: localStorage["sessionid"] }),
                    headers: {
                        'Content-Type': "application/json"
                    }
                });
            }
        };

        const handleBeforeUnload = (e) => {
            // Call cleanupSession synchronously (async not supported in beforeunload)
            cleanupSession();
        };

        window.addEventListener("beforeunload", handleBeforeUnload);

        return () => {
            window.removeEventListener("beforeunload", handleBeforeUnload);
        };
    }, []);


    
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

            // console.log(response)

            let sess = JSON.stringify(response)
            sess = JSON.parse(response)
            let token = sess.token

            localStorage.setItem("sessionid", token) // yay!
            console.log("created session with uuid ",token)

            console.log(sess.output);

            if(sess.output){
                setMessages(messages => [...messages, {id: 0, text:sess.output, isUser: false}]); 
            } else {
                console.warn("no message received?")
            }

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
                <Textinput initialVal="" _placeholder="http cloning url" id="github_url" onsubmit={extractGithub}/>
                <div id = "chatbot_area" className="w-full max-w-4xl h-[70vh] mt-4 rounded-lg">
                    <ChatbotInterface initialMessages={[]} messages={messages} setMessages={setMessages}/>
                </div>
            </div>
        </div>
    );
};

export default Main;