import { useEffect, useState } from "react";
import { Crown } from "lucide-react";
import { useAuth } from "@/hooks/auth/checkAuth";
import { useSession } from "@/hooks/auth/getSession";
import { getUserInfo } from "@/scripts/getUserinfo";
import { Link } from "react-router-dom";

type UserInfo = {
  username: string;
  email: string;
  rank: string;
  avatar?: string | null;
};

const greetings = [
  { time: "morning", phrases: ["Good morning", "Morning", "Top of the morning"] },
  { time: "afternoon", phrases: ["Good afternoon", "Afternoon", "Lovely afternoon"] },
  { time: "evening", phrases: ["Good evening", "Evening", "Superb evening innit?", "Fine evening"] },
  { time: "night", phrases: ["Good night", "Late night study session?", "Burning the midnight oil?", "Night owl mode"] },
];

function getGreeting(): string {
  const hour = new Date().getHours();
  let timeOfDay: "morning" | "afternoon" | "evening" | "night";
  
  if (hour < 12) timeOfDay = "morning";
  else if (hour < 17) timeOfDay = "afternoon";
  else if (hour < 22) timeOfDay = "evening";
  else timeOfDay = "night";

  const greetingSet = greetings.find(g => g.time === timeOfDay);
  const phrases = greetingSet?.phrases || ["Hello"];
  return phrases[Math.floor(Math.random() * phrases.length)];
}

function getRankStyles(rank: string) {
  switch (rank.toLowerCase()) {
    case "admin":
      return "bg-red-500 text-white";
    case "moderator":
      return "bg-yellow-400 text-black";
    default:
      return "bg-emerald-400 text-black";
  }
}

export default function HeroBanner() {
  const isAuthenticated = useAuth();
  const session = useSession();
  const [user, setUser] = useState<UserInfo | null>(null);
  const [currentTime, setCurrentTime] = useState("");
  const [greeting] = useState(() => getGreeting());

  useEffect(() => {
    if (session) {
      getUserInfo(session).then((data) => setUser(data));
    }
  }, [session]);

  useEffect(() => {
    const updateTime = () => {
      const now = new Date();
      setCurrentTime(
        now.toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        })
      );
    };
    updateTime();
    const interval = setInterval(updateTime, 1000);
    return () => clearInterval(interval);
  }, []);

  if (!user) {
    return (
      <div className="w-full h-48 bg-gray-100 border-4 border-black rounded-lg shadow-[8px_8px_0px_0px_#000] animate-pulse" />
    );
  }

  return (
    <div className="w-full border-4 border-black bg-white rounded-lg shadow-[8px_8px_0px_0px_#000]">
      <div className="p-4 md:p-6">
        {/* Time - Large and prominent */}
        <div className="mb-3">
          <div className="text-4xl md:text-5xl font-black text-black tracking-tight leading-none">
            {currentTime}
          </div>
        </div>

        {/* Greeting */}
        <div className="mb-4">
          <h1 className="text-2xl md:text-3xl font-bold text-black mb-3">
            {greeting}, {user.username}
          </h1>

          {/* Rank badge */}
          <div
            className={`inline-flex items-center gap-2 border-2 border-black px-3 py-1.5 font-bold text-xs uppercase ${getRankStyles(user.rank)}`}
          >
            <Crown className="h-3 w-3" />
            {user.rank}
          </div>
        </div>

        {/* CTA 
        <Link to="/sheets/create">
          <button className="bg-black text-white border-3 border-black px-8 py-4 font-bold text-base shadow-[4px_4px_0px_0px_#000] hover:shadow-[6px_6px_0px_0px_#000] hover:-translate-y-0.5 active:shadow-none active:translate-x-1 active:translate-y-1 transition-all">
            Create new sheet
          </button>
        </Link>*/}
      </div>

      {/* Accent stripe 
      <div className="h-2 bg-black" />*/}
    </div>
  );
}