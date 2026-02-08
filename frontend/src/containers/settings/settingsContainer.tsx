import PageBlock from "@/components/Content/PageBlock";
import LogoutCard from "@/components/Cards/LogoutCard";
import { SettingsIcon } from "lucide-react";
import SetCard from "@/components/Cards/Setcard";
import RefreshCache from "@/components/Cards/Education/RefershCache";
import SystemInfoCard from "@/components/Cards/SystemInfo";
import AIConfigEditor from "@/components/Cards/AIConfigEditor";

export function SettingsContainer() {
  return (
    <PageBlock header="Settings" icon={<SettingsIcon size={56} />}>
      <div className="space-y-4">
        <div className="max-w-3xl">
          <AIConfigEditor />
        </div>
        
       {/* <div className="max-w-3xl">
          <SetCard />
        </div>*/}
        
       <div className="max-w-3xl">
          <SystemInfoCard />
        </div>
        
        <div className="max-w-3xl">
          <RefreshCache />
        </div>
        
        <div className="max-w-3xl">
          <LogoutCard />
        </div>
      </div>
    </PageBlock>
  );
}