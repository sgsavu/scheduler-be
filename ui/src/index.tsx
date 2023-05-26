import { useState } from 'react'
import { Infer } from './components/Infer'
import { Navbar } from './components/Navbar'
import { Tasks } from './components/Tasks'
import { Train } from './components/Train'
import {
  Infer as InferIcon,
  Tasks as TasksIcon,
  Train as TrainIcon
} from './components/Icons'
import './index.css'

const tabMap: Record<number, JSX.Element> = {
  0: <Infer></Infer>,
  1: <Train></Train>,
  2: <Tasks></Tasks>,
}

const NAV_ITEMS = [
  (
    <>
      <InferIcon />
      <span>Infer</span>
    </>
  ),
  (
    <>
      <TrainIcon />
      <span>Train</span>
    </>
  ),
  (
    <>
      <TasksIcon />
      <span>Tasks</span>
    </>
  ),
]

function App() {
  const [tab, setTab] = useState<number>(0)

  return (
    <div className='appcontainer'>
      <Navbar
        items={NAV_ITEMS}
        onSelectedTabChange={setTab}
        selectedTab={tab}
      />
      <div className='tab'>
        {tabMap[tab]}
      </div>
    </div>
  )
}

export default App
