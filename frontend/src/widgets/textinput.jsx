import logo from '../logo.svg';
import '../output.css'

const Textinput = ({initialVal = "", _placeholder = "", id = "text_input", onsubmit}) => {


    return(
        <div id="textinputcontainer" className="flex w-8/12 bg-white h-8 shadow rounded-[15px] items-center border-4 border-red-500 ">
            <div id="icon_container" className="flex items-center mr-4">
                <img src={logo} alt="placeholder" className="w-4 h-4"/>
            </div>
            <form onSubmit={(e) => {e.preventDefault(); onsubmit();}} className="flex flex-1 w-full p-10">
                <input
                    className="border-0 w-full outline-none"
                    placeholder={_placeholder}
                    id={id}
                    defaultValue={initialVal}
                />
            </form>
        </div>
    )
}


export default Textinput;