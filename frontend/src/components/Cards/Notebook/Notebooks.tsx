import React, { useEffect, useState, useCallback } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Loader2, Edit, Trash2, X, RefreshCw, BookOpen } from "lucide-react";
import { addSheetToNotebook } from "@/scripts/notebooks/addItems";
import CreateNotebook from "./createNotebook";
import { toast } from "sonner";
import http from "@/http";
import { NeoButton } from "@/components/ui/neo-button";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
  DialogClose,
} from "@/components/ui/dialog";

// API functions for notebook operations
const deleteItemFromNotebook = async (notebookId: number, itemName: string) => {
  const res = await http.delete(`/api/v1/notebooks/${notebookId}/items/${encodeURIComponent(itemName)}`);
  return res.data;
};

const deleteNotebook = async (notebookId: number) => {
  const res = await http.delete(`/api/v1/notebooks/${notebookId}`);
  return res.data;
};

const updateNotebook = async (notebookId: number, data: any) => {
  const res = await http.put(`/api/v1/notebooks/${notebookId}`, data);
  return res.data;
};

// Function to get notebooks with cache busting
const getNotebooks = async () => {
  // Add timestamp to prevent caching
  const timestamp = new Date().getTime();
  const { data } = await http.get(`/api/v1/notebooks?_t=${timestamp}`);
  return data;
};

