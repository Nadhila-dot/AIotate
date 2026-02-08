import http from "@/http";

export interface SheetCreateData {
  subject: string;
  course: string;
  description: string;
  tags: string;
  curriculum: string;
  specialInstructions: string;
  visibility: string;
  styleName?: string;
  mode: string;
  webSearchQuery?: string;
  webSearchEnabled?: boolean;
}

export async function createSheet(data: SheetCreateData, files?: File[]) {
  if (files && files.length > 0) {
    const form = new FormData();
    Object.entries(data).forEach(([key, value]) => {
      if (value !== undefined && value !== null) {
        form.append(key, String(value));
      }
    });
    files.forEach((file) => form.append("files", file));

    const res = await http.post("/api/v1/sheets/create", form, {
      headers: { "Content-Type": "multipart/form-data" },
    });
    return res.data;
  }

  const res = await http.post("/api/v1/sheets/create", data);
  return res.data;
}

export async function getSheetQueue() {
  const res = await http.get("/api/v1/sheets/queue");
  return res.data;
}