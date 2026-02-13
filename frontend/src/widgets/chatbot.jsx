import React, { useState, useRef, useEffect } from 'react';
// import '../main.css'
import '../output.css';


import TextBox from "./Textbox"




const ChatbotInterface = ({ initialMessages = [], messages, setMessages }) => {
    const [currentInput, setCurrentInput] = useState('');
    const [hasToken, setToken] = useState(0);
    const [loading, setLoading] = useState(0);
    const messagesEndRef = useRef(null);

    console.log(hasToken); // shitty hack fix

    // Auto-scroll to bottom when messages change # this looks unused?
    useEffect(() => {
        if (localStorage.getItem("sessionid")) {
            setToken(1);
        }
        // scrollToBottom();
    }, [messages]);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };


    const handleInputChange = (e) => {
        setCurrentInput(e.target.value);
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (currentInput.trim() === '' || !localStorage.getItem("sessionid")) return;

        // Add user message to chat
        const newUserMessage = {
            id: Date.now(),
            text: currentInput,
            isUser: true,
        };

        setMessages([...messages, newUserMessage]);
        setCurrentInput('');
        try {
            const sessionId = localStorage.getItem('sessionid');
            setLoading(1);
            const response = await fetch("http://tutorial-vfba.onrender.com/queryRepo", {
                method: "POST",
                mode: "cors",
                body: JSON.stringify({
                    text: newUserMessage.text,
                    sessionid: sessionId
                }),
                headers: {
                    'Content-Type': 'application/json'
                },
            }
            ).then(resp => resp.text());
            if (!response.ok) {
                console.log("http error");
                console.log(response);
            } else {
                console.log(response);
            }
            setLoading(0);

            var sess = JSON.stringify(response);
            sess = JSON.parse(response);
        } catch (e) {
            console.log(e);
        }
        // console.log(testResponse);
        const botResponse = {
            id: Date.now(),
            text: sess.output,
            isUser: false,
        };

        setMessages(prev => [...prev, botResponse]);
        scrollToBottom();
    }


    return (
        <div className="flex flex-col h-full w-full">
            <div className="flex flex-col w-full h-full bg-gray-600 bg-opacity-30 rounded-lg overflow-hidden items-center py-10">
                {/* Chat messages display area */}
                <div className="h-full w-11/12 min-h-[200px] px-6 overflow-y-auto">
                    <div className={`flex justify-center items-center py-4 ${loading? "visible" : "hidden"}`}>
                        <span className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-gray-700"></span>
                        <span className="ml-2 text-gray-400">Loading...</span>
                    </div>
                    {messages.map((message) => (
                        <TextBox
                            key={message.id}
                            isUser={message.isUser}
                            text={message.text}
                        />
                    ))}
                    <div ref={messagesEndRef} />
                </div>
                    <div className={`w-11/12 ${messages.length ? "hidden h-0" : "h-24 visible"} align-middle items-center text-4xl text-gray-500 text-center`}>
                    <h1>
                        Clone a git repository to get started.
                    </h1>
                </div>
                {/* Input area */}
                <form onSubmit={handleSubmit} className="border-t border-gray-600 pt-4 w-11/12">
                    <div className="items-center flex">
                        <input
                            type="text"
                            value={currentInput}
                            onChange={handleInputChange}
                            placeholder="ask a question about the repository"
                            className="flex-grow px-4 py-2 bg-gray-500 bg-opacity-90 rounded-full text-white focus:outline-none"
                        />
                        <button
                            type="submit"
                            disabled={currentInput.trim() === ''}
                            className={`ml-2 rounded-full transition-colors duration-300 ${currentInput.trim() === '' ? "bg-gray-500 cursor-not-allowed" : "bg-black hover:bg-gray-700"} text-white p-2 focus:outline-none`}
                        >
                            <svg xmlns="http://www.w3.org/2000/svg" className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 15l7-7 7 7" />
                            </svg>
                        </button>
                    </div>
                </form>
            </div>
        </div>
    );
};




export default ChatbotInterface;