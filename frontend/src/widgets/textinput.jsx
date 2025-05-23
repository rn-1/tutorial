import logo from '../logo.svg';
import '../output.css'

const Textinput = ({initialVal = "", _placeholder = "", id = "text_input", onsubmit}) => {


    return(
        <div id = "textinputcontainer" className = "flex w-8/12 bg-white h-8 shadow rounded-[15px] p-10 items-center content-evenly border-4 border-red-500">
            <div id = "icon_container" className="flex">
                <i src={logo} alt = "placeholder" className = "w-4"/>
            </div>
            <form onSubmit={(e) => {e.preventDefault(); onsubmit();}}>
                <input className = "border-0 w-full" placeholder={_placeholder} id = {id}></input>
            </form>
        </div>
    )
}


export default Textinput;