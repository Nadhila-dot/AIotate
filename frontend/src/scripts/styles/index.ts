import http from "@/http";

export interface StyleItem {
  name: string;
  username: string;
  prompt: string;
  description: string;
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export async function listStyles() {
  const res = await http.get("/api/v1/styles");
  return res.data as StyleItem[];
}

export async function getDefaultStyle() {
  const res = await http.get("/api/v1/styles/default");
  return res.data as StyleItem;
}

export async function createStyle(data: { name: string; prompt: string; description?: string; isDefault?: boolean }) {
  const res = await http.post("/api/v1/styles", data);
  return res.data as StyleItem;
}

export async function updateStyle(name: string, data: { prompt: string; description?: string; isDefault?: boolean }) {
  const res = await http.put(`/api/v1/styles/${encodeURIComponent(name)}`, data);
  return res.data as StyleItem;
}

export async function deleteStyle(name: string) {
  const res = await http.delete(`/api/v1/styles/${encodeURIComponent(name)}`);
  return res.data;
}

export async function setDefaultStyle(name: string) {
  const res = await http.post(`/api/v1/styles/${encodeURIComponent(name)}/default`);
  return res.data as StyleItem;
}
