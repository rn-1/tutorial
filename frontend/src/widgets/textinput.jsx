import logo from '../logo.svg';
import '../output.css'

const Textinput = ({initialVal = "", _placeholder = "", id = "text_input", onsubmit}) => {


    return(
        <div id="textinputcontainer" className="p-15 flex w-8/12 bg-white h-8 shadow rounded-[15px] items-center">
            <div id="icon_container" className="flex items-center mr-4">
                <img src={logo} alt="placeholder" className="w-4 h-4"/>
            </div>
            <form onSubmit={(e) => {e.preventDefault(); onsubmit();}} className="flex flex-1 w-full focus-within:outline-none">
                <input
                    className="w-full outline-none focus:outline-none"
                    placeholder={_placeholder}
                    id={id}
                    defaultValue={initialVal}
                />
            </form>
        </div>
    )
}


export default Textinput;