import { useEffect, useMemo, useState } from "react";
import PageBlock from "@/components/Content/PageBlock";
import { Card, CardContent } from "@/components/ui/card";
import { NeoButton } from "@/components/ui/neo-button";
import { SimpleFormField } from "@/components/forms/SimpleField";
import { Palette } from "lucide-react";
import { toast } from "sonner";
import { createStyle, deleteStyle, listStyles, setDefaultStyle, updateStyle, StyleItem } from "@/scripts/styles";
import { createSheet } from "@/scripts/sheets";
import http from "@/http";

export function DesignerContainer() {
  const defaultStylePrompt = `Document styling for visual appeal
\\definecolor{primary}{RGB}{25,103,210}
\\definecolor{secondary}{RGB}{234,67,53}
\\definecolor{accent}{RGB}{251,188,4}
\\definecolor{light}{RGB}{242,242,242}

\\hypersetup{colorlinks=true,linkcolor=primary}
\\setlength{\\parindent}{0pt}
\\setlength{\\parskip}{6pt}

\\title{\\textcolor{primary}{\\Large TITLE_OF_WORKSHEET}}
\\author{\\textcolor{secondary}{Course: COURSE_NAME}}
\\date{\\today}`;


  const [styles, setStyles] = useState<StyleItem[]>([]);
  const [selectedName, setSelectedName] = useState<string>("");
  const [name, setName] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [prompt, setPrompt] = useState<string>(defaultStylePrompt);
  const [loading, setLoading] = useState(false);
  const [testing, setTesting] = useState(false);
  const [previewHtml, setPreviewHtml] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState<string | null>(null);

  const previewLatex = useMemo(() => (
    `\\textcolor{primary}{Primary Accent}\\\\
\\textcolor{secondary}{Secondary Accent}\\\\
\\textcolor{accent}{Accent}\\\\
\\textcolor{light}{Light Accent}`
  ), []);

  const buildPreviewDocument = (stylePrompt: string) => `\\documentclass[11pt]{article}
\\usepackage{xcolor}
\\usepackage{hyperref}
\\usepackage{geometry}
\\usepackage{graphicx}
\\geometry{margin=1in}

${stylePrompt}

\\begin{document}
\\section*{Style Preview}
This preview uses your current style prompt to render a sample layout.\\\\
${previewLatex}

\\vspace{12pt}
\\fcolorbox{primary}{light}{\\parbox{0.88\\linewidth}{\\centering
\\textbf{Sample callout}\\\\
Use this block to verify your primary/secondary palette, spacing, and typography.
}}

\\end{document}`;

  const selectedStyle = useMemo(() => styles.find((s) => s.name === selectedName), [styles, selectedName]);

  const loadStyles = async () => {
    try {
      const data = await listStyles();
      setStyles(data);
      const defaultStyle = data.find((s) => s.isDefault);
      if (defaultStyle) {
        setSelectedName(defaultStyle.name);
      }
    } catch (error) {
      toast.error("Failed to load styles");
    }
  };

  useEffect(() => {
    loadStyles();
  }, []);

  useEffect(() => {
    if (selectedStyle) {
      setName(selectedStyle.name);
      setDescription(selectedStyle.description || "");
      setPrompt(selectedStyle.prompt || "");
    } else if (!selectedName) {
      setPrompt(defaultStylePrompt);
    }
  }, [selectedStyle]);

  useEffect(() => {
    setPreviewError(null);
  }, [prompt]);

  const handleNewStyle = () => {
    setSelectedName("");
    setName("");
    setDescription("");
    setPrompt(defaultStylePrompt);
  };

  const handleSave = async () => {
    if (!name.trim() || !prompt.trim()) {
      toast.error("Name and prompt are required");
      return;
    }

    setLoading(true);
    try {
      const exists = styles.some((s) => s.name === name.trim());
      if (exists) {
        await updateStyle(name.trim(), { prompt: prompt.trim(), description: description.trim() });
        toast.success("Style updated");
      } else {
        await createStyle({ name: name.trim(), prompt: prompt.trim(), description: description.trim(), isDefault: false });
        toast.success("Style created");
      }
      await loadStyles();
      setSelectedName(name.trim());
    } catch (error) {
      toast.error("Failed to save style");
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!selectedName) {
      toast.error("Select a style to delete");
      return;
    }

    setLoading(true);
    try {
      await deleteStyle(selectedName);
      toast.success("Style deleted");
      handleNewStyle();
      await loadStyles();
    } catch (error) {
      toast.error("Failed to delete style");
    } finally {
      setLoading(false);
    }
  };

  const handleSetDefault = async () => {
    if (!name.trim()) {
      toast.error("Select a style to set as default");
      return;
    }

    setLoading(true);
    try {
      await setDefaultStyle(name.trim());
      toast.success("Default style updated");
      await loadStyles();
    } catch (error) {
      toast.error("Failed to set default style");
    } finally {
      setLoading(false);
    }
  };

  const handleQuickGenerate = async () => {
    if (!name.trim()) {
      toast.error("Select a style to test");
      return;
    }

    setTesting(true);
    try {
      const result = await createSheet({
        subject: "Algebra",
        course: "Linear Equations",
        description: "A short diagnostic worksheet on solving linear equations.",
        tags: "algebra, equations, linear",
        curriculum: "Solving and checking linear equations",
        specialInstructions: "Keep it concise and include 5 questions.",
        visibility: "private",
        styleName: name.trim(),
      });

      if (result.jobId) {
        toast.success("Quick test started. Redirecting to status page...");
        setTimeout(() => {
          window.location.href = `/sheets/status?jobId=${result.jobId}`;
        }, 1500);
      } else {
        toast.error("Failed to retrieve job ID");
      }
    } catch (error) {
      toast.error("Failed to start quick generation");
    } finally {
      setTesting(false);
    }
  };

  const handlePreview = async () => {
    setPreviewLoading(true);
    setPreviewError(null);

    try {
      const latexDoc = (prompt.trim() || defaultStylePrompt);
      const res = await http.post("/api/v1/latex/preview", { latex: latexDoc }, {
        responseType: "text",
        headers: { "x-cache-bypass": "true" },
      });
      setPreviewHtml(res.data as string);
    } catch (error: any) {
      setPreviewError("Failed to compile preview.");
      toast.error("Failed to compile preview");
    } finally {
      setPreviewLoading(false);
    }
  };

  return (
    <PageBlock header="Designer" icon={<Palette size={56} />}>
      <Card className="max-w-full border-4 border-black rounded-xl shadow-[8px_8px_0_0_#000] bg-white text-black p-0">
        <CardContent className="mt-1 px-4 py-4">
          <h1 className="text-4xl font-extrabold tracking-tight mb-2" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
            Design System
          </h1>
          <p className="mb-6 text-base font-medium">
            Create a new design for your sheets, This design is simply a prompt that commands the ai to use specific styles. Use latex to define your styles simply
          </p>

          <div className="grid grid-cols-1 lg:grid-cols-[260px_1fr_420px] gap-6">
            <div>
              <label className="block font-bold mb-1 text-black text-lg">Your Styles</label>
              <select
                value={selectedName}
                onChange={(e) => setSelectedName(e.target.value)}
                className="w-full border-2 text-black border-black rounded-lg px-3 py-2 shadow-[2px_2px_0_0_#000] focus:outline-none focus:ring-2 focus:ring-black"
              >
                <option value="">Select a style</option>
                {styles.map((style) => (
                  <option key={style.name} value={style.name}>
                    {style.name}{style.isDefault ? " (default)" : ""}
                  </option>
                ))}
              </select>

              <div className="mt-4 flex flex-col gap-3">
                <NeoButton color="blue" title="New Style" onClick={handleNewStyle} />
                <NeoButton color="green" title={loading ? "Saving..." : "Save Style"} onClick={handleSave} />
                <NeoButton color="yellow" title={loading ? "Working..." : "Set Default"} onClick={handleSetDefault} />
                <NeoButton color="red" title={loading ? "Deleting..." : "Delete Style"} onClick={handleDelete} />
                <NeoButton color="purple" title={testing ? "Testing..." : "Quick Generate"} onClick={handleQuickGenerate} />
              </div>
            </div>

            <div className="space-y-4">
              <SimpleFormField
                label="Style Name"
                name="styleName"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Clean Academic, Retro Neon"
                isTextarea={false}
              />

              <SimpleFormField
                label="Description"
                name="styleDescription"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Short description of what this style does"
                rows={2}
              />

              <SimpleFormField
                label="Style Prompt (LaTeX styling block)"
                name="stylePrompt"
                value={prompt}
                onChange={(e) => setPrompt(e.target.value)}
                placeholder="Paste LaTeX styling block here (colors, title, spacing, etc.)"
                rows={12}
              />

              <div className="flex flex-wrap items-center gap-3">
                <NeoButton
                  color="blue"
                  title={previewLoading ? "Compiling..." : "Compile & Preview"}
                  onClick={handlePreview}
                />
                <span className="text-xs text-gray-600">
                  Uses Tectonic to compile and render a PNG preview.
                </span>
              </div>
            </div>

            <div className="border-2 border-black rounded-xl px-4 py-4 shadow-[3px_3px_0_0_#000] bg-white">
              <div className="flex items-center justify-between mb-3">
                <div>
                  <div className="text-lg font-bold">Preview</div>
                  <div className="text-xs text-gray-600">Rendered output from Tectonic</div>
                </div>
                {previewLoading && (
                  <span className="text-xs font-semibold text-blue-700">Compiling…</span>
                )}
              </div>

              <div className="min-h-[280px] rounded-lg border-2 border-dashed border-black bg-gray-50 flex items-center justify-center p-3">
                {previewHtml ? (
                  <iframe
                    title="LaTeX preview"
                    className="w-full h-[520px] rounded-md bg-white"
                    sandbox="allow-same-origin"
                    srcDoc={previewHtml}
                  />
                ) : (
                  <div className="text-center text-sm text-gray-600">
                    {previewError ? previewError : "Click Compile & Preview to render a sample sheet."}
                  </div>
                )}
              </div>
            </div>
          </div>
        </CardContent>
      </Card>
    </PageBlock>
  );
}
