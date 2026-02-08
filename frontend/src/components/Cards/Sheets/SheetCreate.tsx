import React, { useState, KeyboardEvent, useEffect } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { NeoButton } from "@/components/ui/neo-button";
import { FormField } from "@/components/forms/Field";
import { VisibilitySelector } from "@/components/forms/VisibilitySelector";
import { useFormGeneration, FormData } from "@/hooks/forms/useFormGeneration";
import { useTips } from "@/hooks/forms/useTips";
import { SimpleFormField } from "@/components/forms/SimpleField";
import { toast } from "sonner";
import { createSheet } from "@/scripts/sheets";
import { BookOpen, ClipboardList, Zap } from "lucide-react";

const MODES = [
  {
    id: "notes",
    label: "Notes",
    icon: BookOpen,
    color: "bg-blue-100 border-blue-500 text-blue-800",
    activeColor: "bg-blue-500 text-white border-blue-700",
    description: "Comprehensive 3+ page study notes with professional design, definitions, examples, and summaries.",
  },
  {
    id: "prep-test",
    label: "Prep Test",
    icon: ClipboardList,
    color: "bg-orange-100 border-orange-500 text-orange-800",
    activeColor: "bg-orange-500 text-white border-orange-700",
    description: "Full practice exam with mixed question types, answer key, and proper test formatting.",
  },
  {
    id: "super-lazy",
    label: "Super Lazy",
    icon: Zap,
    color: "bg-purple-100 border-purple-500 text-purple-800",
    activeColor: "bg-purple-500 text-white border-purple-700",
    description: "Memory-optimized study guide using key points, mnemonics, and cheat sheets. Read it once, pass the exam.",
  },
] as const;

type CreateSheetProps = {
  initialData?: Partial<FormData>;
};

