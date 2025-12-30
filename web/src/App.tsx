import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Detail from './pages/Detail'
import MentalModel from './pages/MentalModel'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/pr/:id" element={<Detail type="pr" />} />
        <Route path="/issue/:id" element={<Detail type="issue" />} />
        <Route path="/mental-model" element={<MentalModel />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
