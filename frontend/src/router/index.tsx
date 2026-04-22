import { createBrowserRouter } from 'react-router-dom';
import { ProtectedRoute } from '../components/ProtectedRoute';

// Public Pages
import { Home } from '../pages/Home';
import { Login } from '../pages/Login';
import { Register } from '../pages/Register';

// Protected Pages
import { Wallet } from '../pages/Wallet';
import { Trade } from '../pages/Trade';
import AppLayout from '../App';

export const router = createBrowserRouter([
  {
    path: '/',
    element: <AppLayout />, // Standard layout wrapper if needed
    children: [
      {
        index: true,
        element: <Home />,
      },
      {
        path: 'login',
        element: <Login />,
      },
      {
        path: 'register',
        element: <Register />,
      },
      {
        element: <ProtectedRoute/>, // All children of this route require auth
        children: [
          {
            path: 'wallet',
            element: <Wallet />,
          },
          {
            path: 'trade',
            element: <Trade />,
          },
        ]
      }
    ]
  }
]);
