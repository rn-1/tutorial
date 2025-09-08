import "../output.css"
import ReactMarkdown from 'react-markdown';

const TextBox = ({isUser, text}) => {

    // conditional
    if(isUser){
        return(
            <div className="w-full bg-opacity-0 text-white py-4 px-2 flex justify-end w-full">
                <div className="items-right w-11/12 p-2 bg-gray-950 rounded">
                    <p>{text}</p>
                </div>
            </div>
        )
    }
    else{
        return(
            <div className = "w-full bg-opacity-0 text-white py-4 px-2 flex justify-start w-full">
                <div className="items-left w-11/12 p-2 bg-gray-900 rounded">
                    <p>{text}</p>
                    
                </div>
            </div>
        )
    }
    
}

export default TextBox