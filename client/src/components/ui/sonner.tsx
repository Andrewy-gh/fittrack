import {
  CircleCheckIcon,
  InfoIcon,
  Loader2Icon,
  OctagonXIcon,
  TriangleAlertIcon,
} from "lucide-react";
import { useTheme } from "next-themes";
import type { CSSProperties } from "react";
import { Toaster as Sonner, type ToasterProps } from "sonner";

type ToasterStyle = CSSProperties & {
  [property: `--${string}`]: string | number;
};

const Toaster = ({ ...props }: ToasterProps) => {
  const { theme = "system" } = useTheme();
  const toasterTheme = theme === "light" || theme === "dark" ? theme : "system";
  const toasterStyle: ToasterStyle = {
    "--normal-bg": "var(--popover)",
    "--normal-text": "var(--popover-foreground)",
    "--normal-border": "var(--border)",
    "--success-bg": "var(--popover)",
    "--success-text": "var(--popover-foreground)",
    "--success-border": "var(--border)",
    "--error-bg": "var(--primary)",
    "--error-text": "var(--primary-foreground)",
    "--error-border": "var(--foreground)",
    "--warning-bg": "var(--accent)",
    "--warning-text": "var(--accent-foreground)",
    "--warning-border": "var(--accent)",
    "--info-bg": "var(--secondary)",
    "--info-text": "var(--secondary-foreground)",
    "--info-border": "var(--secondary)",
    "--border-radius": "var(--radius)",
  };

  return (
    <Sonner
      theme={toasterTheme}
      className="toaster group"
      richColors
      icons={{
        success: <CircleCheckIcon className="size-4" />,
        info: <InfoIcon className="size-4" />,
        warning: <TriangleAlertIcon className="size-4" />,
        error: <OctagonXIcon className="size-4" />,
        loading: <Loader2Icon className="size-4 animate-spin" />,
      }}
      style={toasterStyle}
      {...props}
    />
  );
};

export { Toaster };
