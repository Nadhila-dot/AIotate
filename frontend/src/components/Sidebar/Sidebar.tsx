import React from "react";
import { Home, FolderOpen, BookOpen, User, Mail, Settings, ChevronLeft, ChevronRight, Search, SheetIcon, ShieldQuestionIcon, HelpCircleIcon, Palette } from "lucide-react";
import { Outlet, Link, useLocation, Routes, Route } from "react-router-dom";
import { HomeContainer } from "../../containers/index/container";
import { LoginContainer } from "../../containers/login/loginContainer";
import JsonResponse from "../../json/json-route";
import { NotFound } from "../ScreenBlock";
import MainContent from "../Main/Maincontent";
import { useSearch } from "@/hooks/search/useSearch";

const sidebarData = [
  { name: "Home", path: "/home", icon: Home },
  { name: "Sheets", path: "/sheets", icon: SheetIcon },
  { name: "Notebooks", path: "/notebooks", icon: BookOpen },
  { name: "Designer", path: "/designer", icon: Palette },
  { name: "Help", path: "/help", icon: HelpCircleIcon },
  { name: "Settings", path: "/settings", icon: Settings },
];

const Sidebar = ({ isAuthenticated }: { isAuthenticated: boolean }) => {
  const [isCollapsed, setIsCollapsed] = React.useState(false);
  const [hoverIdx, setHoverIdx] = React.useState<number | null>(null);
  const [toggleHover, setToggleHover] = React.useState(false);
  const location = useLocation();
  const { setOpen, SearchModal } = useSearch();

  const activeIdx = sidebarData.findIndex(item => location.pathname.includes(item.path));

  const sidebarStyle: React.CSSProperties = {
    width: isCollapsed ? "80px" : "280px",
    height: "100vh",
    background: "#fff",
    border: "4px solid #222",
    borderRadius: "0 24px 24px 0",
    boxShadow: "12px 0 0 #222",
    padding: "24px 16px",
    position: "fixed",
    left: 0,
    top: 0,
    zIndex: 40,
    transition: "all 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55)",
    display: "flex",
    flexDirection: "column",
    gap: "24px",
  };

  const logoStyle: React.CSSProperties = {
    fontWeight: 900,
    fontSize: isCollapsed ? "1.2rem" : "1.8rem",
    letterSpacing: "2px",
    color: "#222",
    textTransform: "uppercase",
    fontFamily: "'Space Grotesk', sans-serif",
    textAlign: "center",
    padding: "16px 8px",
    border: "3px solid #222",
    borderRadius: "16px",
    background: "#f8f8f8",
    boxShadow: "6px 6px 0 #222",
    transition: "all 0.3s ease",
    overflow: "hidden",
    whiteSpace: "nowrap",
  };

  const searchStyle: React.CSSProperties = {
    display: "flex",
    alignItems: "center",
    gap: "12px",
    padding: "12px 16px",
    border: "3px solid #222",
    borderRadius: "16px",
    background: "#f8f8f8",
    boxShadow: "4px 4px 0 #222",
    transition: "all 0.2s ease",
  };

  const searchInputStyle: React.CSSProperties = {
    flex: 1,
    border: "none",
    background: "transparent",
    outline: "none",
    fontSize: "14px",
    fontFamily: "'Space Grotesk', sans-serif",
    fontWeight: 500,
    color: "#222",
  };

  const navStyle: React.CSSProperties = {
    display: "flex",
    flexDirection: "column",
    gap: "8px",
    flex: 1,
  };

  const navItemStyle: React.CSSProperties = {
    display: "flex",
    alignItems: "center",
    gap: "16px",
    padding: "16px",
    border: "3px solid #222",
    borderRadius: "16px",
    background: "#f8f8f8",
    boxShadow: "4px 4px 0 #222",
    cursor: "pointer",
    transition: "all 0.2s cubic-bezier(0.68, -0.55, 0.265, 1.55)",
    textDecoration: "none",
    color: "#222",
    fontWeight: 700,
    fontSize: "14px",
    fontFamily: "'Space Grotesk', sans-serif",
    textTransform: "uppercase",
    letterSpacing: "1px",
    position: "relative",
    overflow: "hidden",
    justifyContent: isCollapsed ? "center" : "flex-start",
  };

  const navItemHoverStyle: React.CSSProperties = {
    background: "#222",
    color: "#fff",
    transform: "translate(-2px, -2px)",
    boxShadow: "6px 6px 0 #222",
  };

  const navItemActiveStyle: React.CSSProperties = {
    background: "#ff6b6b",
    color: "#fff",
    boxShadow: "4px 4px 0 #222",
  };

  const toggleButtonStyle: React.CSSProperties = {
    position: "absolute",
    right: "-20px",
    top: "50%",
    transform: "translateY(-50%)",
    width: "40px",
    height: "40px",
    background: "#fff",
    border: "3px solid #222",
    borderRadius: "50%",
    boxShadow: "4px 4px 0 #222",
    cursor: "pointer",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    color: "#222",
    transition: "all 0.2s ease",
    zIndex: 50,
  };

  const toggleButtonHoverStyle: React.CSSProperties = {
    background: "#222",
    color: "#fff",
    transform: "translateY(-50%) scale(1.1)",
  };

  const getNavItemStyle = (idx: number) => {
    if (activeIdx === idx) {
      return { ...navItemStyle, ...navItemActiveStyle };
    }
    if (hoverIdx === idx) {
      return { ...navItemStyle, ...navItemHoverStyle };
    }
    return navItemStyle;
  };

  return (
    <div style={{ display: "flex", fontFamily: "'Space Grotesk', sans-serif" }}>
      {/* Sidebar */}
      <nav style={sidebarStyle}>
        {/* Logo */}
        <div style={logoStyle}>
          {isCollapsed ? "AE" : "AIotate"}
        </div>

        {/* Search (only show when expanded) */}
        {!isCollapsed && (
          <div style={searchStyle}>
            <Search size={20} color="#222" />
            <input
              type="text"
              placeholder="Search..."
              style={searchInputStyle}
              onFocus={() => setOpen(true)}
              onClick={() => setOpen(true)}
              readOnly // Prevent typing, just open modal
            />
          </div>
        )}
        <SearchModal />

        {/* Navigation Items */}
        <div style={navStyle}>
          {sidebarData.map((item, idx) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.path}
                to={item.path}
                style={getNavItemStyle(idx)}
                onMouseEnter={() => setHoverIdx(idx)}
                onMouseLeave={() => setHoverIdx(null)}
              >
                <Icon size={20} />
                {!isCollapsed && <span>{item.name}</span>}
              </Link>
            );
          })}
        </div>

        {/* Toggle Button */}
        <button
          style={toggleHover ? { ...toggleButtonStyle, ...toggleButtonHoverStyle } : toggleButtonStyle}
          onMouseEnter={() => setToggleHover(true)}
          onMouseLeave={() => setToggleHover(false)}
          onClick={() => setIsCollapsed(!isCollapsed)}
        >
          {isCollapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
        </button>
      </nav>

      {/* Main Content Area */}
      <div
        style={{
          flex: 1,
          marginLeft: isCollapsed ? "80px" : "320px",
          marginRight: "32px",
          transition: "margin-left 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55), margin-right 0.3s cubic-bezier(0.68, -0.55, 0.265, 1.55)",
        }}
      >
        <MainContent isAuthenticated={isAuthenticated} />
      </div>
       
    </div>
  );
};

export default Sidebar;