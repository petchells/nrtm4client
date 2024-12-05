import { createBrowserRouter } from "react-router-dom";
import ErrorPage from "../error-page";
import Sources from "./Sources";
import MainGrid from "./dashboard/MainGrid";

export const router = createBrowserRouter([
  {
    path: "/",
    element: <Sources />,
    errorElement: <ErrorPage />,
  },
  {
    path: "/sources",
    element: <Sources />,
  },
  {
    path: "/queries",
    element: <Sources />,
  },
  {
    path: "/dashboard",
    element: <MainGrid />,
  },
]);
