import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Bridge from './components/Bridge';
import Swap from './components/Swap';
import Portfolio from './components/Portfolio';
import Pools from './components/Pools';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<Layout />}>
          <Route path="/" element={<Bridge />} />
          <Route path="/bridge" element={<Bridge />} />
          <Route path="/swap" element={<Swap />} />
          <Route path="/portfolio" element={<Portfolio />} />
          <Route path="/pools" element={<Pools />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;