export default function CreateSheet({ initialData }: CreateSheetProps) {
  const [formData, setFormData] = useState<FormData>({
    subject: "",
    course: "",
    description: "",
    tags: "",
    curriculum: "",
    specialInstructions: "",
    visibility: "private",
    mode: "notes",
    webSearchQuery: "",
    webSearchEnabled: false,
  });

  const [files, setFiles] = useState<File[]>([]);

  const {
    loadingState,
    generateSubject,
    generateCourse,
    generateDescription,
    generateTags,
    generateMissingFields,
  } = useFormGeneration(formData, setFormData);

  const { showTips } = useTips(formData);

  useEffect(() => {
    if (!initialData) return;
    setFormData((prev) => ({
      ...prev,
      ...initialData,
    }));
  }, [initialData]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleToggleWebSearch = () => {
    setFormData(prev => ({ ...prev, webSearchEnabled: !prev.webSearchEnabled }));
  };

  const addFiles = (incoming: FileList | File[]) => {
    const list = Array.from(incoming);
    const next = [...files, ...list];
    const totalSize = next.reduce((sum, file) => sum + file.size, 0);
    const maxSize = 20 * 1024 * 1024;
    if (totalSize > maxSize) {
      toast.error("Upload limit is 20MB total.");
      return;
    }
    setFiles(next);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      addFiles(e.target.files);
      e.target.value = "";
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      addFiles(e.dataTransfer.files);
      e.dataTransfer.clearData();
    }
  };

  const removeFile = (index: number) => {
    setFiles(prev => prev.filter((_, i) => i !== index));
  };

  const handleKeyDown = (e: KeyboardEvent<HTMLInputElement | HTMLTextAreaElement>, field: string) => {
    if (e.key === 'Tab' && !e.shiftKey) {
      const fieldActions = {
        subject: () => showTips.subject && generateSubject(),
        course: () => showTips.course && generateCourse(),
        description: () => showTips.description && generateDescription(),
        tags: () => showTips.tags && generateTags(),
      };

      const action = fieldActions[field as keyof typeof fieldActions];
      //@ts-ignore
      if (action && action()) {
        e.preventDefault();
      }
    }
  };

  const handleCreateSheet = async () => {
  console.log("Create sheet with data:", formData);

  toast.dismiss();
  try {
    const result = await createSheet(formData, files); // Call the API to create the sheet
    console.log(result);

    if (result.jobId) {
      toast.success(`Sheet created successfully! Redirecting to status page...`);
      // Redirect to the status page with the job ID
      setTimeout(() => {
        window.location.href = `/sheets/status?jobId=${result.jobId}`;
      }, 2000); // Add a slight delay for the toast to display
    } else {
      toast.error("Failed to retrieve job ID.");
    }
  } catch (error) {
    toast.error("Failed to create sheet.");
    console.error(error);
  }
};

  return (
    <Card className="max-w-full border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
      <CardContent className="mt-1 px-4 py-4">
        <h1 className="text-5xl font-extrabold tracking-tight mb-2" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
          Create a new sheet
        </h1>
        <div className="mb-6 text-base font-medium">
          A sheet is the simplest form of a data structure, It's a blank thought, waiting to be filled with your new paper / idea.
        </div>

        {/* Mode Selector */}
        <div className="mb-6">
          <label className="block font-bold mb-2 text-xl">Generation Mode</label>
          <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
            {MODES.map((mode) => {
              const isActive = formData.mode === mode.id;
              const Icon = mode.icon;
              return (
                <button
                  key={mode.id}
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, mode: mode.id }))}
                  className={`relative flex flex-col items-start gap-2 p-4 rounded-xl border-3 border-black transition-all duration-200 text-left shadow-[3px_3px_0_0_#000] hover:shadow-[5px_5px_0_0_#000] hover:-translate-y-0.5 ${
                    isActive
                      ? mode.activeColor + " ring-2 ring-black ring-offset-2"
                      : mode.color + " hover:brightness-95"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Icon className="w-5 h-5" strokeWidth={2.5} />
                    <span className="font-extrabold text-lg">{mode.label}</span>
                  </div>
                  <span className={`text-xs leading-snug ${
                    isActive ? "opacity-90" : "opacity-70"
                  }`}>
                    {mode.description}
                  </span>
                  {isActive && (
                    <div className="absolute top-2 right-2 w-3 h-3 rounded-full bg-white border-2 border-current" />
                  )}
                </button>
              );
            })}
          </div>
        </div>

        {/* Web Search */}
        <div className="mb-6 border-2 border-black rounded-xl p-4 bg-gray-50 shadow-[3px_3px_0_0_#000]">
          <div className="flex items-center justify-between mb-3">
            <label className="font-bold text-lg">Web Search</label>
            <button
              type="button"
              onClick={handleToggleWebSearch}
              className={`px-3 py-1 rounded-full border-2 border-black font-bold text-xs ${
                formData.webSearchEnabled ? "bg-green-500 text-white" : "bg-white text-black"
              }`}
            >
              {formData.webSearchEnabled ? "Enabled" : "Disabled"}
            </button>
          </div>
          <input
            type="text"
            name="webSearchQuery"
            value={formData.webSearchQuery}
            onChange={handleChange}
            placeholder="Search query (e.g., 'Photosynthesis key concepts and diagrams')"
            className="w-full border-2 border-black rounded-lg px-3 py-2 text-sm shadow-[2px_2px_0_0_#000] focus:outline-none"
          />
          <p className="text-xs text-gray-600 mt-2">
            When enabled, Vela will pull context from top web sources and include it in the generation.
          </p>
        </div>

        {/* Attachments */}
        <div
          className="mb-6 border-2 border-dashed border-black rounded-xl p-4 bg-white shadow-[3px_3px_0_0_#000]"
          onDragOver={(e) => e.preventDefault()}
          onDrop={handleDrop}
        >
          <div className="flex items-center justify-between mb-2">
            <label className="font-bold text-lg">Attach Files (max 20MB)</label>
            <label className="px-3 py-1 bg-black text-white rounded-md cursor-pointer text-xs font-bold">
              Add Files
              <input
                type="file"
                multiple
                className="hidden"
                onChange={handleFileChange}
              />
            </label>
          </div>
          <p className="text-xs text-gray-600 mb-3">Drop files here or click "Add Files". Text files are parsed; other formats are sent as raw text if supported.</p>

          {files.length > 0 ? (
            <ul className="space-y-2">
              {files.map((file, idx) => (
                <li key={`${file.name}-${idx}`} className="flex items-center justify-between bg-gray-50 border-2 border-black rounded-lg px-3 py-2">
                  <div>
                    <div className="font-semibold text-sm">{file.name}</div>
                    <div className="text-xs text-gray-600">{(file.size / 1024).toFixed(1)} KB</div>
                  </div>
                  <button
                    type="button"
                    onClick={() => removeFile(idx)}
                    className="text-xs font-bold text-red-600"
                  >
                    Remove
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            <div className="text-sm text-gray-500">No files added yet.</div>
          )}
        </div>

        <div className="space-y-5">
          <FormField
            label="Subject"
            name="subject"
            value={formData.subject}
            placeholder="e.g. Mathematics, Computer Science, Physics"
            onChange={handleChange}
            onKeyDown={(e) => handleKeyDown(e, 'subject')}
            isLoading={loadingState.subject}
            showTip={showTips.subject}
            tipText="Press Tab"
            tipAction="generate subject"
            onTipClick={generateSubject}
            isTextarea={true}
            rows={1}
          />

          <FormField
            label="Course"
            name="course"
            value={formData.course}
            placeholder="e.g. Calculus 101, Introduction to AI"
            onChange={handleChange}
            onKeyDown={(e) => handleKeyDown(e, 'course')}
            isLoading={loadingState.course}
            showTip={showTips.course}
            tipText="Press Tab"
            tipAction="generate course"
            onTipClick={generateCourse}
            isTextarea={true}
            rows={1}
          />

          <FormField
            label="Description"
            name="description"
            value={formData.description}
            placeholder="What's this sheet about? Add some context here..."
            onChange={handleChange}
            onKeyDown={(e) => handleKeyDown(e, 'description')}
            isLoading={loadingState.description}
            showTip={showTips.description}
            tipText="Press Tab"
            tipAction="generate description"
            onTipClick={generateDescription}
            isTextarea={true}
            rows={3}
          />

          <FormField
            label="Tags"
            name="tags"
            value={formData.tags}
            placeholder="Separate tags with commas"
            onChange={handleChange}
            onKeyDown={(e) => handleKeyDown(e, 'tags')}
            isLoading={loadingState.tags}
            showTip={showTips.tags}
            tipText="Press Tab"
            tipAction="generate tags"
            onTipClick={generateTags}
            rows={1}
            isTextarea={true}
          />

          <h1 className="block font-bold mb-1 text-3xl">
            Optional
          </h1>

          <SimpleFormField
            label="Guidence / Curriculum"
            name="curriculum"
            value={formData.curriculum}
            placeholder="Tell vela more about the curicullum your following and what it needs to base it's information off of."
            onChange={handleChange} // change handler
            rows={2}
          />

          <SimpleFormField
            label="Special Instructions"
            name="specialInstructions"
            value={formData.specialInstructions}
            placeholder="Any special instructions for vela to follow when creating this sheet?"
            onChange={handleChange} // change handler
            rows={2}
          />

          <VisibilitySelector
            value={formData.visibility}
            onChange={handleChange}
          />
        </div>

        <div className="mt-8 flex gap-4">
          <NeoButton
            color="green"
            title="Fill Missing Fields"
            onClick={generateMissingFields}
            className="mr-4"
          />
          <NeoButton
            color="red"
            title="Create Sheet"
            onClick={handleCreateSheet}
          />
        </div>
      </CardContent>
    </Card>
  );
}