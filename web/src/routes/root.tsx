import { createBrowserRouter } from "react-router-dom";
import ErrorPage from "../error-page";
import SourceList from "../components/sourcelist";
import LandingPage from "../components/LandingPage";
import Dashboard from "../components/dashboard/Dashboard";

export const router = createBrowserRouter([
  {
    path: '/',
    element: < LandingPage />,
    errorElement: <ErrorPage />,
  },
  {
    path: '/sources',
    element: < SourceList />,
  },
  {
    path: '/dashboard',
    element: < Dashboard />,
  },
]);


