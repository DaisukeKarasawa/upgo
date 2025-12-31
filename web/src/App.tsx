import { BrowserRouter, Routes, Route } from 'react-router-dom'
import Dashboard from './pages/Dashboard'
import Detail from './pages/Detail'

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/change/:id" element={<Detail />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
