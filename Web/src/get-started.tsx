import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import GetStartedPage from "./pages/GetStartedPage";

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <GetStartedPage />
  </StrictMode>,
);
