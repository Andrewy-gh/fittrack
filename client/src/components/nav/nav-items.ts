import { Activity, Bot, ChartColumn, Dumbbell } from "lucide-react";

export const navItems = [
  { to: "/workouts", label: "Workouts", icon: Dumbbell, search: undefined },
  { to: "/exercises", label: "Exercises", icon: Activity, search: undefined },
  {
    to: "/analytics",
    label: "Analytics",
    icon: ChartColumn,
    search: undefined,
  },
  {
    to: "/chat",
    label: "AI Chat",
    icon: Bot,
    search: { createChat: true },
  },
] as const;
