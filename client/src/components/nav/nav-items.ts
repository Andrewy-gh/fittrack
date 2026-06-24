import { Activity, Bot, ChartColumn, Dumbbell } from "lucide-react";

export const navItems = [
  { to: "/workouts", label: "Workouts", icon: Dumbbell },
  { to: "/exercises", label: "Exercises", icon: Activity },
  { to: "/analytics", label: "Analytics", icon: ChartColumn },
  { to: "/chat", label: "AI Chat", icon: Bot },
] as const;
