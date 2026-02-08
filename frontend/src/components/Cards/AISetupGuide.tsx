import React, { useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { BookOpen, ExternalLink, ChevronDown, ChevronUp } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function AISetupGuide() {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <Card className="max-w-3xl border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
      <CardContent className="mt-1 px-2 py-2">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-2">
            <BookOpen className="w-6 h-6" />
            <h1
              className="text-3xl font-extrabold tracking-tight"
              style={{ fontFamily: "'Space Grotesk', sans-serif" }}
            >
              AI Setup Guide
            </h1>
          </div>
          <Button
            variant="outline"
            size="sm"
            onClick={() => setIsExpanded(!isExpanded)}
            className="border-2 border-black"
          >
            {isExpanded ? (
              <>
                <ChevronUp className="w-4 h-4 mr-1" />
                Hide
              </>
            ) : (
              <>
                <ChevronDown className="w-4 h-4 mr-1" />
                Show Guide
              </>
            )}
          </Button>
        </div>

        {isExpanded && (
          <div className="space-y-4">
            {/* Gemini Setup */}
            <div className="border-2 border-blue-300 rounded-lg p-4 bg-blue-50">
              <h3 className="text-xl font-bold mb-2 text-blue-900">Option 1: Using Gemini (Recommended)</h3>
              <ol className="list-decimal list-inside space-y-2 text-sm text-blue-900">
                <li>
                  Get a free API key from{" "}
                  <a
                    href="https://aistudio.google.com/app/apikey"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-bold underline inline-flex items-center gap-1"
                  >
                    Google AI Studio
                    <ExternalLink className="w-3 h-3" />
                  </a>
                </li>
                <li>
                  Open <code className="bg-blue-100 px-2 py-0.5 rounded font-mono">set.json</code> in the server directory
                </li>
                <li>
                  Set <code className="bg-blue-100 px-2 py-0.5 rounded font-mono">"AI_PROVIDER": "gemini"</code>
                </li>
                <li>
                  Add your API key to <code className="bg-blue-100 px-2 py-0.5 rounded font-mono">"GEMINI_API_KEY"</code>
                </li>
                <li>Restart the application</li>
              </ol>
              <div className="mt-3 p-2 bg-blue-100 rounded">
                <p className="text-xs font-mono text-blue-900">
                  {`{`}<br />
                  &nbsp;&nbsp;"AI_PROVIDER": "gemini",<br />
                  &nbsp;&nbsp;"GEMINI_API_KEY": "AIzaSy...",<br />
                  &nbsp;&nbsp;...<br />
                  {`}`}
                </p>
              </div>
            </div>

            {/* OpenRouter Setup */}
            <div className="border-2 border-purple-300 rounded-lg p-4 bg-purple-50">
              <h3 className="text-xl font-bold mb-2 text-purple-900">Option 2: Using OpenRouter</h3>
              <ol className="list-decimal list-inside space-y-2 text-sm text-purple-900">
                <li>
                  Get an API key from{" "}
                  <a
                    href="https://openrouter.ai/keys"
                    target="_blank"
                    rel="noopener noreferrer"
                    className="font-bold underline inline-flex items-center gap-1"
                  >
                    OpenRouter
                    <ExternalLink className="w-3 h-3" />
                  </a>
                </li>
                <li>
                  Open <code className="bg-purple-100 px-2 py-0.5 rounded font-mono">set.json</code> in the server directory
                </li>
                <li>
                  Set <code className="bg-purple-100 px-2 py-0.5 rounded font-mono">"AI_PROVIDER": "openrouter"</code>
                </li>
                <li>
                  Add your API key to <code className="bg-purple-100 px-2 py-0.5 rounded font-mono">"OPENROUTER_API_KEY"</code>
                </li>
                <li>Restart the application</li>
              </ol>
              <div className="mt-3 p-2 bg-purple-100 rounded">
                <p className="text-xs font-mono text-purple-900">
                  {`{`}<br />
                  &nbsp;&nbsp;"AI_PROVIDER": "openrouter",<br />
                  &nbsp;&nbsp;"OPENROUTER_API_KEY": "sk-or-v1-...",<br />
                  &nbsp;&nbsp;...<br />
                  {`}`}
                </p>
              </div>
            </div>

            {/* Custom Models */}
            <div className="border-2 border-green-300 rounded-lg p-4 bg-green-50">
              <h3 className="text-xl font-bold mb-2 text-green-900">Custom Models (Optional)</h3>
              <p className="text-sm text-green-900 mb-2">
                You can specify custom models for different tasks:
              </p>
              <ul className="list-disc list-inside space-y-1 text-sm text-green-900 ml-4">
                <li>
                  <code className="bg-green-100 px-2 py-0.5 rounded font-mono">AI_MAIN_MODEL</code> - For LaTeX generation (high quality)
                </li>
                <li>
                  <code className="bg-green-100 px-2 py-0.5 rounded font-mono">AI_UTILITY_MODEL</code> - For descriptions, tags, fixes (fast)
                </li>
              </ul>
              <p className="text-xs text-green-800 mt-2">
                Leave empty to use recommended defaults for your provider.
              </p>
            </div>

            {/* Help */}
            <div className="border-2 border-gray-300 rounded-lg p-4 bg-gray-50">
              <h3 className="text-lg font-bold mb-2">Need Help?</h3>
              <p className="text-sm text-gray-700">
                Check the{" "}
                <code className="bg-gray-200 px-2 py-0.5 rounded font-mono">AI_PROVIDER_SETUP.md</code>{" "}
                file in the server directory for detailed instructions, troubleshooting, and model recommendations.
              </p>
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  );
}
