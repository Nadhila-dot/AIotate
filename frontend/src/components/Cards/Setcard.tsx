import React, { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { getSystemInfo } from "@/scripts/getSet";
import { Badge } from "@/components/ui/badge";
import { Sparkles, Zap, Key, Settings, Database } from "lucide-react";

interface AIConfig {
  AI_PROVIDER?: string;
  GEMINI_API_KEY?: string;
  OPENROUTER_API_KEY?: string;
  AI_MAIN_MODEL?: string;
  AI_UTILITY_MODEL?: string;
  MAX_SESSIONS?: number;
  SHEET_QUEUE_DIR?: string;
}

export default function SetCard() {
  const [setData, setSetData] = useState<AIConfig | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getSystemInfo()
      .then((data) => {
        setSetData(data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  const maskApiKey = (key: string | undefined) => {
    if (!key || key === "") return "Not configured";
    if (key.length <= 8) return "••••••••";
    return key.substring(0, 8) + "•".repeat(Math.min(key.length - 8, 20));
  };

  const getProviderBadge = (provider: string | undefined) => {
    if (!provider) return null;
    
    const colors = {
      gemini: "bg-blue-100 text-blue-800 border-blue-300",
      openrouter: "bg-purple-100 text-purple-800 border-purple-300",
    };
    
    return (
      <Badge className={`${colors[provider as keyof typeof colors] || "bg-gray-100 text-gray-800"} border-2 font-bold`}>
        {provider.toUpperCase()}
      </Badge>
    );
  };

  const getDefaultModel = (provider: string | undefined, isMain: boolean) => {
    if (provider === "gemini") {
      return isMain ? "gemini-2.5-pro" : "gemini-2.0-flash-exp";
    } else if (provider === "openrouter") {
      return isMain 
        ? "google/gemini-2.5-pro-exp-03-25:free" 
        : "google/gemini-2.0-flash-exp:free";
    }
    return "Not configured";
  };

  const isKeyConfigured = (key: string | undefined) => {
    return key && key !== "";
  };

  return (
    <Card className="max-w-3xl border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
      <CardContent className="mt-1 px-2 py-2">
        <h1 
          className="text-5xl font-extrabold tracking-tight mb-6" 
          style={{ fontFamily: "'Space Grotesk', sans-serif" }}
        >
          AI Configuration
        </h1>

        {loading ? (
          <div className="text-base font-medium text-gray-600">Loading configuration...</div>
        ) : !setData ? (
          <div className="text-base font-medium text-red-600">Failed to load configuration</div>
        ) : (
          <div className="space-y-6">
            {/* AI Provider Section */}
            <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
              <div className="flex items-center gap-2 mb-3">
                <Sparkles className="w-5 h-5" />
                <h2 className="text-xl font-bold">AI Provider</h2>
              </div>
              <div className="flex items-center gap-3">
                {getProviderBadge(setData.AI_PROVIDER)}
                <span className="text-sm text-gray-600">
                  {setData.AI_PROVIDER === "gemini" && "Google Gemini API"}
                  {setData.AI_PROVIDER === "openrouter" && "OpenRouter (Multi-Model Access)"}
                  {!setData.AI_PROVIDER && "Not configured"}
                </span>
              </div>
            </div>

            {/* API Keys Section */}
            <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
              <div className="flex items-center gap-2 mb-3">
                <Key className="w-5 h-5" />
                <h2 className="text-xl font-bold">API Keys</h2>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-semibold">Gemini API Key:</span>
                  <div className="flex items-center gap-2">
                    <code className="bg-white px-3 py-1 rounded border border-gray-300 text-sm font-mono">
                      {maskApiKey(setData.GEMINI_API_KEY)}
                    </code>
                    {isKeyConfigured(setData.GEMINI_API_KEY) ? (
                      <Badge className="bg-green-100 text-green-800 border-green-300 border">
                        Active
                      </Badge>
                    ) : (
                      <Badge className="bg-gray-100 text-gray-600 border-gray-300 border">
                        Not Set
                      </Badge>
                    )}
                  </div>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-semibold">OpenRouter API Key:</span>
                  <div className="flex items-center gap-2">
                    <code className="bg-white px-3 py-1 rounded border border-gray-300 text-sm font-mono">
                      {maskApiKey(setData.OPENROUTER_API_KEY)}
                    </code>
                    {isKeyConfigured(setData.OPENROUTER_API_KEY) ? (
                      <Badge className="bg-green-100 text-green-800 border-green-300 border">
                        Active
                      </Badge>
                    ) : (
                      <Badge className="bg-gray-100 text-gray-600 border-gray-300 border">
                        Not Set
                      </Badge>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* Models Section */}
            <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
              <div className="flex items-center gap-2 mb-3">
                <Zap className="w-5 h-5" />
                <h2 className="text-xl font-bold">AI Models</h2>
              </div>
              <div className="space-y-3">
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-semibold">Main Model (LaTeX Generation):</span>
                    <Badge className="bg-blue-100 text-blue-800 border-blue-300 border text-xs">
                      High Quality
                    </Badge>
                  </div>
                  <code className="block bg-white px-3 py-2 rounded border border-gray-300 text-sm font-mono">
                    {setData.AI_MAIN_MODEL || getDefaultModel(setData.AI_PROVIDER, true)}
                  </code>
                  {!setData.AI_MAIN_MODEL && (
                    <p className="text-xs text-gray-500 mt-1">Using default model</p>
                  )}
                </div>
                <div>
                  <div className="flex items-center gap-2 mb-1">
                    <span className="font-semibold">Utility Model (Descriptions, Tags, Fixes):</span>
                    <Badge className="bg-green-100 text-green-800 border-green-300 border text-xs">
                      Fast & Efficient
                    </Badge>
                  </div>
                  <code className="block bg-white px-3 py-2 rounded border border-gray-300 text-sm font-mono">
                    {setData.AI_UTILITY_MODEL || getDefaultModel(setData.AI_PROVIDER, false)}
                  </code>
                  {!setData.AI_UTILITY_MODEL && (
                    <p className="text-xs text-gray-500 mt-1">Using default model</p>
                  )}
                </div>
              </div>
            </div>

            {/* System Settings Section */}
            <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
              <div className="flex items-center gap-2 mb-3">
                <Settings className="w-5 h-5" />
                <h2 className="text-xl font-bold">System Settings</h2>
              </div>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="font-semibold">Max Sessions:</span>
                  <Badge className="bg-gray-100 text-gray-800 border-gray-300 border">
                    {setData.MAX_SESSIONS || 2}
                  </Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="font-semibold">Queue Directory:</span>
                  <code className="bg-white px-3 py-1 rounded border border-gray-300 text-xs font-mono">
                    {setData.SHEET_QUEUE_DIR || "./storage/queue_data"}
                  </code>
                </div>
              </div>
            </div>

            {/* Info Box */}
            <div className="border-2 border-blue-300 rounded-lg p-4 bg-blue-50">
              <p className="text-sm text-blue-900">
                <span className="font-bold">💡 Tip:</span> To change these settings, edit the{" "}
                <code className="bg-blue-100 px-2 py-0.5 rounded font-mono text-xs">set.json</code>{" "}
                file in the server directory and restart the application.
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}