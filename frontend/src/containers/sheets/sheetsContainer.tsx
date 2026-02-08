import PageBlock from "@/components/Content/PageBlock";
import LogoutCard from "@/components/Cards/LogoutCard";
import { Button } from "@/components/ui/button";
import { Settings2Icon, SettingsIcon, SheetIcon } from "lucide-react";
import SetCard from "@/components/Cards/Setcard";
import CreateSheet from "@/components/Cards/Sheets/SheetCreate";
import { useSearchParams } from "react-router-dom";
import { FormData } from "@/hooks/forms/useFormGeneration";


export function SheetsContainer() {
  const [searchParams] = useSearchParams();

  const initialData: Partial<FormData> = {
    subject: searchParams.get("subject") || "",
    course: searchParams.get("course") || "",
    description: searchParams.get("description") || "",
    tags: searchParams.get("tags") || "",
    curriculum: searchParams.get("curriculum") || "",
    specialInstructions: searchParams.get("specialInstructions") || "",
    styleName: searchParams.get("styleName") || "",
    mode: searchParams.get("mode") || "notes",
    webSearchQuery: searchParams.get("webSearchQuery") || "",
    webSearchEnabled: (searchParams.get("webSearchEnabled") || "").toLowerCase() === "true",
  };


  return (
    <PageBlock header="Sheets" icon={<SheetIcon size={56} />}>
       <CreateSheet initialData={initialData} />
    </PageBlock>
  );
}