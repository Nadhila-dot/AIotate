import React, { useEffect, useState } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Info, Calendar, User, Package } from "lucide-react";
import http from "@/http";

interface SystemData {
  build?: string;
  date?: string;
  author?: string;
}

export default function SystemInfoCard() {
  const [systemData, setSystemData] = useState<SystemData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    http
      .get("/api/v1/system")
      .then((res) => {
        setSystemData(res.data?.data);
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, []);

  return (
    <Card className="max-w-3xl border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
      <CardContent className="mt-1 px-2 py-2">
        <h1
          className="text-5xl font-extrabold tracking-tight mb-6"
          style={{ fontFamily: "'Space Grotesk', sans-serif" }}
        >
          System Information
        </h1>

        {loading ? (
          <div className="text-base font-medium text-gray-600">Loading system info...</div>
        ) : !systemData ? (
          <div className="text-base font-medium text-red-600">Failed to load system info</div>
        ) : (
          <div className="space-y-4">
            <div className="border-2 border-black rounded-lg p-4 bg-gray-50">
              <div className="space-y-3">
                <div className="flex items-center gap-3">
                  <Package className="w-5 h-5 text-blue-600" />
                  <div>
                    <span className="font-semibold text-gray-700">Build Version:</span>
                    <p className="text-lg font-bold">{systemData.build || "Unknown"}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <Calendar className="w-5 h-5 text-green-600" />
                  <div>
                    <span className="font-semibold text-gray-700">Release Date:</span>
                    <p className="text-lg font-bold">{systemData.date || "Unknown"}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <User className="w-5 h-5 text-purple-600" />
                  <div>
                    <span className="font-semibold text-gray-700">Author:</span>
                    <p className="text-lg font-bold">{systemData.author || "Unknown"}</p>
                  </div>
                </div>
              </div>
            </div>

          {/*}
            <div className="border-2 border-green-300 rounded-lg p-4 bg-green-50">
              <div className="flex items-start gap-2">
                <Info className="w-5 h-5 text-green-700 mt-0.5" />
                <div className="text-sm text-green-900">
                  <p className="font-bold mb-1">AIotate - AI-Powered Educational Worksheet Generator</p>
                  <p>
                    An open-source platform for creating question papers and LaTeX documents using AI.
                    Supports both Gemini and OpenRouter for flexible AI model selection.
                  </p>
                </div>
              </div>
            </div> */}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
