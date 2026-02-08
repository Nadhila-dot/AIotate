import PageBlock from "@/components/Content/PageBlock";
import { HomeIcon } from "lucide-react";
import SheetListCard from "@/components/Cards/Sheets/SheetList";
import HeroBanner from "@/components/Content/HeroBanner";
import QuickActions from "@/components/Content/QuickActions";


export function HomeContainer() {
  return (
    <>
      <PageBlock header="Home" icon={<HomeIcon size={56} />}>
        <div className="flex flex-col gap-4">
          <HeroBanner />
          <QuickActions />
          <SheetListCard />
        </div>
      </PageBlock>
    </>
  );
}