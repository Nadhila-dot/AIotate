import { HomeContainer } from "@/containers/index/container";
import { LoginContainer } from "@/containers/login/loginContainer";
import JsonContainer from "@/json/json-route";
import { Routes, Route, Link } from "react-router-dom";
import { NotFound } from "../ScreenBlock";
import { SettingsContainer } from "@/containers/settings/settingsContainer";
import { HelpContainer } from "@/containers/help/helpContainer";
import { SheetsContainer } from "@/containers/sheets/sheetsContainer";
import { SheetsStatusContainer } from "@/containers/sheets/sheetsStatus";
import { NotebooksContainer } from "@/containers/notebooks/notebookContainer";
import { DesignerContainer } from "@/containers/designer/designerContainer";


interface MainContentProps {
    isAuthenticated: boolean;
}

const MainContent: React.FC<MainContentProps> = ({ isAuthenticated }) => (
    <Routes>
        <Route path="/home" element={<HomeContainer />} />
        <Route path="/settings" element={<SettingsContainer />} />
        <Route path="/help" element={<HelpContainer />} />
        <Route path="/sheets" element={<SheetsContainer />} />
        <Route path="/sheets/status" element={<SheetsStatusContainer />} />
        <Route path="/notebooks" element={<NotebooksContainer />} />
        <Route path="/designer" element={<DesignerContainer />} />
        
        <Route path="/login" element={<LoginContainer />} />
        {isAuthenticated && (
            <Route path="/json" element={<JsonContainer />} />
        )}
        {/*//@ts-ignore */}
        <Route path="*" element={<NotFound path={window.location.pathname} />} />
    </Routes>
);

export default MainContent;