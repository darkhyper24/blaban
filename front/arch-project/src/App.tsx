import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import { AuthProvider } from './context/authContext';
import NavBar from './components/navbar';
import Home from './pages/home';
import LoginPage from './pages/login';
import SignupPage from './pages/signup';
import CategoryItems from './pages/CategoryItems';

// Layout component that includes navbar for routes that need it
const MainLayout: React.FC<{ children: React.ReactNode }> = ({ children }) => (
  <>
    <NavBar />
    {children}
  </>
);

function App() {
  return (
    <AuthProvider>
      <Router>
        <Routes>
          {/* Home page with navbar */}
          <Route path="/" element={
            <MainLayout>
              <Home />
            </MainLayout>
          } />
          
          {/* Category items page with navbar */}
          <Route path="/category/:categoryId" element={
            <MainLayout>
              <CategoryItems />
            </MainLayout>
          } />
          
          {/* Auth pages without navbar */}
          <Route path="/login" element={<LoginPage />} />
          <Route path="/signup" element={<SignupPage />} />
        </Routes>
      </Router>
    </AuthProvider>
  );
}

export default App;