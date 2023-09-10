import { JSXElement, createSignal } from 'solid-js'
import { Navbar } from './components/Navbar'
import { Infer } from './components/Infer'
import './index.css'
import { Train } from './components/Train'

const tabMap: Record<number, JSXElement> = {
  0: <Infer></Infer>,
  1: <Train></Train>,
  // 2: <Tasks></Tasks>,
}

function App() {
  const [getTab, setTab] = createSignal(0)

  return (
      <div class="appcontainer">
        <Navbar
          getSelected={getTab}
          setSelected={setTab}
        />
        <div class='tab'>
          {tabMap[getTab()]}
        </div>
      </div>
  )
}

export default App
