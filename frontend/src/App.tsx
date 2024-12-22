import { BrowserRouter, Routes, Route } from 'react-router-dom';
import AnimationGenerator from './pages/AnimationGenerator';
import EditorPage from './pages/EditorPage';


function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AnimationGenerator />} />
        <Route path="/editor" element={<EditorPage />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App;