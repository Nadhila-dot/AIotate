import { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Sparkles, Save, RefreshCw, Eye, EyeOff, AlertCircle } from "lucide-react";
import { toast } from "sonner";
import http from "@/http";

interface AIConfig {
  AI_PROVIDER: string;
  GEMINI_API_KEY: string;
  OPENROUTER_API_KEY: string;
  AI_MAIN_MODEL: string;
  AI_UTILITY_MODEL: string;
  MAX_SESSIONS: number;
  SHEET_QUEUE_DIR: string;
}

export default function AIConfigEditor() {
  const [config, setConfig] = useState<AIConfig>({
    AI_PROVIDER: "gemini",
    GEMINI_API_KEY: "",
    OPENROUTER_API_KEY: "",
    AI_MAIN_MODEL: "",
    AI_UTILITY_MODEL: "",
    MAX_SESSIONS: 2,
    SHEET_QUEUE_DIR: "./storage/queue_data",
  });

  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [showGeminiKey, setShowGeminiKey] = useState(false);
  const [showOpenRouterKey, setShowOpenRouterKey] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    try {
      const { data } = await http.get("/api/v1/set");
      setConfig({
        AI_PROVIDER: data.AI_PROVIDER || "gemini",
        GEMINI_API_KEY: data.GEMINI_API_KEY || "",
        OPENROUTER_API_KEY: data.OPENROUTER_API_KEY || "",
        AI_MAIN_MODEL: data.AI_MAIN_MODEL || "",
        AI_UTILITY_MODEL: data.AI_UTILITY_MODEL || "",
        MAX_SESSIONS: data.MAX_SESSIONS || 2,
        SHEET_QUEUE_DIR: data.SHEET_QUEUE_DIR || "./storage/queue_data",
      });
      setLoading(false);
      setHasChanges(false);
    } catch (error) {
      toast.error("Failed to load configuration");
      setLoading(false);
    }
  };

  const saveConfig = async () => {
    setSaving(true);
    try {
      const session = localStorage.getItem("session");
      
      await http.post("/api/v1/set", config, {
        headers: {
          Authorization: `Bearer ${session}`,
        },
      });
      
      toast.success("Configuration saved successfully! Please restart the application for changes to take effect.");
      setHasChanges(false);
    } catch (error: any) {
      if (error.response?.status === 401) {
        toast.error("Unauthorized. Please log in again.");
      } else {
        toast.error("Failed to save configuration");
      }
    } finally {
      setSaving(false);
    }
  };

  const handleChange = (field: keyof AIConfig, value: string | number) => {
    setConfig((prev) => ({ ...prev, [field]: value }));
    setHasChanges(true);
  };

  const getDefaultModel = (provider: string, isMain: boolean) => {
    if (provider === "gemini") {
      return isMain ? "gemini-2.5-pro" : "gemini-2.0-flash-exp";
    } else if (provider === "openrouter") {
      return isMain
        ? "google/gemini-2.5-pro-exp-03-25:free"
        : "google/gemini-2.0-flash-exp:free";
    }
    return "";
  };

  if (loading) {
    return (
      <Card className="max-w-3xl border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
        <CardContent className="mt-1 px-2 py-2">
          <div className="text-base font-medium text-black">Loading configuration...</div>
        </CardContent>
      </Card>
    );
  }

  return (
    <Card className="max-w-3xl border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
      <CardContent className="mt-1 px-2 py-2">
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-2">
            <Sparkles className="w-8 h-8 text-black" />
            <h1
              className="text-5xl font-extrabold tracking-tight text-black"
              style={{ fontFamily: "'Space Grotesk', sans-serif" }}
            >
              Edit AI Configuration
            </h1>
          </div>
        </div>

        {hasChanges && (
          <div className="mb-4 border-2 border-yellow-400 rounded-lg p-3 bg-yellow-50 flex items-start gap-2">
            <AlertCircle className="w-5 h-5 text-yellow-700 mt-0.5" />
            <div className="text-sm text-yellow-900">
              <p className="font-bold">Unsaved Changes</p>
              <p>Remember to restart the application after saving for changes to take effect.</p>
            </div>
          </div>
        )}

        <div className="space-y-6">
          {/* AI Provider Selection */}
          <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
            <label className="block text-lg font-bold mb-3 text-black">AI Provider</label>
            <div className="flex gap-4">
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="provider"
                  value="gemini"
                  checked={config.AI_PROVIDER === "gemini"}
                  onChange={(e) => handleChange("AI_PROVIDER", e.target.value)}
                  className="w-4 h-4"
                />
                <Badge className="bg-blue-100 text-blue-800 border-blue-300 border-2 font-bold">
                  GEMINI
                </Badge>
                <span className="text-sm text-black">Google Gemini API</span>
              </label>
              <label className="flex items-center gap-2 cursor-pointer">
                <input
                  type="radio"
                  name="provider"
                  value="openrouter"
                  checked={config.AI_PROVIDER === "openrouter"}
                  onChange={(e) => handleChange("AI_PROVIDER", e.target.value)}
                  className="w-4 h-4"
                />
                <Badge className="bg-purple-100 text-purple-800 border-purple-300 border-2 font-bold">
                  OPENROUTER
                </Badge>
                <span className="text-sm text-black">Multi-Model Access</span>
              </label>
            </div>
          </div>

          {/* API Keys */}
          <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
            <label className="block text-lg font-bold mb-3 text-black">API Keys</label>
            <div className="space-y-4">
              {/* Gemini API Key */}
              <div>
                <label className="block text-sm font-semibold mb-2 text-black">
                  Gemini API Key
                  {config.AI_PROVIDER === "gemini" && (
                    <span className="text-red-600 ml-1">*</span>
                  )}
                </label>
                <div className="flex gap-2">
                  <input
                    type={showGeminiKey ? "text" : "password"}
                    value={config.GEMINI_API_KEY}
                    onChange={(e) => handleChange("GEMINI_API_KEY", e.target.value)}
                    placeholder="AIzaSy..."
                    className="flex-1 px-3 py-2 border-2 border-black rounded-lg font-mono text-sm text-black bg-white"
                  />
                  <Button
                    variant="neutral"
                    size="sm"
                    onClick={() => setShowGeminiKey(!showGeminiKey)}
                  >
                    {showGeminiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
                <p className="text-xs text-gray-700 mt-1">
                  Get your key from:{" "}
                  <a
                    href="https://aistudio.google.com/app/apikey"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-blue-600 underline"
                  >
                    Google AI Studio
                  </a>
                </p>
              </div>

              {/* OpenRouter API Key */}
              <div>
                <label className="block text-sm font-semibold mb-2 text-black">
                  OpenRouter API Key
                  {config.AI_PROVIDER === "openrouter" && (
                    <span className="text-red-600 ml-1">*</span>
                  )}
                </label>
                <div className="flex gap-2">
                  <input
                    type={showOpenRouterKey ? "text" : "password"}
                    value={config.OPENROUTER_API_KEY}
                    onChange={(e) => handleChange("OPENROUTER_API_KEY", e.target.value)}
                    placeholder="sk-or-v1-..."
                    className="flex-1 px-3 py-2 border-2 border-black rounded-lg font-mono text-sm text-black bg-white"
                  />
                  <Button
                    variant="neutral"
                    size="sm"
                    onClick={() => setShowOpenRouterKey(!showOpenRouterKey)}
                  >
                    {showOpenRouterKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
                <p className="text-xs text-gray-700 mt-1">
                  Get your key from:{" "}
                  <a
                    href="https://openrouter.ai/keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-purple-600 underline"
                  >
                    OpenRouter
                  </a>
                </p>
              </div>
            </div>
          </div>

          {/* AI Models */}
          <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
            <label className="block text-lg font-bold mb-3 text-black">AI Models</label>
            <div className="space-y-4">
              {/* Main Model */}
              <div>
                <label className="block text-sm font-semibold mb-2 text-black">
                  Main Model (LaTeX Generation)
                  <Badge className="ml-2 bg-blue-100 text-blue-800 border-blue-300 border text-xs">
                    High Quality
                  </Badge>
                </label>
                <input
                  type="text"
                  value={config.AI_MAIN_MODEL}
                  onChange={(e) => handleChange("AI_MAIN_MODEL", e.target.value)}
                  placeholder={getDefaultModel(config.AI_PROVIDER, true)}
                  className="w-full px-3 py-2 border-2 border-black rounded-lg font-mono text-sm text-black bg-white"
                />
                <p className="text-xs text-gray-700 mt-1">
                  Leave empty to use default: <code className="bg-gray-200 px-1 rounded text-black">{getDefaultModel(config.AI_PROVIDER, true)}</code>
                </p>
              </div>

              {/* Utility Model */}
              <div>
                <label className="block text-sm font-semibold mb-2 text-black">
                  Utility Model (Descriptions, Tags, Fixes)
                  <Badge className="ml-2 bg-green-100 text-green-800 border-green-300 border text-xs">
                    Fast & Efficient
                  </Badge>
                </label>
                <input
                  type="text"
                  value={config.AI_UTILITY_MODEL}
                  onChange={(e) => handleChange("AI_UTILITY_MODEL", e.target.value)}
                  placeholder={getDefaultModel(config.AI_PROVIDER, false)}
                  className="w-full px-3 py-2 border-2 border-black rounded-lg font-mono text-sm text-black bg-white"
                />
                <p className="text-xs text-gray-700 mt-1">
                  Leave empty to use default: <code className="bg-gray-200 px-1 rounded text-black">{getDefaultModel(config.AI_PROVIDER, false)}</code>
                </p>
              </div>
            </div>
          </div>

          {/* System Settings */}
          <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
            <label className="block text-lg font-bold mb-3 text-black">System Settings</label>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-semibold mb-2 text-black">Max Sessions</label>
                <input
                  type="number"
                  value={config.MAX_SESSIONS}
                  onChange={(e) => handleChange("MAX_SESSIONS", parseInt(e.target.value) || 2)}
                  min="1"
                  max="10"
                  className="w-full px-3 py-2 border-2 border-black rounded-lg text-black bg-white"
                />
              </div>

              <div>
                <label className="block text-sm font-semibold mb-2 text-black">Queue Directory</label>
                <input
                  type="text"
                  value={config.SHEET_QUEUE_DIR}
                  onChange={(e) => handleChange("SHEET_QUEUE_DIR", e.target.value)}
                  className="w-full px-3 py-2 border-2 border-black rounded-lg font-mono text-sm text-black bg-white"
                />
              </div>
            </div>
          </div>

          {/* Save Button (Bottom) */}
          <div className="flex justify-end gap-2">
            <button
              onClick={loadConfig}
              disabled={saving}
              className="bg-red-400 text-white border-2 border-black rounded-lg px-6 py-2 font-bold shadow-[4px_4px_0_0_#000] hover:bg-red-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <RefreshCw className="w-4 h-4 inline mr-2" />
              Reset Changes
            </button>
            <button
              onClick={saveConfig}
              disabled={!hasChanges || saving}
              className="bg-green-600 text-white border-2 border-black rounded-lg px-6 py-2 font-bold shadow-[4px_4px_0_0_#000] hover:bg-green-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <Save className="w-4 h-4 inline mr-2" />
              {saving ? "Saving..." : "Save Configuration"}
            </button>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}
