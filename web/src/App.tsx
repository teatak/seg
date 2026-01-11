import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { ThemeProvider } from '@/components/ThemeProvider';
import Layout from '@/components/Layout';
import Dashboard from '@/pages/Dashboard';
import Segment from '@/pages/Segment';
import Corpus from '@/pages/Corpus';
import Dictionary from '@/pages/Dictionary';

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Dashboard />} />
            <Route path="segment" element={<Segment />} />
            <Route path="corpus" element={<Corpus />} />
            <Route path="dictionary" element={<Dictionary />} />
          </Route>
        </Routes>
      </BrowserRouter>
    </ThemeProvider>
  );
}

export default App;