export default function NotebookListCard() {
  const [notebooks, setNotebooks] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [editNotebook, setEditNotebook] = useState<any | null>(null);
  const [editingNotebook, setEditingNotebook] = useState<any | null>(null);
  const [sheets, setSheets] = useState<any[]>([]);
  const [sheetLoading, setSheetLoading] = useState(false);
  const [viewNotebook, setViewNotebook] = useState<any | null>(null);
  const [viewSheets, setViewSheets] = useState<any[]>([]);
  const [refreshCounter, setRefreshCounter] = useState(0);

  // Form state for editing notebook
  const [editName, setEditName] = useState("");
  const [editDescription, setEditDescription] = useState("");
  const [editTags, setEditTags] = useState("");
  const [editColor, setEditColor] = useState("");

  // Fetch notebooks with cache busting
  const refreshNotebooks = useCallback(() => {
    setLoading(true);
    getNotebooks()
      .then((data) => {
        setNotebooks(data || []);
      })
      .catch(error => {
        console.error("Failed to fetch notebooks:", error);
        toast.error("Failed to load notebooks");
      })
      .finally(() => setLoading(false));
  }, []);

  // Refresh when the component mounts or refreshCounter changes
  useEffect(() => {
    refreshNotebooks();
   // toast.info("Notebooks refreshed");
  }, [refreshCounter, refreshNotebooks]);

  // Mutate function to trigger refreshes
  const mutateData = () => {
    setRefreshCounter(prev => prev + 1);
  };

  // Fetch sheets for adding to notebook with cache busting
  const fetchSheets = async () => {
    setSheetLoading(true);
    try {
      const timestamp = new Date().getTime();
      const { data } = await http.get(`/api/v1/sheets/get?latest=true&obj_num=50&_t=${timestamp}`);
      setSheets(data || []);
    } catch (error) {
      console.error("Failed to fetch sheets:", error);
      setSheets([]);
      toast.error("Failed to load sheets");
    }
    setSheetLoading(false);
  };

  // Fetch sheets included in a notebook
  const fetchNotebookSheets = async (notebook: any) => {
    if (!notebook || !notebook.items) {
      setViewSheets([]);
      return;
    }
    
    try {
      const includedSheets = [];
      for (const [sheetName, url] of Object.entries(notebook.items)) {
        includedSheets.push({ sheetName, url });
      }
      setViewSheets(includedSheets);
    } catch (error) {
      console.error("Error processing notebook sheets:", error);
      setViewSheets([]);
    }
  };

  const handleAddSheet = async (notebookId: number, sheet: any) => {
    try {
      // Extract PDF URL from various possible locations in the sheet object
      const pdfUrl = sheet.result?.pdf_url || sheet.pdf_url || sheet.url || "";
      
      // Extract sheet name from various possible locations
      let sheetName = "";
      
      if (sheet.prompt) {
        try {
          const promptData = typeof sheet.prompt === 'string' 
            ? JSON.parse(sheet.prompt) 
            : sheet.prompt;
          sheetName = promptData.subject || "";
        } catch (e) {
          // If parsing fails, use fallback
          sheetName = sheet.subject || sheet.id || "Untitled Sheet";
        }
      } else {
        sheetName = sheet.subject || sheet.id || "Untitled Sheet";
      }

      if (!sheetName) {
        toast.error("Sheet must have a name");
        return;
      }

      await addSheetToNotebook(notebookId, sheetName, pdfUrl);
      toast.success("Sheet added to notebook!");
      setEditNotebook(null);
      
      // Refresh data after adding
      mutateData();
      
      // If viewing the notebook, update the view
      if (viewNotebook && viewNotebook.id === notebookId) {
        fetchNotebookSheets({
          ...viewNotebook,
          items: {...(viewNotebook.items || {}), [sheetName]: pdfUrl}
        });
      }
    } catch (err) {
      console.error("Failed to add sheet:", err);
      toast.error("Failed to add sheet");
    }
  };

  const handleRemoveSheet = async (notebookId: number, sheetName: string) => {
    try {
      await deleteItemFromNotebook(notebookId, sheetName);
      toast.success("Sheet removed from notebook!");

      // Refresh the current view
      if (viewNotebook && viewNotebook.id === notebookId) {
        const updatedSheets = viewSheets.filter((s) => s.sheetName !== sheetName);
        setViewSheets(updatedSheets);
      }

      // Refresh all data
      mutateData();
    } catch (error) {
      console.error("Failed to remove sheet:", error);
      toast.error("Failed to remove sheet");
    }
  };

  const handleDeleteNotebook = async (notebookId: number) => {
    if (confirm("Are you sure you want to delete this notebook?")) {
      try {
        await deleteNotebook(notebookId);
        toast.success("Notebook deleted!");
        setViewNotebook(null);
        mutateData();
      } catch (error) {
        console.error("Failed to delete notebook:", error);
        toast.error("Failed to delete notebook");
      }
    }
  };

  const startEditingNotebook = (notebook: any) => {
    setEditingNotebook(notebook);
    setEditName(notebook.name || "");
    setEditDescription(notebook.description || "");
    setEditTags((notebook.optional?.tags || []).join(", "));
    setEditColor(notebook.optional?.color || "");
  };

  const handleUpdateNotebook = async () => {
    try {
      await updateNotebook(editingNotebook.id, {
        name: editName,
        description: editDescription,
        tags: editTags.split(",").map((t) => t.trim()).filter(Boolean),
        color: editColor,
      });
      toast.success("Notebook updated!");
      setEditingNotebook(null);
      
      // Refresh all data
      mutateData();

      // Update view if viewing the same notebook
      if (viewNotebook && viewNotebook.id === editingNotebook.id) {
        const updatedNotebook = {
          ...viewNotebook,
          name: editName,
          description: editDescription,
          optional: {
            ...viewNotebook.optional,
            tags: editTags.split(",").map((t) => t.trim()).filter(Boolean),
            color: editColor,
          },
        };
        setViewNotebook(updatedNotebook);
      }
    } catch (error) {
      console.error("Failed to update notebook:", error);
      toast.error("Failed to update notebook");
    }
  };

  return (
    <Card
      className="max-w-full border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0"
      style={{ height: "calc(70vh - 20px)" }}
    >
      <CardContent className="mt-1 px-2 h-full py-2 overflow-auto">
        <div className="flex items-center justify-between mb-4">
          <h1
            className="text-5xl font-extrabold tracking-tight"
            style={{ fontFamily: "'Space Grotesk', sans-serif" }}
          >
            Your notebooks
          </h1>
          <div className="flex items-center gap-4">
            <NeoButton
              color="blue"
              title="Refresh"
              
              className="px-3 py-1 text-sm font-bold"
              onClick={mutateData}
            />
            <CreateNotebook onCreated={mutateData} />
          </div>
        </div>
        {loading ? (
          <div className="flex flex-col items-center justify-center py-8">
            <Loader2 className="h-10 w-10 text-black animate-spin mb-2" />
            <div className="text-lg font-medium text-gray-700">Loading notebooks...</div>
          </div>
        ) : notebooks.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-12">
            <img 
              src="/undraw/empty.svg" 
              alt="No notebooks" 
              className="w-64 h-64 mb-6 opacity-80"
            />
            <h2 className="text-2xl font-bold text-gray-800 mb-2">No Notebooks Found</h2>
            <p className="text-base text-gray-600 text-center max-w-md mb-4">
              You haven't created any notebooks yet. Notebooks help you organize your sheets by subject or topic.
            </p>
            <CreateNotebook onCreated={mutateData} />
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {notebooks.map((nb, idx) => (
              <div
                key={nb.id || idx}
                className="border-2 border-black rounded-lg p-4 bg-gray-50 shadow-[2px_2px_0_0_#000] flex flex-col h-full mb-4 relative cursor-pointer hover:bg-gray-100 transition-colors"
                onClick={() => {
                  setViewNotebook(nb);
                  fetchNotebookSheets(nb);
                }}
              >
                {/* Color indicator based on notebook color */}
                {nb.optional?.color && (
                  <div 
                    className="absolute top-0 left-0 w-full h-2 rounded-t-lg" 
                    style={{ backgroundColor: nb.optional.color }}
                  ></div>
                )}
                

                {/* Edit/Delete buttons for notebook */}
                <div className="absolute top-2 right-2 flex gap-1">
                  <button
                    className="p-1 text-gray-500 hover:text-blue-600 bg-white rounded-full"
                    onClick={(e) => {
                      e.stopPropagation();
                      startEditingNotebook(nb);
                    }}
                    title="Edit Notebook"
                  >
                    <Edit size={16} />
                  </button>
                  <button
                    className="p-1 text-gray-500 hover:text-red-600 bg-white rounded-full"
                    onClick={(e) => {
                      e.stopPropagation();
                      handleDeleteNotebook(nb.id);
                    }}
                    title="Delete Notebook"
                  >
                    <Trash2 size={16} />
                  </button>
                </div>

                <span className="font-bold text-lg truncate mt-4">{nb.name}</span>
                <span className="text-xs text-gray-500 mb-2">{nb.description}</span>
                <span className="text-xs text-gray-400 mb-2">
                  Created: {nb.createdAt ? new Date(nb.createdAt).toLocaleString() : "N/A"}
                </span>

                {nb.optional && nb.optional.tags && nb.optional.tags.length > 0 && (
                  <div className="flex gap-1 flex-wrap mb-2">
                    {nb.optional.tags.map((tag: string) => (
                      <span key={tag} className="bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs">
                        {tag}
                      </span>
                    ))}
                  </div>
                )}

                <span className="text-xs text-gray-600 mb-2">
                  Sheets: {nb.items ? Object.keys(nb.items).length : 0}
                </span>
                
                <div className="flex space-x-2 mt-2">
                  <button
                    className="px-3 py-1 bg-green-600 text-white rounded font-bold flex items-center"
                    onClick={(e) => {
                      e.stopPropagation();
                      setEditNotebook(nb);
                      fetchSheets();
                    }}
                  >
                    <span>Add Sheet</span>
                  </button>
                  
                  <button
                    className="px-3 py-1 bg-blue-600 text-white rounded font-bold flex items-center"
                    onClick={(e) => {
                      e.stopPropagation();
                      setViewNotebook(nb);
                      fetchNotebookSheets(nb);
                    }}
                  >
                    <BookOpen size={16} className="mr-1" />
                    <span>View</span>
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
        
        {/* Modal for adding sheets */}
        <Dialog open={!!editNotebook} onOpenChange={open => !open && setEditNotebook(null)}>
          <DialogContent className="max-w-3xl max-h-[90vh] overflow-hidden flex flex-col">
            <DialogHeader>
              <DialogTitle>Add Sheet to {editNotebook?.name}</DialogTitle>
            </DialogHeader>
            {sheetLoading ? (
              <div className="flex items-center justify-center py-12">
                <Loader2 className="h-8 w-8 text-black animate-spin mb-2" />
                <span className="ml-2">Loading sheets...</span>
              </div>
            ) : sheets.length === 0 ? (
              <div className="text-base font-medium text-gray-600 py-8 text-center">
                No sheets available. Create some sheets first!
              </div>
            ) : (
              <div className="max-h-[60vh] overflow-y-auto p-1">
                {sheets.map((sheet: any) => {
                  let sheetName = "Untitled Sheet";
                  let description = "";
                  
                  // Extract sheet info from various possible structures
                  if (sheet.prompt) {
                    try {
                      const promptData = typeof sheet.prompt === 'string' 
                        ? JSON.parse(sheet.prompt) 
                        : sheet.prompt;
                      sheetName = promptData.subject || sheet.subject || sheet.id || "Untitled Sheet";
                      description = promptData.description || "";
                    } catch (e) {
                      sheetName = sheet.subject || sheet.id || "Untitled Sheet";
                    }
                  } else {
                    sheetName = sheet.subject || sheet.id || "Untitled Sheet";
                  }
                  
                  // Extract status for display
                  const status = sheet.status || "unknown";
                  const isComplete = status === "completed";
                  
                  return (
                    <div 
                      key={sheet.id} 
                      className={`flex items-center justify-between border-b py-3 px-2 ${isComplete ? 'bg-white' : 'bg-gray-50'} hover:bg-gray-100 rounded-md mb-1`}
                    >
                      <div className="flex-1 pr-2">
                        <div className="font-bold line-clamp-1">{sheetName}</div>
                        <div className="text-xs text-gray-500 line-clamp-2 mt-1">
                          {description}
                        </div>
                        <div className="flex items-center mt-1">
                          <span className={`text-xs px-2 py-0.5 rounded ${
                            isComplete ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                          }`}>
                            {status}
                          </span>
                          {sheet.createdAt && (
                            <span className="text-xs text-gray-400 ml-2">
                              {new Date(sheet.createdAt).toLocaleString()}
                            </span>
                          )}
                        </div>
                      </div>
                      <NeoButton
                        color="green"
                        title="Add"
                        className="px-4 py-1 text-sm font-bold"
                        disabled={!isComplete}
                        onClick={() => {
                          if (isComplete) {
                            handleAddSheet(editNotebook.id, sheet);
                          } else {
                            toast.error("Can only add completed sheets");
                          }
                        }}
                      />
                    </div>
                  );
                })}
              </div>
            )}
            <DialogFooter className="pt-4 border-t mt-2">
              <DialogClose asChild>
                <NeoButton color="red" title="Close" />
              </DialogClose>
            </DialogFooter>
          </DialogContent>
        </Dialog>

        {/* Modal for editing notebook details */}
        <Dialog open={!!editingNotebook} onOpenChange={open => !open && setEditingNotebook(null)}>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Edit Notebook</DialogTitle>
            </DialogHeader>
            <div className="space-y-3 pt-2">
              <div>
                <label className="block text-sm font-medium mb-1">Name</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border-2 border-black rounded-lg"
                  value={editName}
                  onChange={(e) => setEditName(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Description</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border-2 border-black rounded-lg"
                  value={editDescription}
                  onChange={(e) => setEditDescription(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Tags (comma separated)</label>
                <input
                  type="text"
                  className="w-full px-3 py-2 border-2 border-black rounded-lg"
                  value={editTags}
                  onChange={(e) => setEditTags(e.target.value)}
                />
              </div>
              <div>
                <label className="block text-sm font-medium mb-1">Color</label>
                <div className="flex space-x-2 items-center">
                  <input
                    type="text"
                    className="flex-1 px-3 py-2 border-2 border-black rounded-lg"
                    value={editColor}
                    onChange={(e) => setEditColor(e.target.value)}
                    placeholder="e.g., #ff0000 or red"
                  />
                  <input
                    type="color"
                    value={editColor.startsWith('#') ? editColor : '#000000'}
                    onChange={(e) => setEditColor(e.target.value)}
                    className="h-10 w-10 cursor-pointer border-2 border-black rounded"
                  />
                </div>
              </div>
              <DialogFooter className="pt-3">
                <NeoButton
                  color="blue"
                  title="Save Changes"
                  onClick={handleUpdateNotebook}
                />
                <DialogClose asChild>
                  <NeoButton color="red" title="Cancel" />
                </DialogClose>
              </DialogFooter>
            </div>
          </DialogContent>
        </Dialog>

        {/* Modal for viewing notebook sheets */}
        <Dialog open={!!viewNotebook} onOpenChange={open => !open && setViewNotebook(null)}>
          <DialogContent className="max-w-5xl max-h-[90vh] overflow-hidden flex flex-col">
            <DialogHeader>
              <DialogTitle>
                <span className="flex items-center">
                  {viewNotebook?.optional?.color && (
                    <span 
                      className="inline-block w-4 h-4 mr-2 rounded-full" 
                      style={{ backgroundColor: viewNotebook.optional.color }}
                    ></span>
                  )}
                  {viewNotebook?.name}
                </span>
              </DialogTitle>
              <DialogDescription>{viewNotebook?.description}</DialogDescription>
            </DialogHeader>
            
            <div className="flex gap-2 mb-4">
              <NeoButton
                color="green"
                title="Add Sheet"
                className="text-xs px-3 py-1"
                onClick={() => {
                  setEditNotebook(viewNotebook);
                  fetchSheets();
                }}
              />
              <NeoButton
                color="blue"
                title="Edit Notebook"
                className="text-xs px-3 py-1"
                onClick={() => startEditingNotebook(viewNotebook)}
              />
              <NeoButton
                color="yellow"
                title="Refresh"
               
                className="text-xs px-3 py-1"
                onClick={() => {
                  if (viewNotebook) {
                    fetchNotebookSheets(viewNotebook);
                    mutateData();
                  }
                }}
              />
            </div>
            
            {viewSheets.length === 0 ? (
              <div className="text-base font-medium text-gray-600 py-10 text-center">
                No sheets in this notebook.
                <div className="mt-2">
                  <NeoButton
                    color="green"
                    title="Add Sheets"
                    onClick={() => {
                      setEditNotebook(viewNotebook);
                      fetchSheets();
                    }}
                  />
                </div>
              </div>
            ) : (
              <div className="max-h-[60vh] overflow-y-auto">
                <div className="grid grid-cols-1 gap-6">
                  {viewSheets.map((sheet: any, idx: number) => (
                    <div key={idx} className="border-2 border-black rounded-lg p-4 bg-gray-50 relative">
                      <button 
                        className="absolute top-2 right-2 text-gray-400 hover:text-red-600 bg-white p-1 rounded-full shadow-sm"
                        onClick={() => handleRemoveSheet(viewNotebook.id, sheet.sheetName)}
                        title="Remove Sheet"
                      >
                        <X size={16} />
                      </button>
                      
                      <h3 className="font-bold text-lg mb-3 pr-8">{sheet.sheetName}</h3>
                      
                      {sheet.url && (
                        <div className="mb-3 border border-gray-300 rounded-lg overflow-hidden">
                          <iframe
                            src={sheet.url}
                            title={`Sheet-${sheet.sheetName}`}
                            className="w-full h-[500px] border-0"
                            sandbox="allow-same-origin allow-scripts"
                            loading="lazy"
                          />
                        </div>
                      )}
                      
                      <div className="flex justify-end">
                        <NeoButton
                          color="blue"
                          title="View Full Sheet"
                          className="text-sm"
                          onClick={() => window.open(sheet.url, "_blank")}
                        />
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
            
            <DialogFooter className="pt-4 border-t mt-2">
              <DialogClose asChild>
                <NeoButton color="red" title="Close" />
              </DialogClose>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </CardContent>
    </Card>
  );
}