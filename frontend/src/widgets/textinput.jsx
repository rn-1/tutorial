import linkicon from '../link_icon.png'
import '../output.css';
// import '../main.css';

const Textinput = ({initialVal = "", _placeholder = "", id = "text_input", onsubmit}) => {


    return(
        <div id="textinputcontainer" className="pl-m-4 w-8/12 bg-white h-10 shadow rounded-[15px] flex flex-row items-center px-4">
            <div id="icon_container" className="flex items-center">
                <img src={linkicon} alt="placeholder" className="w-4 h-4"/>
            </div>
            <form onSubmit={(e) => {e.preventDefault(); onsubmit();}} className="flex w-11/12 focus-within:outline-none px-2">
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