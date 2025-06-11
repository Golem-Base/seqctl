import "./index.css";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import { NetworkList } from "@/components/features/NetworkList";
import { NetworkDetail } from "@/components/features/NetworkDetail";
import { Layout } from "@/components/layout/Layout";
import { ModalManager } from "@/components/modals/ModalManager";
import { Toaster } from "@/components/ui/toaster";

export function App() {
  return (
    <BrowserRouter>
      <Layout>
        <div className="container mx-auto px-4 py-8">
          <Routes>
            <Route path="/" element={<NetworkList />} />
            <Route path="/networks/:networkId" element={<NetworkDetail />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </div>
        <ModalManager />
        <Toaster />
      </Layout>
    </BrowserRouter>
  );
}

export default App;
