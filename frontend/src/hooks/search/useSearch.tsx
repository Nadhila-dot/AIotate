import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";

export function useSearch() {
  const [open, setOpen] = useState(false);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.code === "Space") {
        e.preventDefault();
        setOpen(true);
      }
      if (e.key === "Escape") {
        setOpen(false);
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  const SearchModal = () =>
    open ? (
      <div 
        className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4"
        onClick={() => setOpen(false)}
      >
        <div
          className="bg-white text-black border-4 border-black rounded-xl shadow-[12px_12px_0_0_#000] p-8 max-w-md w-full text-center transform transition-all"
          style={{
            fontFamily: "'Space Grotesk', sans-serif",
          }}
          onClick={(e) => e.stopPropagation()}
        >
          <div className="mb-6">
            <img
              src="/undraw/not-done.svg"
              alt="Under Construction"
              className="w-64 h-64 mx-auto opacity-90"
            />
          </div>
          
          <h2 className="text-4xl font-extrabold mb-3 tracking-tight text-gray-900">
            Search Coming Soon
          </h2>
          
          <p className="text-base text-gray-600 mb-6 leading-relaxed">
            We're working hard to bring you powerful search functionality. 
            Stay tuned for updates!
          </p>
          
          <div className="bg-gray-50 border-2 border-gray-200 rounded-lg p-3 mb-6">
            <p className="text-sm text-gray-500">
              <span className="font-bold text-gray-700">Tip:</span> Press{" "}
              <kbd className="px-2 py-1 bg-white border border-gray-300 rounded text-xs font-mono">
                Cmd/Ctrl + Space
              </kbd>{" "}
              to open search
            </p>
          </div>
          
          <Button
            variant="noShadow"
            size="lg"
            className="bg-red-600 text-white border-2 border-black rounded-lg px-6 py-3 font-bold shadow-[4px_4px_0_0_#000] hover:bg-red-700 hover:shadow-[6px_6px_0_0_#000] transition-all active:shadow-[2px_2px_0_0_#000] active:translate-x-1 active:translate-y-1"
            onClick={() => setOpen(false)}
          >
            Got it!
          </Button>
        </div>
      </div>
    ) : null;

  return { open, setOpen, SearchModal };
}