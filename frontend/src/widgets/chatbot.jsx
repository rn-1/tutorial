import React, { useState, useRef, useEffect } from 'react';
import '../output.css'

const ChatbotInterface = ({ initialMessages = [] }) => {
    const [messages, setMessages] = useState(initialMessages);
    const [currentInput, setCurrentInput] = useState('');
    const messagesEndRef = useRef(null);
    
    // Auto-scroll to bottom when messages change
    useEffect(() => {
      scrollToBottom();
    }, [messages]);
    
    const scrollToBottom = () => {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    
    const handleInputChange = (e) => {
      setCurrentInput(e.target.value);
    };
    
    const handleSubmit = (e) => {
      e.preventDefault();
      if (currentInput.trim() === '') return;

      // Add user message to chat
      const newUserMessage = {
        id: Date.now(),
        text: currentInput,
        isUser: true,
      };

      
      setMessages([...messages, newUserMessage]);
      setCurrentInput('');
      
      // Here you would typically handle sending the message to your backend
      // and then add the response to the messages
      // This is a placeholder for that functionality
        const test = async () => {
            try{
                const response = await fetch("http://127.0.0.1:8081/", {method:"GET", mode: "cors"}).then(response => response.text());
                if(!response.ok){
                    console.log("http error");
                } else{
                    console.log(response);
                }
            } catch(e){
                console.log(e);
            }
        }
        const testResponse = test();
        console.log(testResponse);
        const botResponse = {
            id: Date.now() + 1,
            text: "test",
            isUser: false,
        };
        
        setMessages(prev => [...prev, botResponse]);
    }


    return (
        <div className="flex flex-col h-full w-full m-40">
            <div className="flex flex-col w-full h-full bg-gray-400 bg-opacity-30 rounded-lg overflow-hidden items-center py-10">
                {/* Chat messages display area */}
                <div className="h-full w-8/12 min-h-[200px] px-6 overflow-y-auto">
                {messages.map((message) => (
                    <div 
                    key={message.id} 
                    className={`mb-4 p-2.5 rounded-lg w-full ${
                        message.isUser 
                        ? 'ml-auto bg-blue-500 text-white' 
                        : 'mr-auto bg-gray-600 bg-opacity-50 text-white'
                    }`}
                    >
                    {message.text}
                    </div>
                ))}
                <div ref={messagesEndRef} />
                </div>
                
                {/* Input area */}
                <form onSubmit={handleSubmit} className="mt-4 p-40 border-t border-gray-600 w-7/12">
                    <div className="flex items-center">
                    <input
                        type="text"
                        value={currentInput}
                        onChange={handleInputChange}
                        placeholder="ask a question about the repository"
                        className="flex-grow px-4 py-2 bg-white bg-opacity-90 rounded-full text-gray-800 focus:outline-none"
                    />
                    <button
                        type="submit"
                        className="ml-2 rounded-full bg-gray-800 text-white p-2 hover:bg-gray-700 focus:outline-none"
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