import { ArrowRight, SheetIcon, BookOpen, Palette } from "lucide-react";
import { Link } from "react-router-dom";

interface QuickAction {
  title: string;
  description: string;
  icon: React.ReactNode;
  path: string;
  color: string;
  borderColor: string;
  shadowColor: string;
  iconBg: string;
}

const actions: QuickAction[] = [
  {
    title: "New Sheet",
    description: "Generate AI-powered worksheets with custom styling",
    icon: <SheetIcon className="h-5 w-5" />,
    path: "/sheets",
    color: "bg-blue-100 hover:bg-blue-150",
    borderColor: "border-blue-600",
    shadowColor: "shadow-[3px_3px_0_0_#2563eb]",
    iconBg: "bg-blue-500",
  },
  {
    title: "Notebooks",
    description: "Organize notes, resources, and study material",
    icon: <BookOpen className="h-5 w-5" />,
    path: "/notebooks",
    color: "bg-emerald-100 hover:bg-emerald-150",
    borderColor: "border-emerald-600",
    shadowColor: "shadow-[3px_3px_0_0_#059669]",
    iconBg: "bg-emerald-500",
  },
  {
    title: "Designer",
    description: "Customize templates and visual styles",
    icon: <Palette className="h-5 w-5" />,
    path: "/designer",
    color: "bg-violet-100 hover:bg-violet-150",
    borderColor: "border-violet-600",
    shadowColor: "shadow-[3px_3px_0_0_#7c3aed]",
    iconBg: "bg-violet-500",
  },
];

export default function QuickActions() {
  return (
    <div className="w-full">
      <h2 className="text-xl font-black text-black tracking-tight mb-2">
        Quick Actions
      </h2>
      <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
        {actions.map((action) => (
          <Link key={action.title} to={action.path} className="no-underline">
            <div
              className={`group relative ${action.color} border-3 ${action.borderColor} rounded-xl shadow-[3px_3px_0_0_#000] p-4 cursor-pointer transition-all duration-200 hover:shadow-[5px_5px_0_0_#000] hover:-translate-y-0.5 active:shadow-[0px_0px_0px_0px] active:translate-x-[3px] active:translate-y-[3px] h-full`}
            >
              {/* Icon */}
              <div
                className={`${action.iconBg} text-white w-10 h-10 rounded-xl border-2 border-black shadow-[2px_2px_0px_0px_#000] flex items-center justify-center mb-3`}
              >
                {action.icon}
              </div>

              {/* Text */}
              <h3 className="text-lg font-extrabold text-black tracking-tight mb-1">
                {action.title}
              </h3>
              <p className="text-xs font-medium text-gray-700 leading-snug opacity-70">
                {action.description}
              </p>

              {/* Arrow indicator */}
              <div className="flex items-center gap-1 text-black font-bold text-xs mt-3">
                <span>Open</span>
                <ArrowRight className="h-3 w-3 transition-transform group-hover:translate-x-1 duration-200" />
              </div>
            </div>
          </Link>
        ))}
      </div>
    </div>
  );
}
