import { createBrowserRouter } from "react-router-dom";
import ErrorPage from "../error-page";
import Sources from "../components/Sources";
import LandingPage from "../components/LandingPage";
import Dashboard from "../components/dashboard/Dashboard";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <LandingPage />,
    errorElement: <ErrorPage />,
  },
  {
    path: "/sources",
    element: <Sources />,
  },
  {
    path: "/dashboard",
    element: <Dashboard />,
  },
]);
